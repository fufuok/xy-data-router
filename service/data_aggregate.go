package service

import (
	"context"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/go-redis/redis/v8"
	"github.com/sheerun/queue"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var aggsQ = queue.New()

// 把内存队列数据推送到 Redis
func DataAggregate(ctx context.Context) {
	pushChan := make(chan bool)
	exitChan := make(chan struct{})
	pipe := common.RedisDB.Pipeline()
	keyExpire := time.Duration(conf.Config.SYSConf.SRCKeyExpire) * time.Second
	interval := conf.Config.SYSConf.DataAggsPushDuration
	ticker := time.NewTicker(interval)

	defer func() {
		_ = pipe.Close()
		ticker.Stop()
	}()

	// 弹出数据, 构建 Pipeline
	go func() {
		data := ""
		n := 0
		for {
			data, _ = aggsQ.Pop().(string)
			select {
			case <-ctx.Done():
				// 收到退出指令, 不处理数据, 写回队列
				aggsQ.Append(data)
				// 发出安全退出指令
				close(exitChan)
				return
			default:
				pos := strings.Index(data, conf.DataAggsQueueSep)
				key := data[:pos]
				value := data[pos+conf.DataAggsQueueSepLen:]
				pipe.RPush(common.CtxBG, key, value)
				pipe.Expire(common.CtxBG, key, keyExpire)
				// 数据达到批处理要求, 发出推送指令, 重置超时计时器
				n += 1
				if n > conf.Config.SYSConf.RedisPipelineLimit {
					pushChan <- true
					ticker.Reset(interval)
					n = 0
				}
			}
		}
	}()

	// 数据汇聚到 Redis
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pushChan:
				// 一定数量后推送到 Redis
				pushToRedis(pipe)
			case <-ticker.C:
				// 一定时间后推送到 Redis
				pushToRedis(pipe)
			}
		}
	}()

	<-ctx.Done()
	// 推送已汇聚的数据
	pushToRedis(pipe)
	// 收到 Pop 退出完成标识
	<-exitChan
	common.Log.Warn().Err(ctx.Err()).Msgf("DataAggregate worker exited: %s", ctx.Value("start"))
}

// 推送数据到 Redis
func pushToRedis(pipe redis.Pipeliner) {
	_, err := pipe.Exec(common.CtxBG)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("data aggregate rpush failed")
	}
}

// 接收请求, 将数据写入内存队列
func PushDataToQueue(apiname string, body []byte, ip string) {
	now := common.GetGlobalTime()
	// 按接口汇聚数据 key + sep + body
	aggsQ.Append(utils.AddString(
		// key: apiname:SRC:list:060102150405
		apiname, common.SrcTagKey, now.Format(common.RedisKeyTime),
		// sep
		conf.DataAggsQueueSep,
		// body: {apiname}=--={sysfield}=--={body}=-:-={body}
		apiname, conf.ESIndexSep,
		// sysfield: 增加系统字段
		`{"_ctime":"`, now.Format("2006-01-02T15:04:05Z"),
		`","_gtime":"`, now.Format(time.RFC3339),
		`","_cip":"`, ip, `"}`, conf.ESIndexSep,
		// body: 清除换行符等空白字符
		common.SpaceReplacer.Replace(utils.B2S(body)),
	))
}

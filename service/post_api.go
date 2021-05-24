package service

import (
	"context"
	"strings"
	"time"

	"github.com/fufuok/redislock"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xid"
	"github.com/imroc/req"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

func PostAPI(ctx context.Context, key string, apiConf conf.TPostAPIConf) {
	postInterval := time.Duration(apiConf.Interval) * time.Second
	newKey := strings.ReplaceAll(key, common.PostTagKey, common.PostTodoTagKey+xid.NewString())
	lockPostAPI := redislock.New(common.RedisDB, common.RedisLockKey+key).New(postInterval)
	timer := time.NewTimer(postInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			common.Log.Warn().Err(ctx.Err()).Msgf("PostAPI worker exited: %s", ctx.Value("start"))
			return
		case <-timer.C:
			// 取该接口名的全局锁
			if lockPostAPI.SafeLock() {
				doPostAPI(key, newKey, apiConf)
			}
		}

		// 下次执行时间校准
		timer.Reset(lockPostAPI.TTL())
	}
}

// 定时推送数据到第三方
func doPostAPI(key string, newKey string, apiConf conf.TPostAPIConf) {
	// 改键名: POST:list(存在时) -> GoPOST:list:xxx
	if ok, err := common.RenameScript.Run(common.CtxBG, common.RedisDB, []string{key},
		newKey, conf.Config.SYSConf.SRCKeyExpire).Bool(); !ok || err != nil {
		return
	}

	// 读数据键名列表
	keys, err := common.RedisDB.LRange(common.CtxBG, newKey, 0, -1).Result()
	if err != nil {
		common.LogSampled.Error().Err(err).Str("key", newKey).Msg("redis.lrange post keys")
		return
	}

	if len(keys) == 0 {
		return
	}

	var jsonList []string
	for _, todoKey := range keys {
		// 读数据
		data, err := common.RedisDB.LRange(common.CtxBG, todoKey, 0, -1).Result()
		if err != nil {
			common.LogSampled.Error().Err(err).Str("key", todoKey).Msg("redis.lrange todokey")
			continue
		}
		for _, srcStr := range data {
			// nagios=--={sysfield}=--={json}=-:-={json}
			s := strings.SplitN(srcStr, conf.ESIndexSep, 3)
			sysField := s[1]
			for _, js := range strings.Split(s[2], conf.ESBodySep) {
				// 丢弃无效的 JSON
				js, ok := common.IsValidJSON(js)
				if !ok {
					continue
				}
				// 附加系统字段
				if apiConf.WithSYSField {
					js = common.AppendSYSField(js, sysField)
				}
				jsonList = append(jsonList, js)
			}
		}
	}

	// Post JSON
	if len(jsonList) > 0 {
		_ = common.Pool.Submit(func() {
			BulkPostJSON(apiConf.API, jsonList)
		})
	}
}

// 暂无大量数据分批发送需求
func BulkPostJSON(api []string, jsonList []string) {
	body := utils.AddStringBytes("[", strings.Join(jsonList, ","), "]")
	for _, u := range api {
		if _, err := req.Post(u, req.BodyJSON(body), conf.ReqUserAgent); err != nil {
			common.LogSampled.Error().Err(err).Str("url", u).Msg("post json")
		}
	}
}

// 接口分类汇聚
func PostAggregate(key string) {
	// 获取接口名称
	apiname := strings.Split(key, common.RedisKeySep)[0]
	// 判断是否需要汇聚
	if conf.APIConfig[apiname].PostAPI.Interval > 0 {
		common.RedisDB.RPush(common.CtxBG, apiname+common.PostTagKey, key)
	}
}

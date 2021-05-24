package service

import (
	"context"
	"strings"
	"time"

	"github.com/fufuok/redislock"
	"github.com/go-redis/redis/v8"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	lockTODOLife = 1 * time.Second
	lockTODO     = redislock.New(common.RedisDB, common.RedisLockTodoKey).New(lockTODOLife)
	todoPipe     = common.RedisDB.Pipeline()
)

func TodoScheduler(ctx context.Context) {
	timer := time.NewTimer(lockTODOLife)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			common.Log.Warn().Err(ctx.Err()).Msgf("TodoScheduler worker exited: %s", ctx.Value("start"))
			return
		case <-timer.C:
			// 取锁
			if lockTODO.SafeLock() {
				src2todo()
			}
		}

		// 下次执行时间校准
		timer.Reset(lockTODO.TTL())
	}
}

func src2todo() {
	// 取 *:SRC:list:*
	srcList, err := common.RedisDB.Keys(common.CtxBG, "*"+common.SrcTagKey+"*").Result()
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("redis.keys *:src:list:*")
		return
	}

	lockReleseChan := make(chan struct{})
	defer close(lockReleseChan)

	_ = common.Pool.Submit(func() {
		// 保活锁
		ticker := time.NewTicker(lockTODOLife - 100*time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			select {
			case <-lockReleseChan:
				return
			default:
				lockTODO.Keepalive()
			}
		}
	})

	t := common.GetGlobalTime()
	now := t.Format(common.RedisKeyTime)
	now1 := t.Add(-1 * time.Second).Format(common.RedisKeyTime)
	n := 0
	for _, srcKey := range srcList {
		// nagios:SRC:list:0616164643 排除当前秒和上一秒的数据
		if strings.HasSuffix(srcKey, now) || strings.HasSuffix(srcKey, now1) {
			continue
		}
		// 改名并存入待办队列
		newKey := strings.ReplaceAll(srcKey, common.SrcTagKey, common.TodoTagKey)
		todoPipe.Rename(common.CtxBG, srcKey, newKey)
		todoPipe.RPush(common.CtxBG, common.TodoListKey, newKey)
		// 数据达到批处理要求, 发出推送指令
		n += 1
		if n > conf.Config.SYSConf.RedisPipelineLimit {
			todoToRedis(todoPipe)
			n = 0
		}
	}
	if n > 0 {
		todoToRedis(todoPipe)
	}
}

// 执行 Pipeline
func todoToRedis(pipe redis.Pipeliner) {
	if _, err := pipe.Exec(common.CtxBG); err != nil {
		common.LogSampled.Error().Err(err).Msg("todo scheduler rename and rpush failed")
	}
}

package service

import (
	"context"
	"time"

	"github.com/fufuok/xy-data-router/common"
)

func DataRouter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			common.Log.Warn().Err(ctx.Err()).Msgf("DataRouter worker exited: %s", ctx.Value("start"))
			return
		default:
			dispatcher()
		}
	}
}

// 数据分发
func dispatcher() {
	// 获取待办 key
	todoKey, err := common.RedisDB.BLPop(common.CtxBG, 30*time.Second, common.TodoListKey).Result()
	if err != nil {
		time.Sleep(50 * time.Millisecond)
		return
	}

	// 获取待办数据
	data, err := common.RedisDB.LRange(common.CtxBG, todoKey[1], 0, -1).Result()
	if err != nil {
		common.LogSampled.Error().Err(err).Strs("keys", todoKey).Msg("redis.lrange todokey")
		return
	}

	// 分类汇聚
	_ = common.Pool.Submit(func() {
		PostAggregate(todoKey[1])
	})

	// 写 ES
	_ = common.Pool.Submit(func() {
		PostES(todoKey[1], data)
	})
}

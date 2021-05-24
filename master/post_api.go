package master

import (
	"context"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 启动定时推送
func startPostAPI(ctx context.Context) {
	count := 0
	for _, apiConf := range conf.Config.APIConf {
		if apiConf.PostAPI.Interval > 0 {
			go service.PostAPI(ctx, apiConf.APIName+common.PostTagKey, apiConf.PostAPI)
			count += 1
		}
	}
}

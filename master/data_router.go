package master

import (
	"context"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 数据路由器(分发)
func startDataRouter(ctx context.Context) {
	for i := 0; i < conf.Config.SYSConf.DataRouterNum; i++ {
		go service.DataRouter(ctx)
	}
}

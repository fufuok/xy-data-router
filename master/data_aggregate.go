package master

import (
	"context"
	"runtime"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 数据汇聚(数据写入 Redis)
func startDataAggregate(ctx context.Context) {
	readerNum := conf.Config.SYSConf.DataAggsNum1CPU * runtime.NumCPU()
	for i := 0; i < readerNum; i++ {
		go service.DataAggregate(ctx)
	}
}

package master

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/arl/statsviz"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	// 重启信号
	restartChan = make(chan bool)
	// 配置重载信息
	reloadChan = make(chan bool)
)

func startDataRouterServe(ctx context.Context) {
	// 数据汇聚到 Redis
	startDataAggregate(ctx)

	// 待办数据调度器
	startTodoScheduler(ctx)

	// 数据路由器(分发)
	startDataRouter(ctx)

	// 数据推送到第三方
	startPostAPI(ctx)
}

func Start() {
	go func() {
		// 接口服务
		go startWebServer()
		// UDP 接口服务
		go startUDPServer()
		// 心跳日志
		go startHeartbeat()

		for {
			cancelCtx, cancel := context.WithCancel(context.Background())
			ctx := context.WithValue(cancelCtx, "start", time.Now())

			// 获取远程配置
			go startRemoteConf(ctx)

			// 数据处理服务
			go startDataRouterServe(ctx)

			select {
			case <-restartChan:
				// 强制退出, 由 Daemon 重启程序
				common.Log.Warn().Msg("restart <-restartChan")
				os.Exit(0)
			case <-reloadChan:
				// 重载配置及相关服务
				cancel()
				common.Log.Warn().Msg("reload <-reloadChan")
			}
		}
	}()
}

// 统计或性能工具
func StartPProf() {
	if conf.Config.SYSConf.PProfAddr != "" {
		go func() {
			_ = statsviz.RegisterDefault(statsviz.SendFrequency(time.Second * 5))
			_ = http.ListenAndServe(conf.Config.SYSConf.PProfAddr, nil)
		}()
	}
}

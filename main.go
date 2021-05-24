package main

import (
	"github.com/zh-five/xdaemon"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/master"
)

func main() {
	defer common.Clean()

	if !conf.Config.SYSConf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	master.StartPProf()
	master.Start()
	master.Watcher()

	select {}
}

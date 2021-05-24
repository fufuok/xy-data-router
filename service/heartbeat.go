package service

import (
	"fmt"
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/conf"
)

// 心跳日志
func Heartbeat() {
	for range time.Tick(time.Minute) {
		data := fmt.Sprintf(`{"type":"%s","version":"%s","internal_ipv4":"%s","external_ipv4":"%s"}`,
			conf.WebAPPName, conf.CurrentVersion, InternalIPv4, ExternalIPv4)
		PushDataToQueue(conf.Config.SYSConf.HeartbeatIndex, utils.S2B(data), ExternalIPv4)
	}
}

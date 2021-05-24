package master

import (
	"github.com/fufuok/xy-data-router/service"
)

// 心跳日志
func startHeartbeat() {
	go service.Heartbeat()
}

package master

import (
	"log"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 启动 UDP 接口服务
func startUDPServer() {
	exitUDPChan := make(chan error)

	switch conf.Config.SYSConf.UDPProto {
	case "gnet":
		go func() {
			if err := service.UDPServerG(conf.Config.SYSConf.UDPServerRWAddr, true); err != nil {
				exitUDPChan <- err
			}
		}()
		go func() {
			if err := service.UDPServerG(conf.Config.SYSConf.UDPServerRAddr, false); err != nil {
				exitUDPChan <- err
			}
		}()
	default:
		go func() {
			if err := service.UDPServer(conf.Config.SYSConf.UDPServerRWAddr, true); err != nil {
				exitUDPChan <- err
			}
		}()
		go func() {
			if err := service.UDPServer(conf.Config.SYSConf.UDPServerRAddr, false); err != nil {
				exitUDPChan <- err
			}
		}()
	}

	common.Log.Info().
		Str("raddr", conf.Config.SYSConf.UDPServerRAddr).Str("rwaddr", conf.Config.SYSConf.UDPServerRWAddr).
		Msgf("Listening and serving UDP")

	err := <-exitUDPChan
	log.Fatalln("Failed to start UDP Server:", err, "\nbye.")
}

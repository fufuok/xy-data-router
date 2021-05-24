package service

import (
	"net"
	"runtime"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// UDP 接口并发读取数据协程数
var udpGoLimit = utils.MinInt(conf.Config.SYSConf.UDPGoReadNum1CPU*runtime.NumCPU(), conf.UDPGoReadNumMax)

// UDP 返回值
var outBytes = []byte("1")

// 标准的 UDP 服务
func UDPServer(addr string, withSendTo bool) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	// 收发缓冲区
	// _ = conn.SetReadBuffer(1024 * 1024 * 20)
	// _ = conn.SetWriteBuffer(1024 * 1024 * 20)

	for i := 0; i < udpGoLimit; i++ {
		go udpReader(conn, withSendTo)
	}

	select {}
}

// UDP 数据读取
func udpReader(conn *net.UDPConn, withSendTo bool) {
	readerBuf := make([]byte, conf.UDPMaxRW)
	for {
		n, clientAddr, err := conn.ReadFromUDP(readerBuf)
		if err == nil && n > 0 {
			if withSendTo || n < 7 {
				_ = common.Pool.Submit(func() {
					writeToUDP(conn, clientAddr)
				})
			}
			if n >= 7 {
				body := utils.CopyBytes(readerBuf[:n])
				clientIP := clientAddr.IP.String()
				_ = common.Pool.Submit(func() {
					saveUDPData(body, clientIP)
				})
			}
		}
	}
}

// UDP 应答
func writeToUDP(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	_, _ = conn.WriteToUDP(outBytes, clientAddr)
}

// 校验并保存数据
func saveUDPData(body []byte, clientIP string) bool {

	// 接口名称与索引名称相同, 存放在 _x 字段
	esIndex := common.GetUDPESIndex(body, conf.UDPESIndexField)
	if esIndex == "" {
		return false
	}

	// 必有字段校验, 有对应配置项时检查
	if apiConf, ok := conf.APIConfig[esIndex]; ok && !common.CheckRequiredField(body, apiConf.RequiredField) {
		return false
	}

	// 保存数据
	PushDataToQueue(esIndex, body, clientIP)

	return true
}

package service

import (
	"runtime"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/myip"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	// 系统启动时间
	start = time.Now()
	// 服务器 IP
	InternalIPv4 string
	ExternalIPv4 string
)

func init() {
	go func() {
		InternalIPv4 = myip.InternalIPv4()
	}()
	go func() {
		ExternalIPv4 = myip.ExternalIPv4()
	}()
}

// 运行状态
func RunningStatus() common.TStringAnyMaps {
	poolStats := common.RedisDB.PoolStats()
	redisOptions := common.RedisDB.Options()
	redisStatus := common.TStringAnyMaps{
		"Hits":          poolStats.Hits,
		"Misses":        poolStats.Misses,
		"Timeouts":      poolStats.Timeouts,
		"TotalConns":    poolStats.TotalConns,
		"IdleConns":     poolStats.IdleConns,
		"StaleConns":    poolStats.StaleConns,
		"PoolSize":      redisOptions.PoolSize,
		"Addr":          redisOptions.Addr,
		"DB":            redisOptions.DB,
		"DBSize":        getRedisDBSize(),
		"TotalTodoList": getTotalTodoList(),
		"TotalSRCList":  getTotalSRCList(),
	}

	sysStatus := common.TStringAnyMaps{
		"APPName":         conf.WebAPPName,
		"Version":         conf.CurrentVersion,
		"Update":          conf.LastChange,
		"Uptime":          time.Since(start).String(),
		"ConfigVer":       conf.Config.SYSConf.MainConfig.ConfigVer,
		"ConfigMD5":       conf.Config.SYSConf.MainConfig.ConfigMD5,
		"GoVersion":       runtime.Version(),
		"NumCpus":         runtime.NumCPU(),
		"NumGoroutine":    runtime.NumGoroutine(),
		"OS":              runtime.GOOS,
		"NumCgoCall":      runtime.NumCgoCall(),
		"InternalIPv4":    InternalIPv4,
		"ExternalIPv4":    ExternalIPv4,
		"AggsQueueLength": aggsQ.Length(),
	}

	return common.TStringAnyMaps{
		"SYS":   sysStatus,
		"MEM":   memStats(),
		"REDIS": redisStatus,
	}
}

// 内存信息
func memStats() common.TStringAnyMaps {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return common.TStringAnyMaps{
		// 程序启动后累计申请的字节数
		"TotalAlloc":  ms.TotalAlloc,
		"TotalAlloc_": utils.HumanBytes(ms.TotalAlloc),
		// 虚拟占用, 总共向系统申请的字节数
		"HeapSys":  ms.HeapSys,
		"HeapSys_": utils.HumanBytes(ms.HeapSys),
		// 使用中或未使用, 但未被 GC 释放的对象的字节数
		"HeapAlloc":  ms.HeapAlloc,
		"HeapAlloc_": utils.HumanBytes(ms.HeapAlloc),
		// 使用中的对象的字节数
		"HeapInuse":  ms.HeapInuse,
		"HeapInuse_": utils.HumanBytes(ms.HeapInuse),
		// 已释放的内存, 还没被堆再次申请的内存
		"HeapReleased":  ms.HeapReleased,
		"HeapReleased_": utils.HumanBytes(ms.HeapReleased),
		// 没被使用的内存, 包含了 HeapReleased, 可被再次申请和使用
		"HeapIdle":  ms.HeapIdle,
		"HeapIdle_": utils.HumanBytes(ms.HeapIdle),
		// 下次 GC 的阈值, 当 HeapAlloc 达到该值触发 GC
		"NextGC":  ms.NextGC,
		"NextGC_": utils.HumanBytes(ms.NextGC),
		// 上次 GC 时间
		"LastGC": time.Unix(0, int64(ms.LastGC)).Format(time.RFC3339Nano),
		// GC 次数
		"NumGC": utils.Commau(ms.NextGC),
		// 被强制 GC 的次数
		"NumForcedGC": ms.NumForcedGC,
	}
}

// Redis Info
func RedisInfo() common.TStringAnyMaps {
	ret := common.TStringAnyMaps{}
	info := common.RedisDB.Info(common.CtxBG).Val()

	for _, v := range strings.Split(info, "\n") {
		v = strings.TrimSpace(v)
		if v == "" || strings.HasPrefix(v, "#") {
			continue
		}
		items := strings.SplitN(v, ":", 2)
		ret[items[0]] = items[1]
	}

	return ret
}

// 当前数据库键数量
func getRedisDBSize() int {
	num, err := common.RedisDB.DBSize(common.CtxBG).Result()
	if err != nil {
		return -1
	}

	return int(num)
}

func getTotalTodoList() int {
	num, err := common.RedisDB.LLen(common.CtxBG, common.TodoListKey).Result()
	if err != nil {
		return -1
	}

	return int(num)
}

func getTotalSRCList() int {
	srcList, err := common.RedisDB.Keys(common.CtxBG, "*"+common.SrcTagKey+"*").Result()
	if err != nil {
		return -1
	}

	return len(srcList)
}

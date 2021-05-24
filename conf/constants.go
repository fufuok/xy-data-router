package conf

import (
	"time"
)

const (
	WebAPPName     = "XY.DataRouter"
	CurrentVersion = "1.0.12.21051313"
	LastChange     = "Milestone Version"
	ProjectName    = "xydatarouter"

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel = 3
	// 抽样日志设置 (每秒最多 3 个日志)
	LogSamplePeriodDur = time.Second
	LogSampleBurst     = 3
	// 每 100M 自动切割, 保留 30 天内最近 10 个日志文件
	LogFileMaxSize    = 100
	LogFileMaxBackups = 10
	LogFileMaxAge     = 30

	// HTTP 接口端口
	WebServerAddr = ":6600"
	// ES 慢查询日志时间设置, 默认: > 10秒则记录
	ESSlowQueryDuration = 10 * time.Second
	// Web 慢响应日志时间设置, 默认: > 1秒则记录
	WebSlowRespDuration = time.Second
	// HTTP 响应码日志记录, 默认: 500, 即大于等于 500 的状态码记录日志
	WebErrorCodeLog = 500
	// POST 最大 500M, Request Entity Too Large
	BodyLimit = 500 << 20

	// UDP 接口端口, 不应答(Echo包除外)
	UDPServerRAddr = ":6611"
	// UDP 接口端口, 每个包都应答字符: 1
	UDPServerRWAddr = ":6622"
	// 1. 在链路层, 由以太网的物理特性决定了数据帧的长度为 (46+18) - (1500+18)
	//    其中的 18 是数据帧的头和尾, 也就是说数据帧的内容最大为 1500 (不包括帧头和帧尾)
	//    即 MTU (Maximum Transmission Unit) 为 1500
	// 2. 在网络层, 因为 IP 包的首部要占用 20 字节, 所以这的 MTU 为 1500 - 20 = 1480
	// 3. 在传输层, 对于 UDP 包的首部要占用 8 字节, 所以这的 MTU 为 1480 - 8 = 1472
	// 4. UDP 协议中有 16 位的 UDP 报文长度, 即 UDP 报文长度不能超过 65536, 则数据最大为 65507
	UDPMaxRW = 65507
	// UDP Goroutine 并发读取的数量 / CPU
	UDPGoReadNum1CPU = 50
	UDPGoReadNumMax  = 1000

	// ES 数据分隔符
	ESBodySep  = "=-:-="
	ESIndexSep = "=--="

	// ES 单次批量写入最大条数或最大字节数
	ESPostBatchNum   = 3000
	ESPOSTBatchBytes = 30 << 20

	// Redis Auth 短语环境变量 Key
	RedisAuthKeyName = "DR_REDIS_AUTH"
	// Redis 单个 CPU 的连接池数, 默认为 10 connection pool timeout
	RedisPoolSize1CPU = 50
	// Redis 连接池最大值
	RedisPoolSizeMax = 1500
	// 原始数据 Key 过期时间, 默认 30 秒, OOM command not allowed when used memory > 'maxmemory'
	SRCKeyExpire = 30

	// 数据分发协程数(默认是 10 * runtime.NumCPU(), 以配置文件优先)
	DataRouterNum1CPU = 10

	// 数据汇聚分隔符 redis_key=-|-=redis_data
	DataAggsQueueSep    = "=-|-="
	DataAggsQueueSepLen = 5
	// 数据汇聚队列处理协程 / CPU
	DataAggsNum1CPU = 2
	// (毫秒)数据汇聚推送到 Redis 的时间间隔或数量间隔
	DataAggsPushMs     = 200
	RedisPipelineLimit = 8000

	// 项目基础密钥 (环境变量名)
	BaseSecretKeyName = "DR_BASE_SECRET_KEY"
	// 用于解密基础密钥值的密钥 (编译在程序中)
	BaseSecretSalt = "Fufu@dr.777"

	// 文件变化监控时间间隔(分)
	WatcherInterval = 1

	// 心跳日志索引
	HeartbeatIndex = "monitor_heartbeat_report"
	ReqAPIName     = "REQ_API_NAME"
	ReqAPIBody     = "REQ_API_BODY"
)

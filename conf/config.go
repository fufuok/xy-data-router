package conf

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/json"
)

// 接口配置
type TJSONConf struct {
	SYSConf     TSYSConf   `json:"sys_conf"`
	APIConf     []TAPIConf `json:"api_conf"`
	ESWhiteList []string   `json:"es_white_list"`
}

// 主配置, 变量意义见配置文件中的描述及 constants.go 中的默认值
type TSYSConf struct {
	Debug                bool       `json:"debug"`
	PProfAddr            string     `json:"pprof_addr"`
	Log                  tLogConf   `json:"log"`
	WebServerAddr        string     `json:"web_server_addr"`
	BodyLimit            int        `json:"body_limit"`
	ESSlowQuery          int        `json:"es_slow_query"`
	WebSlowResponse      int        `json:"web_slow_response"`
	WebErrCodeLog        int        `json:"web_errcode_log"`
	UDPServerRAddr       string     `json:"udp_server_raddr"`
	UDPServerRWAddr      string     `json:"udp_server_rwaddr"`
	ESAddress            []string   `json:"es_address"`
	RedisAddress         string     `json:"redis_address"`
	RedisAuthKeyName     string     `json:"redis_auth_key_name"`
	RedisPoolSize1CPU    int        `json:"redis_poolsize_1cpu"`
	DataRouterNum1CPU    int        `json:"datarouter_num_1cpu"`
	DataAggsNum1CPU      int        `json:"data_aggs_num_1cpu"`
	DataAggsPushMs       int        `json:"data_aggs_push_ms"`
	RedisPipelineLimit   int        `json:"redis_pipeline_limit"`
	UDPGoReadNum1CPU     int        `json:"udp_go_read_num_1cpu"`
	UDPProto             string     `json:"udp_proto"`
	MainConfig           TFilesConf `json:"main_config"`
	RestartMain          bool       `json:"restart_main"`
	WatcherInterval      int        `json:"watcher_interval"`
	HeartbeatIndex       string     `json:"heartbeat_index"`
	BaseSecretValue      string
	RedisAuthValue       string
	DataRouterNum        int
	SRCKeyExpire         int
	WebSlowRespDuration  time.Duration
	ESSlowQueryDuration  time.Duration
	DataAggsPushDuration time.Duration
	LogLimitDuration     time.Duration
}

type tLogConf struct {
	Level      int    `json:"level"`
	NoColor    bool   `json:"no_color"`
	File       string `json:"file"`
	Period     int    `json:"period"`
	Burst      uint32 `json:"burst"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	PeriodDur  time.Duration
}

type TAPIConf struct {
	APIName       string       `json:"api_name"`
	ESIndex       string       `json:"es_index"`
	ESIndexSplit  string       `json:"es_index_split"`
	RequiredField []string     `json:"required_field"`
	PostAPI       TPostAPIConf `json:"post_api"`
}

type TPostAPIConf struct {
	API          []string `json:"api"`
	Interval     int      `json:"interval"`
	WithSYSField bool     `json:"with_sys_field"`
}

type TFilesConf struct {
	Path            string `json:"path"`
	Method          string `json:"method"`
	SecretName      string `json:"secret_name"`
	API             string `json:"api"`
	Interval        int    `json:"interval"`
	SecretValue     string
	GetConfDuration time.Duration
	ConfigMD5       string
	ConfigVer       time.Time
}

func init() {
	confFile := flag.String("c", ConfigFile, "配置文件绝对路径")
	flag.Parse()
	ConfigFile = *confFile
	if err := LoadConf(); err != nil {
		log.Fatalln("Failed to initialize config:", err, "\nbye.")
	}
}

// 加载配置
func LoadConf() error {
	config, apiConfig, whiteList, err := readConf()
	if err != nil {
		return err
	}

	Config = *config
	APIConfig = apiConfig
	ESWhiteListConfig = whiteList

	return nil
}

// 读取配置
func readConf() (*TJSONConf, map[string]TAPIConf, map[*net.IPNet]struct{}, error) {
	body, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, nil, nil, err
	}

	var config *TJSONConf
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, nil, nil, err
	}

	// 基础密钥 Key
	config.SYSConf.BaseSecretValue = utils.GetenvDecrypt(BaseSecretKeyName, BaseSecretSalt)
	if config.SYSConf.BaseSecretValue == "" {
		return nil, nil, nil, fmt.Errorf("%s cannot be empty", BaseSecretKeyName)
	}

	// RedisAuth
	if config.SYSConf.RedisAuthKeyName == "" {
		config.SYSConf.RedisAuthKeyName = RedisAuthKeyName
	}
	config.SYSConf.RedisAuthValue = utils.GetenvDecrypt(config.SYSConf.RedisAuthKeyName, config.SYSConf.BaseSecretValue)

	// 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	if config.SYSConf.Log.Level > 7 || config.SYSConf.Log.Level < -1 {
		config.SYSConf.Log.Level = LogLevel
	}

	// 抽样日志设置 (x 秒 n 条)
	if config.SYSConf.Log.Burst < 0 || config.SYSConf.Log.Period < 0 {
		config.SYSConf.Log.PeriodDur = LogSamplePeriodDur
		config.SYSConf.Log.Burst = LogSampleBurst
	} else {
		config.SYSConf.Log.PeriodDur = time.Duration(config.SYSConf.Log.Period) * time.Second
	}

	// 日志文件
	if config.SYSConf.Log.File == "" {
		config.SYSConf.Log.File = LogFile
	}

	// 日志大小和保存设置
	if config.SYSConf.Log.MaxSize < 1 {
		config.SYSConf.Log.MaxSize = LogFileMaxSize
	}
	if config.SYSConf.Log.MaxBackups < 1 {
		config.SYSConf.Log.MaxBackups = LogFileMaxBackups
	}
	if config.SYSConf.Log.MaxAge < 1 {
		config.SYSConf.Log.MaxAge = LogFileMaxAge
	}

	// 单个 CPU 的 UDP 并发读取协程数, 默认为 50
	if config.SYSConf.UDPGoReadNum1CPU < 10 {
		config.SYSConf.UDPGoReadNum1CPU = UDPGoReadNum1CPU
	}

	// 数据汇聚推送到 Redis 的数量间隔
	if config.SYSConf.RedisPipelineLimit < 1 {
		config.SYSConf.RedisPipelineLimit = RedisPipelineLimit
	}

	// (毫秒)数据汇聚推送到 Redis 的时间间隔
	if config.SYSConf.DataAggsPushMs < 1 {
		config.SYSConf.DataAggsPushMs = DataAggsPushMs
	}
	config.SYSConf.DataAggsPushDuration = time.Duration(config.SYSConf.DataAggsPushMs) * time.Millisecond

	// 单个 CPU 的队列处理协程数
	if config.SYSConf.DataAggsNum1CPU < 1 {
		config.SYSConf.DataAggsNum1CPU = DataAggsNum1CPU
	}

	// 单个 CPU 的 Redis 连接池数, 默认为 4 connection pool timeout
	if config.SYSConf.RedisPoolSize1CPU < 10 {
		config.SYSConf.RedisPoolSize1CPU = RedisPoolSize1CPU
	}

	// 路由分发每 CPU 并发数
	if config.SYSConf.DataRouterNum1CPU < 1 {
		config.SYSConf.DataRouterNum1CPU = DataRouterNum1CPU
	}

	// 数据分发协程数
	config.SYSConf.DataRouterNum = config.SYSConf.DataRouterNum1CPU * runtime.NumCPU()

	// 优先使用配置中的绑定参数(HTTP)
	if config.SYSConf.WebServerAddr == "" {
		config.SYSConf.WebServerAddr = WebServerAddr
	}

	// 优先使用配置中的绑定参数(UDP带应答)
	if config.SYSConf.UDPServerRWAddr == "" {
		config.SYSConf.UDPServerRWAddr = UDPServerRWAddr
	}

	// 优先使用配置中的绑定参数(UDP不带应答)
	if config.SYSConf.UDPServerRAddr == "" {
		config.SYSConf.UDPServerRAddr = UDPServerRAddr
	}

	// ES 慢查询日志时间设置
	if config.SYSConf.ESSlowQuery < 1 {
		config.SYSConf.ESSlowQueryDuration = ESSlowQueryDuration
	} else {
		config.SYSConf.ESSlowQueryDuration = time.Duration(config.SYSConf.ESSlowQuery) * time.Second
	}

	// Web 慢响应日志时间设置
	if config.SYSConf.WebSlowResponse < 1 {
		config.SYSConf.WebSlowRespDuration = WebSlowRespDuration
	} else {
		config.SYSConf.WebSlowRespDuration = time.Duration(config.SYSConf.WebSlowResponse) * time.Second
	}

	// HTTP 响应码日志设置, 默认 >= 500
	if config.SYSConf.WebErrCodeLog < 1 {
		config.SYSConf.WebErrCodeLog = WebErrorCodeLog
	}

	// 原始数据 Key 过期时间 > PostAPI.Interval + 5
	config.SYSConf.SRCKeyExpire = SRCKeyExpire
	apiConfig := make(map[string]TAPIConf)
	for _, apiConf := range config.APIConf {
		apiConfig[apiConf.APIName] = apiConf
		if apiConf.PostAPI.Interval+5 > config.SYSConf.SRCKeyExpire {
			// 原始数据过期时间 = 最长的 PostAPI 间隔时间 + 5 秒
			config.SYSConf.SRCKeyExpire = apiConf.PostAPI.Interval + 5
		}
	}

	// ES IP 白名单
	whiteList := make(map[*net.IPNet]struct{})
	for _, ip := range config.ESWhiteList {
		// 排除空白行, __ 或 # 开头的注释行
		ip = strings.TrimSpace(ip)
		if ip == "" || strings.HasPrefix(ip, "__") || strings.HasPrefix(ip, "#") {
			continue
		}
		// 补全掩码
		if !strings.Contains(ip, "/") {
			if strings.Contains(ip, ":") {
				ip = ip + "/128"
			} else {
				ip = ip + "/32"
			}
		}
		// 转为网段
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return nil, nil, nil, err
		}
		whiteList[ipNet] = struct{}{}
	}

	// 每次获取远程主配置的时间间隔, < 30 秒则禁用该功能
	if config.SYSConf.MainConfig.Interval > 29 {
		// 远程获取主配置 API, 解密 SecretName
		if config.SYSConf.MainConfig.SecretName != "" {
			config.SYSConf.MainConfig.SecretValue = utils.GetenvDecrypt(config.SYSConf.MainConfig.SecretName,
				config.SYSConf.BaseSecretValue)
			if config.SYSConf.MainConfig.SecretValue == "" {
				return nil, nil, nil, fmt.Errorf("%s cannot be empty", config.SYSConf.MainConfig.SecretName)
			}
		}
		config.SYSConf.MainConfig.GetConfDuration = time.Duration(config.SYSConf.MainConfig.Interval) * time.Second
		config.SYSConf.MainConfig.Path = ConfigFile
	}

	// 文件变化监控时间间隔
	if config.SYSConf.WatcherInterval < 1 {
		config.SYSConf.WatcherInterval = WatcherInterval
	}

	// 心跳日志索引
	if config.SYSConf.HeartbeatIndex == "" {
		config.SYSConf.HeartbeatIndex = HeartbeatIndex
	}

	// HTTP 请求体限制, -1 表示无限
	if config.SYSConf.BodyLimit == 0 {
		config.SYSConf.BodyLimit = BodyLimit
	}

	return config, apiConfig, whiteList, nil
}

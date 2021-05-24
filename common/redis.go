package common

import (
	"context"
	"errors"
	"log"
	"runtime"
	"time"

	"github.com/fufuok/utils"
	"github.com/go-redis/redis/v8"

	"github.com/fufuok/xy-data-router/conf"
)

var CtxBG = context.Background()
var RedisDB *redis.Client
var GTimeSub time.Duration

const (
	// Redis 键名值分隔符, 默认为冒号
	RedisKeySep = ":"

	// Redis 键名时间格式
	RedisKeyTime = "060102150405"

	// Redis 全局锁键名前缀
	RedisLockKey = "DR:REDISLOCK:"

	// 待办调度锁
	RedisLockTodoKey = "DR:TODOLOCK"

	TodoListKey    = "DR:TODO:LIST"
	SrcTagKey      = ":SRC:list:"
	TodoTagKey     = ":TODO:list:"
	PostTagKey     = ":POST:list"
	PostTodoTagKey = ":GoPOST:list:"
)

var RenameScript = redis.NewScript(`
			if redis.call("exists", KEYS[1]) == 1 then
				redis.call("rename", KEYS[1], ARGV[1])
				redis.call("expire", ARGV[1], ARGV[2])
				return 1
			else
				return 0
			end
	`)

func init() {
	if err := InitRedisDB(); err != nil {
		log.Fatalln("Failed to connect Redis:", err, "\nbye.")
	}

	// 校准程序全局时间
	go func() {
		for {
			GTimeSub = GetRedisTime().Sub(time.Now())
			time.Sleep(10 * time.Second)
		}
	}()
}

func InitRedisDB() error {
	rdb, err := newRedisDB()
	if err != nil {
		return err
	}

	RedisDB = rdb

	return nil
}

func newRedisDB() (*redis.Client, error) {
	opt, err := redis.ParseURL(conf.Config.SYSConf.RedisAddress)
	if err != nil {
		return nil, err
	}

	// 每 CPU 连接池数 + 数据分发协程数
	poolSize := conf.Config.SYSConf.RedisPoolSize1CPU*runtime.NumCPU() + conf.Config.SYSConf.DataRouterNum
	opt.PoolSize = utils.MinInt(poolSize, conf.RedisPoolSizeMax)

	// 环境变量中的 Redis 密码
	if conf.Config.SYSConf.RedisAuthValue != "" {
		opt.Password = conf.Config.SYSConf.RedisAuthValue
	}

	rdb := redis.NewClient(opt)

	if rdb.Ping(CtxBG).Val() != "PONG" {
		return nil, errors.New("redis connection failed: " + conf.Config.SYSConf.RedisAddress)
	}

	return rdb, nil
}

// 统一时间: Redis 时间
func GetGlobalTime() time.Time {
	return time.Now().Add(GTimeSub)
}

// 统一时间: Redis 时间并格式化
func GetGlobalDataTime(layout string) string {
	if layout == "" {
		layout = RedisKeyTime
	}
	return GetGlobalTime().Format(layout)
}

// 获取 Redis 时间作为原子钟
func GetRedisTime() time.Time {
	return RedisDB.Time(CtxBG).Val()
}

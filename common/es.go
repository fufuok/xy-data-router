package common

import (
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/conf"
)

var ES *elasticsearch.Client

func init() {
	if err := InitES(); err != nil {
		log.Fatalln("Failed to connect ES:", err, "\nbye.")
	}
}

func InitES() error {
	es, err := newES()
	if err != nil {
		return err
	}

	ES = es

	return nil
}

func newES() (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		EnableDebugLogger:    true,
		Addresses:            conf.Config.SYSConf.ESAddress,
		RetryOnStatus:        []int{502, 503, 504, 429},
		EnableRetryOnTimeout: true,
		MaxRetries:           3,
	})
	if err != nil {
		return nil, err
	}

	if _, err := es.Ping(); err != nil {
		return nil, err
	}

	return es, nil
}

// 获取 ES 索引名称, 不能包含冒号(Redis键分隔符)
func GetUDPESIndex(body []byte, key string) string {
	index := gjson.GetBytes(body, key).String()
	if index != "" {
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(index)), RedisKeySep, "")
	}

	return ""
}

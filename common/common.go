package common

import (
	"strings"

	"github.com/panjf2000/gnet/pool/goroutine"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type TAnyMaps map[interface{}]interface{}
type TStringAnyMaps map[string]interface{}

var (
	IPv4Zero = "0.0.0.0"
	// ES 系统字段
	ESSYSField = [...]string{"_cip", "_ctime", "_gtime"}
	// 空白字符替换器
	SpaceReplacer = strings.NewReplacer(
		"\n", "",
		"\r", "",
		"\f", "",
		"\v", "",
		"\t", "",
		"\u0085", "",
		"\u00a0", "",
	)
	// 协程池
	Pool = goroutine.Default()
)

// 程序退出时清理
func Clean() {
	Pool.Release()
}

// 检查必有字段
func CheckRequiredField(body []byte, fields []string) bool {
	for _, field := range fields {
		if gjson.GetBytes(body, field).String() == "" {
			return false
		}
	}

	return true
}

// 检查是否为正确的 JSON (字典)
func IsValidJSON(data string) (string, bool) {
	data = strings.TrimSpace(data)
	return data, strings.HasPrefix(data, "{") && strings.Contains(data, `"`) && gjson.Valid(data)
}

// 附加系统字段
func AppendSYSField(data, sysField string) string {
	for _, key := range ESSYSField {
		data, _ = sjson.Set(data, key, gjson.Get(sysField, key).String())
	}

	return data
}

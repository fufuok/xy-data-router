package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/sjson"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/service"
)

// 兼容旧接口
func oldAPIHandler(delKeys []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(c.Body()) == 0 {
			return c.SendString("0")
		}

		body := utils.CopyBytes(c.Body())

		// 删除可能非法中文编码的字段
		for _, k := range delKeys {
			body, _ = sjson.DeleteBytes(body, k)
		}

		// 接口名
		apiname := utils.CopyString(c.Path())
		apiname = strings.Trim(strings.Replace(apiname, "/bulk", "", 1), "/")

		// 请求 IP
		ip := utils.GetSafeString(c.IP(), common.IPv4Zero)

		// 写入队列
		_ = common.Pool.Submit(func() {
			service.PushDataToQueue(apiname, body, ip)
		})

		// 旧接口返回值处理
		return c.SendString("1")
	}
}

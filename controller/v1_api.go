package controller

import (
	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

// 处理接口请求
func V1APIHandler(c *fiber.Ctx) error {
	// 按接口汇聚数据
	apiname, _ := c.Locals(conf.ReqAPIName).(string)
	body, _ := c.Locals(conf.ReqAPIBody).([]byte)
	ip := utils.GetSafeString(c.IP(), common.IPv4Zero)
	_ = common.Pool.Submit(func() {
		service.PushDataToQueue(apiname, body, ip)
	})

	return middleware.APISuccess(c, nil, 0)
}

package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
)

// 通用异常处理
func APIException(c *fiber.Ctx, code int, msg string) error {
	if msg == "" {
		msg = "错误的请求"
	}
	return c.Status(code).JSON(common.APIFailureData(msg))
}

// 返回失败, 状态码: 200
func APIFailure(c *fiber.Ctx, msg string) error {
	return APIException(c, fiber.StatusOK, msg)
}

// 返回成功, 状态码: 200
func APISuccess(c *fiber.Ctx, data interface{}, count int) error {
	return c.JSON(common.APISuccessData(data, count))
}

// 返回文本消息
func TxtMsg(c *fiber.Ctx, msg string) error {
	return c.SendString(msg)
}

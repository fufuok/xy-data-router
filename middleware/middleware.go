package middleware

import (
	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/sjson"
)

// 获取所有 GET 请求参数
func query2JSON(c *fiber.Ctx) (body []byte) {
	c.Request().URI().QueryArgs().VisitAll(func(key []byte, val []byte) {
		body, _ = sjson.SetBytes(body, utils.B2S(key), val)
	})

	return
}

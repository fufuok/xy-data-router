package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service"
)

func runningStatusHandler(c *fiber.Ctx) error {
	return c.JSON(service.RunningStatus())
}

func redisInfoHandler(c *fiber.Ctx) error {
	return c.JSON(service.RedisInfo())
}

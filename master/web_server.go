package master

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/controller"
	"github.com/fufuok/xy-data-router/middleware"
)

// 接口服务
func startWebServer() {
	app := fiber.New(fiber.Config{
		ServerHeader:          conf.WebAPPName,
		BodyLimit:             conf.Config.SYSConf.BodyLimit,
		DisableStartupMessage: true,
		// Immutable:             true,
	})
	app.Use(compress.New(), middleware.RecoverLogger())
	app = controller.SetupRouter(app)

	common.Log.Info().Str("addr", conf.Config.SYSConf.WebServerAddr).Msg("Listening and serving HTTP")
	if err := app.Listen(conf.Config.SYSConf.WebServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}

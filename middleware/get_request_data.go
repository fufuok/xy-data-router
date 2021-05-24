package middleware

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"

	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// 获取请求数据
func GetRequestData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Some values returned from *fiber.Ctx are not immutable by default
		apiname := utils.CopyString(c.Params("apiname"))

		// 检查接口配置
		apiConf, ok := conf.APIConfig[apiname]
		if !ok || apiConf.APIName == "" {
			return APIFailure(c, "接口配置有误")
		}

		// 按场景获取数据
		var body []byte
		chkField := true
		if c.Method() == "GET" {
			// GET 单条数据
			body = query2JSON(c)
		} else {
			body = utils.CopyBytes(c.Body())
			uri := strings.TrimRight(c.Path(), "/")

			if strings.HasSuffix(uri, "/gzip") {
				// 请求体解压缩
				uri = uri[:len(uri)-5]
				unRaw, err := gzip.NewReader(bytes.NewReader(body))
				if err != nil {
					return APIFailure(c, "数据解压失败")
				}
				body, err = ioutil.ReadAll(unRaw)
				if err != nil {
					return APIFailure(c, "数据读取失败")
				}
			}

			if strings.HasSuffix(uri, "/bulk") {
				// 批量数据不检查必有字段
				chkField = false
			}
		}

		if chkField {
			// 检查必有字段
			if ok := common.CheckRequiredField(body, apiConf.RequiredField); !ok {
				return APIFailure(c, utils.AddString("必填字段: ", strings.Join(apiConf.RequiredField, ",")))
			}
		}

		c.Locals(conf.ReqAPIName, apiname)
		c.Locals(conf.ReqAPIBody, body)

		return c.Next()
	}
}

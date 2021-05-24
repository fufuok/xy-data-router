package service

import (
	"bytes"
	"strings"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/json"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

type tESBulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

func PostES(key string, data []string) {
	// 索引名称, 配置为空时使用接口名称, 按天索引
	s := strings.Split(key, common.RedisKeySep)
	apiname := s[0]
	esIndex := conf.APIConfig[apiname].ESIndex
	if esIndex == "" {
		esIndex = apiname
	}

	// naios:todx:list:200615125959
	ymd := ""
	if len(s) == 4 {
		ymd = s[3][:6]
	} else {
		ymd = common.GetGlobalDataTime("060102")
	}

	// 索引切割
	switch conf.APIConfig[apiname].ESIndexSplit {
	case "year":
		esIndex = esIndex + "_" + ymd[:2]
	case "month":
		esIndex = esIndex + "_" + ymd[:4]
	case "none":
		break
	default:
		esIndex = esIndex + "_" + ymd
	}

	index := utils.AddStringBytes(`{"index":{"_index":"`, esIndex, `","_type":"_doc"}}`, "\n")

	var bodyBuf bytes.Buffer
	n := 0
	for _, srcStr := range data {
		// nagios=--={sysfield}=--={json}=-:-={json}
		s := strings.SplitN(srcStr, conf.ESIndexSep, 3)
		for _, js := range strings.Split(s[2], conf.ESBodySep) {
			// 确保数据格式正确
			js, ok := common.IsValidJSON(js)
			if !ok {
				if js != "" {
					common.LogSampled.Warn().Str("body", js).Str("es_index", esIndex).Msg("Invalid JSON")
				}
				continue
			}

			// 每个文档附加系统字段
			js = common.AppendSYSField(js, s[1])

			bodyBuf.Write(index)
			bodyBuf.WriteString(js)
			bodyBuf.WriteByte('\n')
			n += 1
			if n%conf.ESPostBatchNum == 0 || bodyBuf.Len() > conf.ESPOSTBatchBytes {
				_ = common.Pool.Submit(func() {
					esBulk(utils.CopyBytes(bodyBuf.Bytes()))
				})
				bodyBuf.Reset()
				n = 0
			}
		}
	}
	if n > 0 {
		_ = common.Pool.Submit(func() {
			esBulk(utils.CopyBytes(bodyBuf.Bytes()))
		})
	}
}

func esBulk(b []byte) {
	resp, err := common.ES.Bulk(bytes.NewReader(b))
	if err != nil {
		common.LogSampled.Error().Err(err).Bytes("body", b).Msg("es bulk")
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// 低级别日志配置时, 开启批量写入错误抽样日志
	if conf.Config.SYSConf.Log.Level < int(zerolog.WarnLevel) {
		var res common.TStringAnyMaps
		var blk tESBulkResponse

		if resp.IsError() {
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				common.LogSampled.Error().Err(err).
					Bytes("body", b).Str("resp", resp.String()).
					Msg("es bulk, parsing the response body")
			} else {
				common.LogSampled.Error().
					Bytes("body", b).Int("http_code", resp.StatusCode).
					Msgf("es bulk, err: %+v", res["error"])
			}

			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&blk); err != nil {
			common.LogSampled.Error().Err(err).
				Bytes("body", b).Str("resp", resp.String()).
				Msg("es bulk, parsing the response body")
		} else if blk.Errors {
			for _, d := range blk.Items {
				if d.Index.Status > 201 {
					common.LogSampled.Error().Bytes("body", b).
						Msgf("error: [%d] %s; %s; %s; %s",
							d.Index.Status,
							d.Index.Error.Type,
							d.Index.Error.Reason,
							d.Index.Error.Cause.Type,
							d.Index.Error.Cause.Reason)
				}
			}
		}
	}
}

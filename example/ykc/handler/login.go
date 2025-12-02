package handler

import (
	"fmt"

	rot "github.com/longan55/Rules-over-TCP"
)

func CheckLogin(fh *rot.FunctionHandler) {
	fh.AddField("pile_code", rot.WithBcd(), rot.WithLength(7), rot.WithString())
	fh.AddField("pile_type",
		rot.WithBin(),
		rot.WithLength(1),
		rot.WithInteger(false, 1, 0),
		rot.WithEnum("未知桩类型", map[int]any{
			0: "直流桩",
			1: "交流桩",
		}))
	fh.AddField("gun_number", rot.WithBin(), rot.WithLength(1), rot.WithInteger(false, 1, 0))
	fh.AddField("protocol_version", rot.WithBin(), rot.WithLength(1), rot.WithInteger(false, 1, 0))
	fh.AddField("program_version", rot.WithAscii(), rot.WithLength(8), rot.WithString())
	fh.AddField("net_type", rot.WithBin(), rot.WithLength(1), rot.WithInteger(false, 1, 0), rot.WithEnum(
		"未知网络类型", map[int]any{
			0: "SIM卡",
			1: "LAN",
			2: "WAN",
			3: "其他",
		},
	))
	fh.AddField("sim_number", rot.WithBcd(), rot.WithLength(10), rot.WithString())
	fh.AddField("operator", rot.WithBin(), rot.WithLength(1), rot.WithInteger(false, 1, 0), rot.WithEnum(
		"未知运营商", map[int]any{
			0: "移动",
			2: "电信",
			3: "联通",
			4: "其他",
		},
	))
	fh.SetHandler(func(parsedData map[string]rot.ParsedData) error {
		for fieldName, parsedData := range parsedData {
			fmt.Printf("[%s]:RealValue:%v; Expected:%v; Source:%# 02X\n", fieldName, parsedData.Origin, parsedData.Explained, parsedData.Bytes)
		}
		return nil
	})
}

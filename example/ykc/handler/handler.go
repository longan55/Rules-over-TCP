package handler

import (
	rot "github.com/longan55/Rules-over-TCP"
)

func RegisterHandlers(builder *rot.ProtocolBuilder) {
	builder.HandleFunc(rot.FunctionCode(0x01), CheckLogin)
}

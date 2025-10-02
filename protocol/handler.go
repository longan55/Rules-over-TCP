package protocol

//功能结构体实现这个接口
// type Handler interface {
// 	Hanlde(data []byte) (map[string]any, error)
// }

func HandlerTest(data []byte) (map[string]any, error) {
	return map[string]any{"test": data[:]}, nil
}

// 功能码,长度应该更加广泛,但暂时使用1字节
type FunctionCode byte

type Handler func(data []byte) (map[string]any, error)

type HandlerConfig struct {
	handlerMap map[FunctionCode]Handler
}

func (hc *HandlerConfig) AddHandler(fc FunctionCode, h Handler) {
	hc.handlerMap[fc] = h
}

func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		handlerMap: make(map[FunctionCode]Handler, 32),
	}
}

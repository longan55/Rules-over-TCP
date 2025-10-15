package protocol

import "errors"

var FakeParser = NewParser()

type Parser interface {
	Parse(data []byte) (map[string]interface{}, error)
	AddItem(name string, item CodecItem)
}

type ParserImpl struct {
	items []struct {
		name string
		item CodecItem
	}
	totalLength int
}

func NewParser() Parser {
	return &ParserImpl{}
}

func (impl *ParserImpl) Parse(data []byte) (map[string]interface{}, error) {
	if len(data) < impl.totalLength {
		return nil, errors.New("数据长度不足")
	}

	result := make(map[string]interface{}, len(impl.items))
	offset := 0

	for _, item := range impl.items {
		value, err := item.item.Decode(data[offset:])
		if err != nil {
			return nil, errors.New("解码" + item.name + "时出错: " + err.Error())
		}
		result[item.name] = value
		offset += item.item.GetLength()
	}

	return result, nil
}

func (impl *ParserImpl) AddItem(name string, item CodecItem) {
	impl.items = append(impl.items, struct {
		name string
		item CodecItem
	}{
		name: name,
		item: item,
	})
	impl.totalLength += item.GetLength()
}

package rot

import (
	"encoding/binary"
	"testing"
)

// func TestDecodeBinInteger(t *testing.T) {
// 	testmap := []struct {
// 		name   string
// 		order  binary.ByteOrder
// 		input  []byte
// 		output int
// 	}{
// 		// 1字节测试用例
// 		{"1字节-大端-最小值", binary.BigEndian, []byte{0x00}, 0},
// 		{"1字节-大端-最大值", binary.BigEndian, []byte{0xFF}, -1}, // 有符号1字节最大值是0x7F，但这里假设是按补码处理
// 		{"1字节-大端-中间值1", binary.BigEndian, []byte{0x40}, 64},
// 		{"1字节-大端-中间值2", binary.BigEndian, []byte{0x80}, -128},
// 		{"1字节-大端-小正数", binary.BigEndian, []byte{0x01}, 1},
// 		{"1字节-大端-小负数", binary.BigEndian, []byte{0xFE}, -2},

// 		{"1字节-小端-最小值", binary.LittleEndian, []byte{0x00}, 0},
// 		{"1字节-小端-最大值", binary.LittleEndian, []byte{0xFF}, -1},
// 		{"1字节-小端-中间值1", binary.LittleEndian, []byte{0x40}, 64},
// 		{"1字节-小端-中间值2", binary.LittleEndian, []byte{0x80}, -128},

// 		// 2字节测试用例
// 		{"2字节-大端-最小值", binary.BigEndian, []byte{0x00, 0x00}, 0},
// 		{"2字节-大端-最大值", binary.BigEndian, []byte{0xFF, 0xFF}, -1},
// 		{"2字节-大端-中间值1", binary.BigEndian, []byte{0x80, 0x00}, -32768},
// 		{"2字节-大端-中间值2", binary.BigEndian, []byte{0x7F, 0xFF}, 32767},
// 		{"2字节-大端-正数", binary.BigEndian, []byte{0x12, 0x34}, 4660},
// 		{"2字节-大端-负数", binary.BigEndian, []byte{0xED, 0xCB}, -4661},

// 		{"2字节-小端-最小值", binary.LittleEndian, []byte{0x00, 0x00}, 0},
// 		{"2字节-小端-最大值", binary.LittleEndian, []byte{0xFF, 0xFF}, -1},
// 		{"2字节-小端-中间值1", binary.LittleEndian, []byte{0x00, 0x80}, -32768},
// 		{"2字节-小端-中间值2", binary.LittleEndian, []byte{0xFF, 0x7F}, 32767},
// 		{"2字节-小端-正数", binary.LittleEndian, []byte{0x34, 0x12}, 4660},
// 		{"2字节-小端-负数", binary.LittleEndian, []byte{0xCB, 0xED}, -4661},

// 		// 4字节测试用例
// 		{"4字节-大端-最小值", binary.BigEndian, []byte{0x00, 0x00, 0x00, 0x00}, 0},
// 		{"4字节-大端-最大值", binary.BigEndian, []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
// 		{"4字节-大端-中间值1", binary.BigEndian, []byte{0x80, 0x00, 0x00, 0x00}, -2147483648},
// 		{"4字节-大端-中间值2", binary.BigEndian, []byte{0x7F, 0xFF, 0xFF, 0xFF}, 2147483647},
// 		{"4字节-大端-正数", binary.BigEndian, []byte{0x12, 0x34, 0x56, 0x78}, 305419896},
// 		{"4字节-大端-负数", binary.BigEndian, []byte{0xED, 0xCB, 0xA9, 0x88}, -305419896},

// 		{"4字节-小端-最小值", binary.LittleEndian, []byte{0x00, 0x00, 0x00, 0x00}, 0},
// 		{"4字节-小端-最大值", binary.LittleEndian, []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
// 		{"4字节-小端-中间值1", binary.LittleEndian, []byte{0x00, 0x00, 0x00, 0x80}, -2147483648},
// 		{"4字节-小端-中间值2", binary.LittleEndian, []byte{0xFF, 0xFF, 0xFF, 0x7F}, 2147483647},
// 		{"4字节-小端-正数", binary.LittleEndian, []byte{0x78, 0x56, 0x34, 0x12}, 305419896},
// 		{"4字节-小端-负数", binary.LittleEndian, []byte{0x88, 0xA9, 0xCB, 0xED}, -305419896},

// 		// 8字节测试用例（注意：int类型在某些系统上可能只有4字节，这里假设测试环境支持64位整数）
// 		{"8字节-大端-最小值", binary.BigEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0},
// 		{"8字节-大端-最大值", binary.BigEndian, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, -1},
// 		{"8字节-大端-中间值1", binary.BigEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42}, 66},
// 		{"8字节-大端-中间值2", binary.BigEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42, 0x00}, 16896},
// 		{"8字节-大端-正数", binary.BigEndian, []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}, 1311768467463790320},  // 可能会被截断，取决于系统
// 		{"8字节-大端-负数", binary.BigEndian, []byte{0xED, 0xCB, 0xA9, 0x87, 0x65, 0x43, 0x21, 0x10}, -1311768467463790320}, // 可能会被截断

// 		{"8字节-小端-最小值", binary.LittleEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0},
// 		{"8字节-小端-最大值", binary.LittleEndian, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, -1},
// 		{"8字节-小端-中间值1", binary.LittleEndian, []byte{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 66},
// 		{"8字节-小端-中间值2", binary.LittleEndian, []byte{0x00, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 16896},
// 		{"8字节-小端-正数", binary.LittleEndian, []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}, 1311768467463790320},  // 可能会被截断
// 		{"8字节-小端-负数", binary.LittleEndian, []byte{0x10, 0x21, 0x43, 0x65, 0x87, 0xA9, 0xCB, 0xED}, -1311768467463790320}, // 可能会被截断
// 	}
// 	for no, v := range testmap {
// 		decoder := NewDecoder(v.order)
// 		i := decoder.BIN().SetByteLength(len(v.input)).Integer()
// 		src := i.SourceValue(v.input)
// 		t.Logf("[%s] 测试用例 %d, 输入:%v, 输出值:%v, 期望输出:%v\n", v.name, no+1, v.input, src, v.output)
// 		if src != v.output {
// 			t.Errorf("[%s] 测试失败! 输入:%v, 实际输出:%v, 期望输出:%v", v.name, v.input, src, v.output)
// 		}
// 	}
// }

// func TestDecodeBinString(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	s := decoder.BIN().SetByteLength(4).String1()
// 	src := s.SourceValue([]byte{0, 0, 0, 1})
// 	t.Logf("source value:%v,Type:%T\n", src, src)
// }

// func TestDecodeBinEnum(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	i := decoder.BIN().SetByteLength(1).Integer().SetEnum(map[int]any{
// 		0: "A",
// 		1: "B",
// 		2: "C",
// 		3: "D",
// 	})
// 	explained := i.ExplainedValue([]byte{2})
// 	t.Logf("explained value:%v\n", explained)
// }

// func TestDecodeBinBitMap(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	i := decoder.BIN().SetByteLength(1).Integer().SetBitMap(map[int]any{
// 		0: "A",
// 		1: "B",
// 		2: "C",
// 		3: "D",
// 	})
// 	explained := i.ExplainedValue([]byte{0b1011})
// 	t.Logf("explained value:%v\n", explained)
// }

// func TestBCDInteger_Integer(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	i := decoder.BCD().Integer()
// 	src := i.SourceValue([]byte{0x31, 0x41, 0x59, 0x26})
// 	t.Logf("source value:%v,Type:%T\n", src, src)

// 	decoder1 := NewDecoder(binary.LittleEndian)
// 	i1 := decoder1.BCD().Integer()
// 	src1 := i1.SourceValue([]byte{0x31, 0x41, 0x59, 0x26})
// 	t.Logf("source value:%v,Type:%T\n", src1, src1)
// }

// func TestBCDFloat_DecimalPlace(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	f := decoder.BCD().Float().DecimalPlace(2)
// 	src := f.SourceValue([]byte{0x31, 0x41, 0x59, 0x26})
// 	t.Logf("source value:%v,Type:%T\n", src, src)
// }

// func TestBCDString_String(t *testing.T) {
// 	decoder := NewDecoder(binary.BigEndian)
// 	s := decoder.BCD().String()
// 	src := s.SourceValue([]byte{0x31, 0x41, 0x59, 0x26})
// 	t.Logf("source value:%v,Type:%T\n", src, src)

// 	decoder1 := NewDecoder(binary.LittleEndian)
// 	s1 := decoder1.BCD().String()
// 	src1 := s1.SourceValue([]byte{0x31, 0x41, 0x59, 0x26})
// 	t.Logf("source value:%v,Type:%T\n", src1, src1)
// }

func TestA(t *testing.T) {
	h := new(FucntionHandler)
	h.NewDecoder("a", binary.BigEndian).BIN().SetByteLength(4).Integer()
	h.NewDecoder("b", binary.BigEndian).BIN().SetByteLength(4).Integer()
	h.NewDecoder("c", binary.BigEndian).BIN().SetByteLength(2).Integer()
	h.NewDecoder("d", binary.BigEndian).BIN().SetByteLength(2).Float1().Multiple(0.01)
	h.NewDecoder("e", binary.BigEndian).BIN().SetByteLength(1).Integer().SetEnum(map[int]any{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
	})
	//2147483647,-2147483648
	result, err := h.Parse([]byte{0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01})
	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}
	t.Logf("result:%v\n", result)
}

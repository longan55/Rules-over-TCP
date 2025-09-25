package protocol

//通过框架配置协议，框架自动解析和封装，无需自己开发。
//1. 定义一个结构体(Data Unit Handler)包含协议组成元素信息
//2. 每次解析和组装都要使用这个结构体(DPH)
//3. DPH需要一个解析数据获取Data域的方法，返回Data域的字节切片
//4. 定义功能码接口(Function),解析数据域和组装数据域
//5. 需要提供多种元素的默认处理
//6. 提供方便扩展元素的接口

// func NewDUHBuilder() DUHBuilder {
// 	return DUHBuilder{
// 		dph: &DPH{
// 			Fields: make([]Fielder, 0, 3),
// 		},
// 	}
// }

//协议元素组成规则
//1. 第一个元素必须是起始符
//2. 第一个元素到数据域之前的元素组成 消息头部元素.
//3. 数据域称为消息体元素
//4. 最后一个元素必须是校验码元素

//注册回调函数,根据功能码,返回对应的功能结构体

// type DUHBuilder struct {
// 	dph *DPH
// }

// func (dphBuilder *DUHBuilder) AddFielder(field Fielder) *DUHBuilder {
// 	dphBuilder.dph.Fields = append(dphBuilder.dph.Fields, field)
// 	return dphBuilder
// }

// func (dphBuilder *DUHBuilder) Build() DataUnitHandler {
// 	//todo 起始码+长度码 的长度
// 	return dphBuilder.dph
// }

// type DataUnitHandler interface {
// 	Handle(ctx context.Context, conn net.Conn)
// 	SetDataLength(length int)
// 	Parse(adu []byte) (Function, error)
// 	Serialize(f Function) []byte
// }

// func NewDPH() DPH {
// 	return DPH{Fields: make([]Fielder, 0, 3)}
// }

// var _ DataUnitHandler = (*DPH)(nil)

// // dph 应用数据单元 结构体
// type DPH struct {
// 	dataLength int
// 	conn       net.Conn
// 	Fields     []Fielder
// }

// func (dph *DPH) Handle(ctx context.Context, conn net.Conn) {
// 	dph.conn = conn
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			//停止读取
// 			return
// 		default:
// 			alldata := make([][]byte, 0, len(dph.Fields))
// 			//第一遍遍历fields, 读取一个完整的数据单元
// 			for _, field := range dph.Fields {
// 				//定义好合适长度的buf,接收该元素数据
// 				var buf []byte
// 				if field.Type() != DATA {
// 					buf = make([]byte, field.Length())
// 				} else {
// 					buf = make([]byte, dph.dataLength)
// 				}
// 				//读取
// 				_, err := io.ReadFull(dph.conn, buf)
// 				if err != nil {
// 					fmt.Println("读取数据失败:", err)
// 					return
// 				}
// 				//检查起始符,如果不对就没必要按正确元素顺序处理了
// 				if field.Type() == START {
// 					//如果是起始符,校验起始符
// 					if bytes.Equal(buf, field.GetDefault()) {
// 						fmt.Println("起始符校验失败")
// 						break
// 					}
// 				}
// 				//如果是长度元素,解析长度元素,赋值给数据长度变量
// 				if field.Type() == LENGTH {
// 					length, err := field.Deal(alldata)
// 					if err != nil {
// 						fmt.Println("读取数据长度失败:", err)
// 						break
// 					}
// 					dph.dataLength = length.(int)
// 				}
// 				//将数据拼接
// 				alldata = append(alldata, buf)
// 			}
// 			//第二遍遍历fields, 解析数据单元
// 			for _, field := range dph.Fields {
// 				if field.Type() == START {
// 					continue
// 				}
// 				_, err := field.Deal(alldata)
// 				if err != nil {
// 					fmt.Println("数据解析失败:", err)
// 					break
// 				}
// 			}
// 			fmt.Println("读取数据:", alldata)
// 			//todo回调
// 		}
// 	}
// }

// func (dph *DPH) SetDataLength(length int) {
// 	dph.dataLength = length
// }

// func (dph *DPH) Parse(adu []byte) (Function, error) {
// 	for _, v := range dph.Fields {
// 		v.Deal(nil)
// 	}
// 	return nil, nil
// }

// func (dph *DPH) Serialize(f Function) []byte {
// 	return nil
// }

// func (dph *DPH) Info() {
// 	for _, v := range dph.Fields {
// 		of := reflect.TypeOf(v)
// 		fmt.Println("类型:", of, " 长度:", v.Length())
// 	}
// }

// // Debug 解析数据
// func (dph *DPH) Debug(r io.Reader, source []byte) {
// 	// 起始符 只需要判断是否相等
// 	// 数据域长度 要传给数据域元素作为长度
// 	// 加密标志 是否对指定元素的值进行加密或解密
// 	// 校验码 是否对指定元素进行校验计算
// 	// offset := 0
// 	// //遍历所有元素
// 	// for _, field := range dph.Fields {
// 	// 	//根据元素Field获取对应数据切片
// 	// 	data := source[offset : offset+field.Length()]
// 	// 	//更新偏移量
// 	// 	offset += field.Length()
// 	// 	//debug打印元素
// 	// 	if field.GetScale() == 0 {
// 	// 		fmt.Printf("[%s] = %0d", field.GetName(), data)
// 	// 	} else {
// 	// 		fmt.Printf("[%s] = %0x", field.GetName(), data)
// 	// 	}

// 	// 	//处理方法
// 	// 	_, err := field.Deal(data)
// 	// 	if err != nil { //log.Println("数据解析出错! [error]:", err)
// 	// 		fmt.Printf("数据解析出错! [error]: %v\n", err)
// 	// 	}
// 	// }
// }

// Fielder 元素接口
// type Fielder interface {
// 	GetIndex() int
// 	SetIndex(index int)
// 	//获取元素名称
// 	GetName() string
// 	SetName(name string)
// 	//获取元素类型
// 	Type() FieldType
// 	SetDefault(value []byte)
// 	GetDefault() []byte
// 	//获取实际值
// 	RealValue() []byte
// 	// SetLen 设置元素长度
// 	SetLen(int)
// 	// Length 获取元素长度
// 	Length() int
// 	SetScale(uint8)
// 	//获取进制
// 	GetScale() uint8
// 	//获取大小端
// 	GetOrder() binary.ByteOrder
// 	SetOrder(order binary.ByteOrder)
// 	// SetRange 设置范围
// 	SetRange(start, end uint8)
// 	// GetRange 获取范围
// 	GetRange() (start, end uint8)
// 	// Deal 解析元素
// 	Deal([][]byte) (any, error)
// }

// type FieldType byte

// const (
// 	// 起始符
// 	START FieldType = iota
// 	// 数据域长度
// 	LENGTH
// 	// 功能码
// 	FUNCTION
// 	// 数据域
// 	DATA
// 	// 校验码
// 	CHECK
// )

// var _ Fielder = (*Field)(nil)

// // Field 基础元素结构体
// type Field struct {
// 	//元数据: 存储该元素的元数据(用于描述说明)
// 	index    int                                             //说明该元素的索引
// 	Typ      FieldType                                       //元素类型
// 	name     string                                          //元素名字
// 	scale    uint8                                           // 1十六进制，0十进制
// 	len      int                                             //元素本身长度
// 	defaultV []byte                                          //默认值
// 	order    binary.ByteOrder                                //大小端
// 	start    uint8                                           //开始索引: 该元素影响的元素区域的第一个元素索引
// 	end      uint8                                           //结束索引: 该元素影响的元素区域的最后一个元素索引
// 	DealFunc func(field Fielder, data [][]byte) (any, error) //处理函数
// 	//临时数据: 存储当前adu的数据
// 	realData   []byte
// 	parsedData any
// }

// func (f *Field) GetIndex() int {
// 	return f.index
// }

// func (f *Field) SetIndex(index int) {
// 	f.index = index
// }

// func (f *Field) GetName() string {
// 	return f.name
// }

// func (f *Field) SetName(name string) {
// 	f.name = name
// }

// func (f *Field) Type() FieldType {
// 	return f.Typ
// }

// func (f *Field) SetDefault(val []byte) {
// 	f.defaultV = val
// }

// func (f *Field) GetDefault() []byte {
// 	return f.defaultV
// }

// func (f *Field) RealValue() []byte {
// 	return f.defaultV
// }
// func (f *Field) SetLen(l int) {
// 	f.len = l
// }
// func (f *Field) Length() int {
// 	return f.len
// }

// func (f *Field) SetScale(u uint8) {
// 	f.scale = u
// }

// func (f *Field) GetScale() uint8 {
// 	return f.scale
// }

// func (f *Field) GetOrder() binary.ByteOrder {
// 	return f.order
// }

// func (f *Field) SetOrder(order binary.ByteOrder) {
// 	f.order = order
// }

// func (f *Field) GetRange() (start, end uint8) {
// 	return f.start, f.end
// }
// func (f *Field) SetRange(start, end uint8) {
// 	f.start = start
// 	f.end = end
// }

// func (f *Field) Deal(data [][]byte) (any, error) {
// 	return f.DealFunc(f, data)
// }

// 起始符
// func NewStarter(start []byte) Fielder {
// 	field := &Field{
// 		name:     "起始符",
// 		defaultV: start,
// 		len:      len(start),
// 	}
// 	field.DealFunc = func(field Fielder, data [][]byte) (any, error) {
// 		if data == nil {
// 			return nil, errors.New("数据为空")
// 		}
// 		if len(data) < field.Length() {
// 			return nil, errors.New("数据长度小于起始符长度")
// 		}
// 		if bytes.Equal(data[0][:field.Length()], field.GetDefault()) {
// 			return nil, fmt.Errorf("起始符错误Need:%s,But:%s", string(field.GetDefault()), string(data[0][:field.Length()]))
// 		}
// 		return nil, nil
// 	}
// 	return field
// }

// func NewDataLen(length int) Fielder {
// 	field := &Field{
// 		name:     "数据域长度",
// 		defaultV: nil,
// 		len:      length,
// 	}
// 	field.DealFunc = func(field Fielder, data [][]byte) (any, error) {
// 		if data == nil {
// 			return nil, errors.New("数据为空")
// 		}
// 		if len(data) < field.Length() {
// 			return nil, errors.New("数据长度小于数据域长度字段长度")
// 		}
// 		lenData := data[field.GetIndex()]
// 		u64, err := BIN2Uint64(lenData, field.GetOrder())
// 		if err != nil {
// 			return nil, err
// 		}
// 		return u64, nil
// 	}
// 	return field
// }

// func NewFuncCode() Fielder {
// 	field := &Field{
// 		name:     "功能码",
// 		defaultV: nil,
// 		len:      1,
// 	}
// 	field.DealFunc = func(field Fielder, data [][]byte) (any, error) {
// 		if data == nil {
// 			return nil, errors.New("数据为空")
// 		}
// 		if len(data) < field.Length() {
// 			return nil, errors.New("数据长度小于功能码字段长度")
// 		}
// 		return data[field.GetIndex()], nil
// 	}
// 	return field
// }

// type Starter struct {
// 	Field
// }

// func (start Starter) Deal(data []byte) error {
// 	if len(data) != int(start.len) {
// 		return errors.New("起始长度不对")
// 	}
// 	for i, v := range start.defaultV {
// 		if v != data[i] {
// 			fmt.Printf("起始：%# 02x，预期：%# 02x\n", data, start.defaultV)
// 			return errors.New("起始值错误")
// 		}
// 	}
// 	fmt.Printf("[起始值]:% 02X\n", data)
// 	return nil
// }
// type DataLen struct {
// 	Field
// }

// func (d DataLen) Deal(data []byte) error {
// 	if len(data) != int(d.len) {
// 		return errors.New("数据域长度字段本省身长度不对")
// 	}
// 	var l = make([]byte, d.len)
// 	data = data[:d.len]
// 	buffer := bytes.NewBuffer(data)
// 	err := binary.Read(buffer, d.order, l)
// 	if err != nil {
// 		fmt.Println("binary read error", err)
// 		return err
// 	}
// 	bin2Uint64, err := BIN2Uint64(data, d.order)
// 	if err != nil {
// 		fmt.Println("b2i error: ", err)
// 		return err
// 	}
// 	fmt.Printf("[数据长度]:%d字节\n", bin2Uint64)
// 	return nil
// }

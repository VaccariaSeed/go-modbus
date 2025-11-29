package go_modbus

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"reflect"
)

const (
	modbusRtu = "rtu"
	modbusTcp = "tcp"
)

var mrFuncCodes []byte

func init() {
	mrFuncCodes = []byte{ReadCoils, ReadDiscreteInputs, ReadHoldingRegisters, ReadInputRegisters, WriteSingleCoil, WriteSingleRegister, WriteMultipleCoils, WriteMultipleRegisters}
}

const (
	ReadCoils              byte = 0x01 //读线圈,位,取得一组逻辑线圈的当前状态(ON/OFF)
	ReadDiscreteInputs     byte = 0x02 //读离散输入寄存器,位,取得一组开关输入的当前状态(ON/OFF)
	ReadHoldingRegisters   byte = 0x03 //读保持寄存器,整型、浮点型、字符型,在一个或多个保持寄存器中取得当前的二进制值
	ReadInputRegisters     byte = 0x04 //读输入寄存器,整型、浮点型,在一个或多个输入寄存器中取得当前的二进制值
	WriteSingleCoil        byte = 0x05 //写单个线圈寄存器,位,强置一个逻辑线圈的通断状态
	WriteSingleRegister    byte = 0x06 //写单个保持寄存器,整型、浮点型、字符型,把具体二进制值装入一个保持寄存器
	WriteMultipleCoils     byte = 0x0F //写多个线圈寄存器,位,强置一串连续逻辑线圈的通断
	WriteMultipleRegisters byte = 0x10 //写多个保持寄存器,整型、浮点型、字符型,把具体的二进制值装入一串连续的保持寄存器
)

type ModbusStatuter interface {
	baseDecode(buf *bufio.Reader) error
	DecodeMasterFrame(frame []byte) error       //解码主站发来的报文
	DecodeMasterReader(buf *bufio.Reader) error //解码主站发来的报文
	DecodeSlaveFrame(frame []byte) error        //解码从站响应的报文
	DecodeSlaveReader(buf *bufio.Reader) error  //解码从站响应的报文
	Data() []byte                               //获取数据域
	Encode() ([]byte, error)                    //编码
	cs(frame []byte) []byte                     //计算cs
	OriginalFrame() []byte                      //获取原始报文
	Copy() ModbusStatuter                       //深拷贝
	SlaveId() byte
	FuncCode() byte
	DataParser() *ModbusDataParser //获取一个解析器
}

type Direction string

const (
	Master Direction = "master"
	Slave  Direction = "slave"
)

func NewModbusRTUDecoder[T ModbusStatuter](direction Direction, slaveId byte, ch chan<- T) *ModbusDecoder[T] {
	ctx, cancelFunc := context.WithCancel(context.Background())
	var decoder T
	decoderType := reflect.TypeOf(decoder)
	if decoderType.Kind() == reflect.Ptr {
		// 如果 T 是指针类型，创建指向的实例
		elemType := decoderType.Elem()
		decoder = reflect.New(elemType).Interface().(T)
	} else {
		// 如果 T 是值类型，创建零值
		decoder = reflect.Zero(decoderType).Interface().(T)
	}
	return &ModbusDecoder[T]{
		dataDirection: direction,
		slaveId:       slaveId,
		ch:            ch,
		decoder:       decoder,
		ctx:           ctx,
		cancel:        cancelFunc,
	}
}

type ModbusDecoder[T ModbusStatuter] struct {
	dataDirection Direction //报文方向
	slaveId       byte      //地址，如果是0，则只在乎报文正确；如果不是0，则即在意报文正确也会比较slaveId
	ch            chan<- T
	decoder       T //解码器
	ctx           context.Context
	cancel        context.CancelFunc
}

// StreamDecoder 流式解码，每解析成功一个，则传入到ch中，然后继续解析
func (m *ModbusDecoder[T]) StreamDecoder(buf *bufio.Reader) error {
	var err error
	for {
		select {
		case <-m.ctx.Done():
			m.ctx, m.cancel = context.WithCancel(context.Background())
			return nil
		default:
			err = m.stream(buf)
			if err != nil {
				return err
			}
		}
	}
}

func (m *ModbusDecoder[T]) Stop() {
	m.cancel()
}

func (m *ModbusDecoder[T]) stream(buf *bufio.Reader) error {
	var err error
	if m.dataDirection == Slave {
		err = m.decoder.DecodeMasterReader(buf)
	} else if m.dataDirection == Master {
		err = m.decoder.DecodeSlaveReader(buf)
	} else {
		return fmt.Errorf("invalid direction:%v", m.dataDirection)
	}
	if err != nil {
		if err == io.EOF {
			return err
		}
		return nil
	}
	frame := m.decoder.Copy()
	if result, ok := any(frame).(T); ok {
		if m.slaveId == 0 || m.slaveId == frame.SlaveId() {
			m.ch <- result
		}
	}
	return nil
}

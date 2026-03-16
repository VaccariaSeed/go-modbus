package statute

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type CoilStatus bool

const (
	ON  CoilStatus = true //线圈开
	OFF CoilStatus = false
)

// 中介
type intermediary struct {
	ident    uint16 //唯一标识
	slaveId  byte
	funcCode byte //功能码
}

func (i *intermediary) ObtainIntermediary() *intermediary {
	return i
}

func (i *intermediary) data4ParseRequest(funcCode byte, data []byte) (addr, number uint16, err error) {
	if i.funcCode != funcCode {
		return 0, 0, errors.New("funcCode mismatch")
	}
	if len(data) != 4 {
		return 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(data[:2])
	number = binary.BigEndian.Uint16(data[2:4])
	return addr, number, nil
}

func (i *intermediary) data4ParseResponse(funcCode byte, result []byte) (length byte, data []byte, err error) {
	if i.funcCode != funcCode {
		return 0, nil, errors.New("funcCode mismatch")
	}
	if len(result) == 0 {
		return 0, nil, errors.New("invalid data length")
	}
	return data[0], data[1:], nil
}

func (i *intermediary) parseCoilsResponse(bytes []byte, number uint16) (length uint16, result []CoilStatus, err error) {
	if bytes == nil || len(bytes) == 0 {
		return 0, nil, errors.New("invalid response, response is empty")
	}
	var strResult []string
	for _, b := range bytes {
		binaryStr := fmt.Sprintf("%08b", b)
		// 将每个字符转换为单独的字符串
		bs := make([]string, 8)
		for i, ch := range binaryStr {
			bs[7-i] = string(ch)
		}
		strResult = append(strResult, bs...)
	}
	if uint16(len(strResult)) < number {
		return 0, nil, errors.New("invalid response, response is error")
	}
	result = make([]CoilStatus, number)
	for index := uint16(0); index < number; index++ {
		result[index] = strResult[index] == "1"
	}
	return uint16(len(result)), result, nil
}

// ParseReadCoilsRequest 解析读线圈的请求
// 参数：
// data 数据域
// 返回值：
// addr 起始地址
// number 寄存器数量
func (i *intermediary) ParseReadCoilsRequest(data []byte) (addr, number uint16, err error) {
	return i.data4ParseRequest(ReadCoils, data)
}

// ParseReadCoilsResponse 解析读线圈的响应
// 参数：
// data 数据域
// 返回值：
// length 实际长度
// result 实际结果
func (i *intermediary) ParseReadCoilsResponse(number uint16, data []byte) (length uint16, result []CoilStatus, err error) {
	return i.parseCoilsResponse(data, number)
}

// ParseReadDiscreteInputsRequest 解析读离散输入寄存器的请求
// 参数：
// data 数据域
// 返回值：
// addr 起始地址
// number 寄存器数量
func (i *intermediary) ParseReadDiscreteInputsRequest(data []byte) (addr, number uint16, err error) {
	return i.data4ParseRequest(ReadDiscreteInputs, data)
}

// ParseReadDiscreteInputsResponse 解析读离散输入寄存器的响应
// length 长度
// data 数据
func (i *intermediary) ParseReadDiscreteInputsResponse(number uint16, data []byte) (length uint16, result []CoilStatus, err error) {
	return i.parseCoilsResponse(data, number)
}

// ParseReadHoldingRegistersRequest 解析读保持寄存器的请求
// 参数：
// data 数据域
// 返回值：
// addr 起始地址
// number 寄存器数量
func (i *intermediary) ParseReadHoldingRegistersRequest(data []byte) (addr, number uint16, err error) {
	return i.data4ParseRequest(ReadHoldingRegisters, data)
}

// ParseReadInputRegistersRequest 解析读输入寄存器的请求
// 参数：
// data 数据域
// 返回值：
// addr 起始地址
// number 寄存器数量
func (i *intermediary) ParseReadInputRegistersRequest(data []byte) (addr, number uint16, err error) {
	return i.data4ParseRequest(ReadInputRegisters, data)
}

// ParseWriteSingleCoil 解析写单个线圈的请求或响应
// addr 起始地址
// status 状态
func (i *intermediary) ParseWriteSingleCoil(data []byte) (addr, status uint16, err error) {
	return i.data4ParseRequest(WriteSingleCoil, data)
}

// ParseWriteSingleRegister 解析写单个保持寄存器的请求或响应
// addr 起始地址
// status 状态
func (i *intermediary) ParseWriteSingleRegister(data []byte) (addr, status uint16, err error) {
	return i.data4ParseRequest(WriteSingleRegister, data)
}

// ParseWriteMultipleCoilsRequest 解析写多个线圈的请求
// addr 起始地址
// number 寄存器数量
// length 字节数
// result 数据
func (i *intermediary) ParseWriteMultipleCoilsRequest(data []byte) (addr, number uint16, length byte, result []byte, err error) {
	if i.funcCode != WriteMultipleCoils {
		return 0, 0, 0, nil, errors.New("funcCode mismatch")
	}
	if len(data) < 6 {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(data[:2])
	number = binary.BigEndian.Uint16(data[2:4])
	length = data[4]
	data = data[5:]
	if len(data) != int(length) {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	return
}

// ParseWriteMultipleCoilsResponse 解析写多个线圈的响应
// addr 起始地址
// number 寄存器数量
// length 字节数
func (i *intermediary) ParseWriteMultipleCoilsResponse(data []byte) (addr, number uint16, err error) {
	if i.funcCode != WriteMultipleCoils {
		return 0, 0, errors.New("funcCode mismatch")
	}
	if len(data) < 4 {
		return 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(data[:2])
	number = binary.BigEndian.Uint16(data[2:4])
	return
}

// ParseWriteMultipleRegistersResponse 解析写多个保持寄存器的响应
// addr 起始地址
// number 寄存器数量
func (i *intermediary) ParseWriteMultipleRegistersResponse(data []byte) (addr, number uint16, err error) {
	return i.data4ParseRequest(WriteMultipleRegisters, data)
}

// PraseWriteMultipleRegisters 解析写多个保持寄存器的响应
func (i *intermediary) PraseWriteMultipleRegisters(data []byte) (uint16, uint16, error) {
	return i.data4ParseRequest(WriteMultipleRegisters, data)
}

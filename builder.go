package go_modbus

import (
	"encoding/binary"
)

// modbusFrameBuilder RTU报文构造器
type modbusFrameBuilder struct {
}

// 生成读线圈请求的数据域, 功能码0x01
// addr 寄存器起始地址
// number 寄存器数量
func (m *modbusFrameBuilder) buildReadCoilsRequest(addr, number uint16) []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint16(data[0:2], addr)
	binary.BigEndian.PutUint16(data[2:4], number)
	return data
}

// 生成读线圈响应的数据域, 功能码0x01
// status 每个线圈的状态，请按照顺序输入;不传数据会返回一个"从站设备忙"，当errResp为true时，功能码需要加0x80
func (m *modbusFrameBuilder) buildReadCoilsResponse(status ...bool) (data []byte, errResp bool) {
	if status == nil || len(status) == 0 {
		return []byte{0x06}, true
	}
	// 计算需要的字节数
	byteCount := (len(status) + 7) / 8
	data = make([]byte, byteCount)
	for i, coil := range status {
		if coil {
			// 计算字节索引和位索引
			byteIndex := i / 8
			bitIndex := uint(i % 8)
			// 设置对应的位 (Modbus使用小端位序：bit0=第一个线圈)
			data[byteIndex] |= 1 << bitIndex
		}
	}
	return append([]byte{byte(byteCount)}, data...), false

}

// 生成读离散输入寄存器的请求,功能码0x02
// addr 寄存器起始地址
// number 寄存器数量
func (m *modbusFrameBuilder) buildReadDiscreteInputsRequest(address, number uint16) []byte {
	return m.buildReadCoilsRequest(address, number)
}

// 生成读离散输入寄存器的数据域, 功能码0x02
// status 每个线圈的状态，请按照顺序输入;不传数据会返回一个"从站设备忙"，当errResp为true时，功能码需要加0x80
func (m *modbusFrameBuilder) buildReadDiscreteInputsResponse(status ...bool) (data []byte, errResp bool) {
	return m.buildReadCoilsResponse(status...)
}

// 生成读保持寄存器的请求，功能码0x03
// addr 寄存器起始地址
// number 寄存器数量
func (m *modbusFrameBuilder) buildReadHoldingInputsRequest(address, number uint16) []byte {
	return m.buildReadCoilsRequest(address, number)
}

// 生成读保持寄存器的数据域，功能码0x03
// value 每个寄存器的状态，请按照顺序输入;不传数据会返回一个"从站设备忙"，当errResp为true时，功能码需要加0x80
func (m *modbusFrameBuilder) buildReadHoldingInputsResponse(value ...uint16) (data []byte, errResp bool) {
	if value == nil || len(value) == 0 {
		return []byte{0x06}, true
	}
	// 每个 uint16 需要 2 个字节
	data = make([]byte, len(value)*2)

	for i, val := range value {
		// 计算当前 uint16 在结果切片中的位置
		start := i * 2
		end := start + 2
		binary.BigEndian.PutUint16(data[start:end], val)
	}
	return append([]byte{byte(len(data))}, data...), false
}

// 生成读输入寄存器的请求 功能码0x04
// addr 寄存器起始地址
// number 寄存器数量
func (m *modbusFrameBuilder) buildReadInputInputsRequest(address, number uint16) []byte {
	return m.buildReadCoilsRequest(address, number)
}

// 生成读输入寄存器的响应 功能码0x04
// value 每个寄存器的状态，请按照顺序输入;不传数据会返回一个"从站设备忙"，当errResp为true时，功能码需要加0x80
func (m *modbusFrameBuilder) buildReadInputInputsResponse(value ...uint16) (data []byte, errResp bool) {
	return m.buildReadHoldingInputsResponse(value...)
}

// 生成写单个线圈寄存器的请求或响应 功能码0x05
// addr 寄存器起始地址
// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
func (m *modbusFrameBuilder) buildWriteSingleCoil(address uint16, value bool) []byte {
	var val uint16 = 0x0000
	if value {
		val = 0xFF00
	}
	return m.buildReadCoilsRequest(address, val)
}

// 生成写单个保持寄存器的请求或响应 功能码0x06
// addr 寄存器起始地址
// value 设定值
func (m *modbusFrameBuilder) buildWriteSingleRegister(address, value uint16) []byte {
	return m.buildReadCoilsRequest(address, value)
}

// 生成写多个线圈的请求，功能码0x0F
// addr 寄存器起始地址
// number 寄存器数量
// status 线圈状态
func (m *modbusFrameBuilder) buildWriteMultipleCoilsRequest(address, number uint16, status ...bool) []byte {
	data, _ := m.buildReadCoilsResponse(status...)
	data1 := append(m.buildReadCoilsRequest(address, number), data...)
	return data1
}

// 生成写多个线圈的响应，功能码0x0F
// addr 寄存器起始地址
// length 字节数
func (m *modbusFrameBuilder) buildWriteMultipleCoilsResponse(address, length uint16) []byte {
	return m.buildReadCoilsRequest(address, length)
}

// 生成写多个保持寄存器的请求，功能码0x10
// addr 寄存器起始地址
// number 寄存器数量
// value 设定值
func (m *modbusFrameBuilder) buildWriteMultipleRegistersRequest(address, number uint16, value ...uint16) (data []byte, errResp bool) {
	data, errResp = m.buildReadHoldingInputsResponse(value...)
	if errResp {
		return data, errResp
	}
	data1 := append(m.buildReadCoilsRequest(address, number), data...)
	return data1, false
}

// 生成写多个保持寄存器的响应，功能码0x10
// addr 寄存器起始地址
// number 寄存器数量
func (m *modbusFrameBuilder) buildWriteMultipleRegistersResponse(address, number uint16) []byte {
	return m.buildReadCoilsRequest(address, number)
}

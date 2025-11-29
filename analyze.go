package go_modbus

import (
	"encoding/binary"
	"errors"
)

// ModbusDataParser 数据解析器
type ModbusDataParser struct {
	frame   ModbusStatuter
	statute string
}

func (m *ModbusDataParser) data4ParseRequest(funcCode byte) (addr, number uint16, err error) {
	if m.frame.FuncCode() != funcCode {
		return 0, 0, errors.New("funcCode mismatch")
	}
	if len(m.frame.Data()) != 4 {
		return 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.frame.Data()[:2])
	number = binary.BigEndian.Uint16(m.frame.Data()[2:4])
	return addr, number, nil
}

func (m *ModbusDataParser) data4ParseResponse(funcCode byte) (length byte, data []byte, err error) {
	if m.frame.FuncCode() != funcCode {
		return 0, nil, errors.New("funcCode mismatch")
	}
	if m.frame.Data() == nil || len(m.frame.Data()) == 0 {
		return 0, nil, errors.New("invalid data length")
	}
	return m.frame.Data()[0], m.frame.Data()[1:], nil
}

// ParseReadCoilsRequest 解析读线圈的请求
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadCoilsRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadCoils)
}

// ParseReadCoilsResponse 解析读线圈的响应
// data 数据
func (m *ModbusDataParser) ParseReadCoilsResponse() (length byte, data []byte, err error) {
	return m.data4ParseResponse(ReadCoils)
}

// ParseReadDiscreteInputsRequest 解析读离散输入寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadDiscreteInputsRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadDiscreteInputs)
}

// ParseReadDiscreteInputsResponse 解析读离散输入寄存器的响应
// length 长度
// data 数据
func (m *ModbusDataParser) ParseReadDiscreteInputsResponse() (length byte, data []byte, err error) {
	return m.data4ParseResponse(ReadDiscreteInputs)
}

// ParseReadHoldingRegistersRequest 解析读保持寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadHoldingRegistersRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadHoldingRegisters)
}

// ParseReadHoldingRegistersResponse 解析读保持寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadHoldingRegistersResponse() (length byte, data []uint16, err error) {
	return m.parseReadHoldingRegistersResponse(ReadHoldingRegisters)
}

func (m *ModbusDataParser) parseReadHoldingRegistersResponse(funcCode byte) (length byte, data []uint16, err error) {
	if m.frame.FuncCode() != funcCode {
		return 0, nil, errors.New("funcCode mismatch")
	}
	if m.frame.Data() == nil || len(m.frame.Data()) == 0 {
		return 0, nil, errors.New("invalid data length")
	}
	length = m.frame.Data()[0]
	if length != byte(len(m.frame.Data())-1) || length%2 != 0 {
		return 0, nil, errors.New("invalid data length")
	}
	resp := m.frame.Data()[1:]
	data = make([]uint16, len(resp)/2)
	for i := 0; i < len(m.frame.Data()[1:]); i += 2 {
		data[i/2] = binary.BigEndian.Uint16(resp[i : i+2])
	}
	return length, data, nil
}

// ParseReadInputRegistersRequest 解析读输入寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadInputRegistersRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadInputRegisters)
}

// ParseReadInputRegistersResponse 解析读输入寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseReadInputRegistersResponse() (length byte, data []uint16, err error) {
	return m.parseReadHoldingRegistersResponse(ReadInputRegisters)
}

// ParseWriteSingleCoil 解析写单个线圈的请求或响应
// addr 起始地址
// status 状态
func (m *ModbusDataParser) ParseWriteSingleCoil() (addr, status uint16, err error) {
	return m.data4ParseRequest(WriteSingleCoil)
}

// ParseWriteSingleRegister 解析写单个线圈的请求或响应
// addr 起始地址
// status 状态
func (m *ModbusDataParser) ParseWriteSingleRegister() (addr, status uint16, err error) {
	return m.data4ParseRequest(WriteSingleRegister)
}

// ParseWriteMultipleCoilsRequest 解析写多个线圈的请求
// addr 起始地址
// number 寄存器数量
// length 字节数
// data 数据
func (m *ModbusDataParser) ParseWriteMultipleCoilsRequest() (addr, number uint16, length byte, data []byte, err error) {
	if m.frame.FuncCode() != WriteMultipleCoils {
		return 0, 0, 0, nil, errors.New("funcCode mismatch")
	}
	if len(m.frame.Data()) < 6 {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.frame.Data()[:2])
	number = binary.BigEndian.Uint16(m.frame.Data()[2:4])
	length = m.frame.Data()[4]
	data = m.frame.Data()[5:]
	if len(data) != int(length) {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	return
}

// ParseWriteMultipleCoilsResponse 解析写多个线圈的响应
// addr 起始地址
// number 寄存器数量
// length 字节数
func (m *ModbusDataParser) ParseWriteMultipleCoilsResponse() (addr, number uint16, length byte, err error) {
	if m.frame.FuncCode() != WriteMultipleCoils {
		return 0, 0, 0, errors.New("funcCode mismatch")
	}
	if len(m.frame.Data()) < 5 {
		return 0, 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.frame.Data()[:2])
	number = binary.BigEndian.Uint16(m.frame.Data()[2:4])
	length = m.frame.Data()[4]
	return
}

// ParseWriteMultipleRegistersRequest 解析写多个保持寄存器的请求
// addr 起始地址
// number 寄存器数量
// length 字节数
// data 数据
func (m *ModbusDataParser) ParseWriteMultipleRegistersRequest() (addr, number uint16, length byte, data []uint16, err error) {
	if m.frame.FuncCode() != WriteMultipleRegisters {
		return 0, 0, 0, nil, errors.New("funcCode mismatch")
	}
	if len(m.frame.Data()) < 5 {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.frame.Data()[:2])
	number = binary.BigEndian.Uint16(m.frame.Data()[2:4])
	length = m.frame.Data()[4]
	resp := m.frame.Data()[5:]
	if length != byte(len(resp)) {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	data = make([]uint16, len(resp)/2)
	for i := 0; i < len(data); i += 2 {
		data[i/2] = binary.BigEndian.Uint16(resp[i : i+2])
	}
	return addr, number, length, data, nil
}

// ParseWriteMultipleRegistersResponse 解析写多个保持寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *ModbusDataParser) ParseWriteMultipleRegistersResponse() (addr, number uint16, err error) {
	return m.data4ParseRequest(WriteMultipleRegisters)
}

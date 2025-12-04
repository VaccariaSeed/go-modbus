package go_modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// modbusDataParser 数据解析器
type modbusDataParser struct {
	codec modbusStatute
}

func (m *modbusDataParser) data4ParseRequest(funcCode byte) (addr, number uint16, err error) {
	if m.codec.obtainFuncCode() != funcCode {
		return 0, 0, errors.New("funcCode mismatch")
	}
	if len(m.codec.obtainData()) != 4 {
		return 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.codec.obtainData()[:2])
	number = binary.BigEndian.Uint16(m.codec.obtainData()[2:4])
	return addr, number, nil
}

func (m *modbusDataParser) data4ParseResponse(funcCode byte) (length byte, data []byte, err error) {
	if m.codec.obtainFuncCode() != funcCode {
		return 0, nil, errors.New("funcCode mismatch")
	}
	if m.codec.obtainData() == nil || len(m.codec.obtainData()) == 0 {
		return 0, nil, errors.New("invalid data length")
	}
	return m.codec.obtainData()[0], m.codec.obtainData()[1:], nil
}

func (m *modbusDataParser) parseCoilsResponse(bytes []byte, number uint16) (length uint16, result []bool, err error) {
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
	result = make([]bool, number)
	for index := uint16(0); index < number; index++ {
		result[index] = strResult[index] == "1"
	}
	return uint16(len(result)), result, nil
}

// 解析读线圈的请求
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadCoilsRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadCoils)
}

// parseReadCoilsResponse 解析读线圈的响应
// data 数据
func (m *modbusDataParser) parseReadCoilsResponse(number uint16) (length uint16, result []bool, err error) {
	_, data, err := m.data4ParseResponse(ReadCoils)
	if err != nil {
		return 0, nil, err
	}
	return m.parseCoilsResponse(data, number)
}

// 解析读离散输入寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadDiscreteInputsRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadDiscreteInputs)
}

// 解析读离散输入寄存器的响应
// length 长度
// data 数据
func (m *modbusDataParser) parseReadDiscreteInputsResponse(number uint16) (length uint16, result []bool, err error) {
	_, data, err := m.data4ParseResponse(ReadDiscreteInputs)
	if err != nil {
		return 0, nil, err
	}
	return m.parseCoilsResponse(data, number)
}

// 解析读保持寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadHoldingRegistersRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadHoldingRegisters)
}

// 解析读保持寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadHoldingRegistersResponse() (length byte, data []uint16, err error) {
	return m.parseReadHoldingRegistersResp(ReadHoldingRegisters)
}

func (m *modbusDataParser) parseReadHoldingRegistersResp(funcCode byte) (length byte, data []uint16, err error) {
	if m.codec.obtainFuncCode() != funcCode {
		return 0, nil, errors.New("funcCode mismatch")
	}
	if m.codec.obtainData() == nil || len(m.codec.obtainData()) == 0 {
		return 0, nil, errors.New("invalid data length")
	}
	length = m.codec.obtainData()[0]
	if length != byte(len(m.codec.obtainData())-1) || length%2 != 0 {
		return 0, nil, errors.New("invalid data length")
	}
	resp := m.codec.obtainData()[1:]
	data = make([]uint16, len(resp)/2)
	for i := 0; i < len(m.codec.obtainData()[1:]); i += 2 {
		data[i/2] = binary.BigEndian.Uint16(resp[i : i+2])
	}
	return length / 2, data, nil
}

// 解析读输入寄存器的请求
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadInputRegistersRequest() (addr, number uint16, err error) {
	return m.data4ParseRequest(ReadInputRegisters)
}

// 解析读输入寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseReadInputRegistersResponse() (length byte, data []uint16, err error) {
	return m.parseReadHoldingRegistersResp(ReadInputRegisters)
}

// 解析写单个线圈的请求或响应
// addr 起始地址
// status 状态
func (m *modbusDataParser) parseWriteSingleCoil() (addr, status uint16, err error) {
	return m.data4ParseRequest(WriteSingleCoil)
}

// 解析写单个保持寄存器的请求或响应
// addr 起始地址
// status 状态
func (m *modbusDataParser) parseWriteSingleRegister() (addr, status uint16, err error) {
	return m.data4ParseRequest(WriteSingleRegister)
}

// 解析写多个线圈的请求
// addr 起始地址
// number 寄存器数量
// length 字节数
// data 数据
func (m *modbusDataParser) parseWriteMultipleCoilsRequest() (addr, number uint16, length byte, data []byte, err error) {
	if m.codec.obtainFuncCode() != WriteMultipleCoils {
		return 0, 0, 0, nil, errors.New("funcCode mismatch")
	}
	if len(m.codec.obtainData()) < 6 {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.codec.obtainData()[:2])
	number = binary.BigEndian.Uint16(m.codec.obtainData()[2:4])
	length = m.codec.obtainData()[4]
	data = m.codec.obtainData()[5:]
	if len(data) != int(length) {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	return
}

// 解析写多个线圈的响应
// addr 起始地址
// number 寄存器数量
// length 字节数
func (m *modbusDataParser) parseWriteMultipleCoilsResponse() (addr, number uint16, err error) {
	if m.codec.obtainFuncCode() != WriteMultipleCoils {
		return 0, 0, errors.New("funcCode mismatch")
	}
	if len(m.codec.obtainData()) < 4 {
		return 0, 0, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.codec.obtainData()[:2])
	number = binary.BigEndian.Uint16(m.codec.obtainData()[2:4])
	return
}

// 解析写多个保持寄存器的请求
// addr 起始地址
// number 寄存器数量
// length 字节数
// data 数据
func (m *modbusDataParser) parseWriteMultipleRegistersRequest() (addr, number uint16, length byte, data []uint16, err error) {
	if m.codec.obtainFuncCode() != WriteMultipleRegisters {
		return 0, 0, 0, nil, errors.New("funcCode mismatch")
	}
	if len(m.codec.obtainData()) < 5 {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	addr = binary.BigEndian.Uint16(m.codec.obtainData()[:2])
	number = binary.BigEndian.Uint16(m.codec.obtainData()[2:4])
	length = m.codec.obtainData()[4]
	resp := m.codec.obtainData()[5:]
	if length != byte(len(resp)) {
		return 0, 0, 0, nil, errors.New("invalid data length")
	}
	data = make([]uint16, len(resp)/2)
	for i := 0; i < len(data); i += 2 {
		data[i/2] = binary.BigEndian.Uint16(resp[i : i+2])
	}
	return addr, number, length, data, nil
}

// 解析写多个保持寄存器的响应
// addr 起始地址
// number 寄存器数量
func (m *modbusDataParser) parseWriteMultipleRegistersResponse() (addr, number uint16, err error) {
	return m.data4ParseRequest(WriteMultipleRegisters)
}

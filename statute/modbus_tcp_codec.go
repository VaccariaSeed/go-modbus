package statute

import (
	"bufio"
	"encoding/binary"
	"errors"
	"slices"
	"sync"
)

// NewModbusTCPCodec 生成一个modbusTCP的编码器
func NewModbusTCPCodec() *ModbusTCPCodec {
	return &ModbusTCPCodec{modbusFrameBuilder: &modbusFrameBuilder{}, intermediary: &intermediary{}}
}

var _ ModbusCodec = (*ModbusTCPCodec)(nil)

// ModbusTCPCodec modbusTCP的编码器
type ModbusTCPCodec struct {
	*intermediary //发送快照
	*modbusFrameBuilder
	identifierNumber uint16
	identLock        sync.Mutex
}

// 生成一个唯一标志
func (m *ModbusTCPCodec) identifier() uint16 {
	m.identLock.Lock()
	defer func() {
		m.identifierNumber++
		m.identLock.Unlock()
	}()
	if m.identifierNumber >= 65535 {
		m.identifierNumber = 1
	}
	return m.identifierNumber
}

// 生成一条完整的报文
func (m *ModbusTCPCodec) buildFrame(slaveId byte, funcCode byte, data []byte) []byte {
	frameId := m.identifier()
	encode := []byte{byte(frameId >> 8), byte(frameId), 0x00, 0x00}
	data1 := append([]byte{slaveId, funcCode}, data...)
	encode = append(encode, byte(len(data1)>>8), byte(len(data1)))
	//保存快照
	m.ident, m.funcCode, m.slaveId = frameId, funcCode, slaveId
	return append(encode, data1...)
}

// BuildReadCoils 读线圈
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusTCPCodec) BuildReadCoils(slaveId byte, addr, number uint16) []byte {
	data := m.buildReadCoilsRequest(addr, number)
	return m.buildFrame(slaveId, ReadCoils, data)
}

// BuildReadDiscreteInputs 读离散输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusTCPCodec) BuildReadDiscreteInputs(slaveId byte, address, number uint16) []byte {
	data := m.buildReadDiscreteInputsRequest(address, number)
	return m.buildFrame(slaveId, ReadDiscreteInputs, data)
}

// BuildReadHoldingRegisters 读保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusTCPCodec) BuildReadHoldingRegisters(slaveId byte, address, number uint16) []byte {
	data := m.buildReadHoldingInputsRequest(address, number)
	return m.buildFrame(slaveId, ReadHoldingRegisters, data)
}

// BuildReadInputInputs 读输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusTCPCodec) BuildReadInputRegisters(slaveId byte, address, number uint16) []byte {
	data := m.buildReadInputRegistersRequest(address, number)
	return m.buildFrame(slaveId, ReadInputRegisters, data)
}

// BuildWriteSingleCoil 写单个线圈
// slaveId 从站id
// addr 起始地址
// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
func (m *ModbusTCPCodec) BuildWriteSingleCoil(slaveId byte, address uint16, value CoilStatus) []byte {
	data := m.buildWriteSingleCoil(address, value)
	return m.buildFrame(slaveId, WriteSingleCoil, data)
}

// BuildWriteSingleRegister 写单个保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// value 设定值
func (m *ModbusTCPCodec) BuildWriteSingleRegister(slaveId byte, address uint16, value uint16) []byte {
	data := m.buildWriteSingleRegister(address, value)
	return m.buildFrame(slaveId, WriteSingleRegister, data)
}

// BuildWriteMultipleCoils 多个线圈的请求
// slaveId 从站id
// addr 寄存器起始地址
// status 线圈状态
func (m *ModbusTCPCodec) BuildWriteMultipleCoils(slaveId byte, address uint16, status ...CoilStatus) ([]byte, error) {
	if status == nil || len(status) == 0 {
		return nil, errors.New("status can not be nil")
	}
	data := m.buildWriteMultipleCoilsRequest(address, uint16(len(status)), status...)
	return m.buildFrame(slaveId, WriteMultipleCoils, data), nil
}

// BuildWriteMultipleRegisters 写多个保持寄存器
func (m *ModbusTCPCodec) BuildWriteMultipleRegisters(slaveId byte, address uint16, value ...uint16) ([]byte, error) {
	if value == nil || len(value) == 0 {
		return nil, errors.New("value can not be nil")
	}
	data, _ := m.buildWriteMultipleRegistersRequest(address, uint16(len(value)), value...)
	return m.buildFrame(slaveId, WriteMultipleRegisters, data), nil
}

// Decode 解码
// result 结果数据集
// error 解码错误
func (m *ModbusTCPCodec) Decode(buf *bufio.Reader) ([]byte, error) {
	//判断frameId
	var frameId uint16
	if err := binary.Read(buf, binary.BigEndian, &frameId); err != nil {
		return nil, err
	}
	if frameId != m.ident {
		return nil, errors.New("invalid identifier")
	}
	//读取两个参数，判定是否为modbusTCP协议
	var flag [2]byte
	if err := binary.Read(buf, binary.BigEndian, &flag); err != nil {
		return nil, err
	}
	if flag[0] != 0 && flag[1] != 0 {
		return nil, errors.New("invalid modbus tcp type")
	}
	//获取长度
	var length uint16
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	length -= 2
	if length < 0 {
		return nil, errors.New("invalid length")
	}
	var data [2]byte
	if err := binary.Read(buf, binary.BigEndian, &data); err != nil {
		return nil, err
	}
	if data[0] != m.slaveId {
		return nil, errors.New("invaild slave id")
	}
	if data[1] != m.funcCode {
		if slices.Contains(errFuncCodes, m.funcCode) {
			return nil, newReturnedAbnormalFuncCode(data[1])
		}
		return nil, errors.New("invaild function code")
	}
	if length > 0 {
		result := make([]byte, length)
		if err := binary.Read(buf, binary.BigEndian, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, nil
}

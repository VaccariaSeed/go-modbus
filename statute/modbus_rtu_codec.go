package statute

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
)

// NewModbusRTUCodec 生成一个modbuRTU的编码器
func NewModbusRTUCodec() *ModbusRTUCodec {
	return &ModbusRTUCodec{modbusFrameBuilder: &modbusFrameBuilder{}, intermediary: &intermediary{}}
}

var _ ModbusCodec = (*ModbusRTUCodec)(nil)

// ModbusRTUCodec modbusRTU的编码器
type ModbusRTUCodec struct {
	*intermediary //发送快照
	*modbusFrameBuilder
}

func (m *ModbusRTUCodec) cs(frame []byte) []byte {
	crc := uint16(0xFFFF)
	for _, b := range frame {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc = crc >> 1
			}
		}
	}
	return []byte{byte(crc & 0xFF), byte(crc >> 8)}
}

// 生成一条完整的报文
func (m *ModbusRTUCodec) buildFrame(slaveId byte, funcCode byte, data []byte) []byte {
	m.slaveId, m.funcCode = slaveId, funcCode
	encode := []byte{slaveId, funcCode}
	encode = append(encode, data...)
	cs := m.cs(encode)
	return append(encode, cs...)
}

// BuildReadCoils 读线圈
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusRTUCodec) BuildReadCoils(slaveId byte, addr, number uint16) []byte {
	data := m.buildReadCoilsRequest(addr, number)
	return m.buildFrame(slaveId, ReadCoils, data)
}

// BuildReadDiscreteInputs 读离散输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusRTUCodec) BuildReadDiscreteInputs(slaveId byte, address, number uint16) []byte {
	data := m.buildReadDiscreteInputsRequest(address, number)
	return m.buildFrame(slaveId, ReadDiscreteInputs, data)
}

// BuildReadHoldingRegisters 读保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusRTUCodec) BuildReadHoldingRegisters(slaveId byte, address, number uint16) []byte {
	data := m.buildReadHoldingInputsRequest(address, number)
	return m.buildFrame(slaveId, ReadHoldingRegisters, data)
}

// BuildReadInputRegisters 读输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (m *ModbusRTUCodec) BuildReadInputRegisters(slaveId byte, address, number uint16) []byte {
	data := m.buildReadInputRegistersRequest(address, number)
	return m.buildFrame(slaveId, ReadInputRegisters, data)
}

// BuildWriteSingleCoil 写单个线圈
// slaveId 从站id
// addr 起始地址
// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
func (m *ModbusRTUCodec) BuildWriteSingleCoil(slaveId byte, address uint16, value CoilStatus) []byte {
	data := m.buildWriteSingleCoil(address, value)
	return m.buildFrame(slaveId, WriteSingleCoil, data)
}

// BuildWriteSingleRegister 写单个保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// value 设定值
func (m *ModbusRTUCodec) BuildWriteSingleRegister(slaveId byte, address uint16, value uint16) []byte {
	data := m.buildWriteSingleRegister(address, value)
	return m.buildFrame(slaveId, WriteSingleRegister, data)
}

// BuildWriteMultipleCoils 多个线圈的请求
// slaveId 从站id
// addr 寄存器起始地址
// status 线圈状态
func (m *ModbusRTUCodec) BuildWriteMultipleCoils(slaveId byte, address uint16, status ...CoilStatus) ([]byte, error) {
	if status == nil || len(status) == 0 {
		return nil, errors.New("status can not be nil")
	}
	data := m.buildWriteMultipleCoilsRequest(address, uint16(len(status)), status...)
	return m.buildFrame(slaveId, WriteMultipleCoils, data), nil
}

// BuildWriteMultipleRegisters 写多个保持寄存器
func (m *ModbusRTUCodec) BuildWriteMultipleRegisters(slaveId byte, address uint16, value ...uint16) ([]byte, error) {
	if value == nil || len(value) == 0 {
		return nil, errors.New("value can not be nil")
	}
	data, _ := m.buildWriteMultipleRegistersRequest(address, uint16(len(value)), value...)
	return m.buildFrame(slaveId, WriteMultipleRegisters, data), nil
}

// Decode 解码
// result 结果数据集
// error 解码错误
func (m *ModbusRTUCodec) Decode(buf *bufio.Reader) ([]byte, error) {
	var slaveId byte
	if err := binary.Read(buf, binary.BigEndian, &slaveId); err != nil {
		return nil, err
	}
	if slaveId != m.slaveId {
		return nil, errors.New("invaild slave id")
	}
	var funcCode byte
	if err := binary.Read(buf, binary.BigEndian, &funcCode); err != nil {
		return nil, err
	}
	if funcCode != m.funcCode {
		if slices.Contains(errFuncCodes, m.funcCode) {
			return nil, newReturnedAbnormalFuncCode(funcCode)
		}
		return nil, errors.New("invaild function code")
	}
	var data []byte
	var result []byte
	if slices.Contains(mrFuncCodes, m.funcCode) {
		if m.funcCode == ReadCoils || m.funcCode == ReadDiscreteInputs || m.funcCode == ReadHoldingRegisters || m.funcCode == ReadInputRegisters {
			//读线圈
			var length byte
			if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
				return nil, err
			}
			if length < 0 {
				return nil, errors.New("invalid length")
			}
			result = make([]byte, length)
			if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
				return nil, err
			}
			data = append([]byte{length}, result...)
		} else if m.funcCode == WriteSingleCoil || m.funcCode == WriteSingleRegister || m.funcCode == WriteMultipleRegisters {
			result = make([]byte, 4)
			if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
				return nil, err
			}
			data = result
		} else {
			result = make([]byte, 4)
			if err := binary.Read(buf, binary.LittleEndian, &result); err != nil {
				return nil, err
			}
			data = result
		}
		if err := m.checkCs(data, buf); err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, fmt.Errorf("error function code:%d", m.funcCode)

}

func (m *ModbusRTUCodec) checkCs(result []byte, buf *bufio.Reader) error {
	data := append([]byte{m.slaveId, m.funcCode}, result...)
	cs := make([]byte, 2)
	if err := binary.Read(buf, binary.BigEndian, &cs); err != nil {
		return err
	}
	checkCs := m.cs(data)
	if checkCs[0] != cs[0] || checkCs[1] != cs[1] {
		return errors.New("cs error")
	}
	return nil
}

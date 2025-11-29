package go_modbus

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"slices"
)

var _ ModbusStatuter = (*ModbusRTUStatute)(nil)

type ModbusRTUStatute struct {
	original []byte //原始报文
	slaveId  byte   //从站地址
	funcCode byte   //功能码
	data     []byte //数据域
}

func (m *ModbusRTUStatute) baseDecode(buf *bufio.Reader) error {
	err := binary.Read(buf, binary.LittleEndian, &m.slaveId)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &m.funcCode)
	if err != nil {
		return err
	}
	if !slices.Contains(mrFuncCodes, m.funcCode) {
		return err
	}
	m.original = []byte{m.slaveId, m.funcCode}
	m.data = nil
	return nil
}

// DecodeMasterFrame 解码主站发来的报文
func (m *ModbusRTUStatute) DecodeMasterFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.DecodeMasterReader(buf)
}

// DecodeMasterReader 解码主站发来的报文
func (m *ModbusRTUStatute) DecodeMasterReader(buf *bufio.Reader) error {
	err := m.baseDecode(buf)
	if err != nil {
		return err
	}
	if m.funcCode == ReadCoils || m.funcCode == ReadDiscreteInputs || m.funcCode == ReadHoldingRegisters || m.funcCode == ReadInputRegisters || m.funcCode == WriteSingleCoil || m.funcCode == WriteSingleRegister {
		//读线圈，读离散输入寄存器.读保持寄存器,读输入寄存器
		m.data = make([]byte, 4)
		err = binary.Read(buf, binary.LittleEndian, &m.data)
		if err != nil {
			return err
		}
	} else {
		startAddrAndSize := make([]byte, 4)
		err = binary.Read(buf, binary.LittleEndian, &startAddrAndSize)
		if err != nil {
			return err
		}
		var length byte
		err = binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return err
		}
		data := make([]byte, length)
		err = binary.Read(buf, binary.LittleEndian, &data)
		if err != nil {
			return err
		}
		m.data = append(startAddrAndSize, length)
		m.data = append(m.data, data...)
	}
	return m.checkCs(buf)
}

// DecodeSlaveFrame 解码从站响应的报文
func (m *ModbusRTUStatute) DecodeSlaveFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.DecodeSlaveReader(buf)
}

// DecodeSlaveReader 解码从站响应的报文
func (m *ModbusRTUStatute) DecodeSlaveReader(buf *bufio.Reader) error {
	err := m.baseDecode(buf)
	if err != nil {
		return err
	}
	if m.funcCode == ReadCoils || m.funcCode == ReadDiscreteInputs || m.funcCode == ReadHoldingRegisters || m.funcCode == ReadInputRegisters {
		//读线圈
		var length byte
		err = binary.Read(buf, binary.LittleEndian, &length)
		if err != nil {
			return err
		}
		data := make([]byte, length)
		err = binary.Read(buf, binary.LittleEndian, &data)
		if err != nil {
			return err
		}
		m.data = append([]byte{length}, data...)
	} else if m.funcCode == WriteSingleCoil || m.funcCode == WriteSingleRegister || m.funcCode == WriteMultipleRegisters {
		m.data = make([]byte, 4)
		err = binary.Read(buf, binary.LittleEndian, &m.data)
		if err != nil {
			return err
		}
	} else {
		m.data = make([]byte, 5)
		err = binary.Read(buf, binary.LittleEndian, &m.data)
		if err != nil {
			return err
		}
	}
	return m.checkCs(buf)
}

func (m *ModbusRTUStatute) checkCs(buf *bufio.Reader) error {
	m.original = append(m.original, m.data...)
	//计算cs
	cs := make([]byte, 2)
	err := binary.Read(buf, binary.BigEndian, &cs)
	if err != nil {
		return err
	}
	checkCs := m.cs(m.original)
	if checkCs[0] != cs[0] || checkCs[1] != cs[1] {
		return errors.New("cs error")
	}
	m.original = append(m.original, cs...)
	return nil
}

// Data 获取数据域
func (m *ModbusRTUStatute) Data() []byte {
	return m.data
}

// Encode 编码
func (m *ModbusRTUStatute) Encode() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m *ModbusRTUStatute) cs(frame []byte) []byte {
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

// OriginalFrame 获取原始报文
func (m *ModbusRTUStatute) OriginalFrame() []byte {
	return m.original
}

func (m *ModbusRTUStatute) Copy() ModbusStatuter {
	return &ModbusRTUStatute{
		original: m.original,
		slaveId:  m.slaveId,
		funcCode: m.funcCode,
		data:     m.data,
	}
}

func (m *ModbusRTUStatute) SlaveId() byte {
	return m.slaveId
}

func (m *ModbusRTUStatute) FuncCode() byte {
	return m.funcCode
}

func (m *ModbusRTUStatute) DataParser() *ModbusDataParser {
	return &ModbusDataParser{frame: m, statute: modbusRtu}
}

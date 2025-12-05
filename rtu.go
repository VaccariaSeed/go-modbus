package go_modbus

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"slices"
)

var _ modbusStatute = (*modbusRTUStatute)(nil)

type modbusRTUStatute struct {
	original []byte //原始报文
	slaveId  byte   //从站地址
	funcCode byte   //功能码
	data     []byte //数据域
}

func (m *modbusRTUStatute) baseDecode(buf *bufio.Reader) error {
	err := binary.Read(buf, binary.LittleEndian, &m.slaveId)
	if err != nil {
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &m.funcCode)
	if err != nil {
		return err
	}
	if !slices.Contains(mrFuncCodes, m.funcCode) {
		return FuncCodeError.setFuncCode(m.funcCode)
	}
	m.original = []byte{m.slaveId, m.funcCode}
	m.data = nil
	return nil
}

// DecodeMasterFrame 解码主站发来的报文
func (m *modbusRTUStatute) decodeMasterFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.decodeMasterReader(buf)
}

// DecodeMasterReader 解码主站发来的报文
func (m *modbusRTUStatute) decodeMasterReader(buf *bufio.Reader) error {
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
func (m *modbusRTUStatute) decodeSlaveFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.decodeSlaveReader(buf)
}

// DecodeSlaveReader 解码从站响应的报文
func (m *modbusRTUStatute) decodeSlaveReader(buf *bufio.Reader) error {
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
		m.data = make([]byte, 4)
		err = binary.Read(buf, binary.LittleEndian, &m.data)
		if err != nil {
			return err
		}
	}
	return m.checkCs(buf)
}

func (m *modbusRTUStatute) checkCs(buf *bufio.Reader) error {
	m.original = append(m.original, m.data...)
	//计算cs
	cs := make([]byte, 2)
	err := binary.Read(buf, binary.BigEndian, &cs)
	if err != nil {
		return err
	}
	checkCs := m.cs(m.original)
	if checkCs[0] != cs[0] || checkCs[1] != cs[1] {
		return CsError
	}
	m.original = append(m.original, cs...)
	return nil
}

// Encode 编码
func (m *modbusRTUStatute) encode(_ uint16, slaveId, functionCode byte, data []byte) ([]byte, error) {
	encode := []byte{slaveId, functionCode}
	encode = append(encode, data...)
	cs := m.cs(encode)
	return append(encode, cs...), nil
}

func (m *modbusRTUStatute) cs(frame []byte) []byte {
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

func (m *modbusRTUStatute) obtainFuncCode() byte {
	return m.funcCode
}

func (m *modbusRTUStatute) obtainData() []byte {
	return m.data
}

func (m *modbusRTUStatute) identifier() uint16 {
	return 0
}

func (m *modbusRTUStatute) obtainSlaveId() byte {
	return m.slaveId
}

func (m *modbusRTUStatute) obtainFrame() []byte {
	return m.original
}

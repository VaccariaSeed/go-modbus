package go_modbus

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"slices"
)

var _ modbusStatute = (*modbusTCPStatute)(nil)

type modbusTCPStatute struct {
	frameId  uint16
	original []byte //原始报文
	slaveId  byte   //从站地址
	funcCode byte   //功能码
	data     []byte //数据域
}

func (m *modbusTCPStatute) baseDecode(buf *bufio.Reader) error {
	peeked, err := buf.Peek(8)
	if err != nil {
		return err
	}
	m.frameId = binary.BigEndian.Uint16([]byte{peeked[0], peeked[1]})
	//获取Modbus TCP协议协议
	if peeked[2] != 0 && peeked[3] != 0 {
		_, _ = buf.ReadByte()
		return ModbusTCPProtocolFlagError
	}
	//获取单元标识符
	m.slaveId = peeked[6]
	//获取功能码
	m.funcCode = peeked[7]
	if !slices.Contains(mrFuncCodes, m.funcCode) {
		_, _ = buf.ReadByte()
		return FuncCodeError.setFuncCode(m.funcCode)
	}
	_, err = buf.Discard(8)
	if err != nil {
		return err
	}
	//获取所有的长度
	length := binary.BigEndian.Uint16([]byte{peeked[4], peeked[5]}) - 2
	m.data = make([]byte, length)
	err = binary.Read(buf, binary.LittleEndian, &m.data)
	if err != nil {
		return err
	}
	m.original = append(peeked, m.data...)
	return err
}

func (m *modbusTCPStatute) decodeMasterFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.decodeMasterReader(buf)
}

func (m *modbusTCPStatute) decodeMasterReader(buf *bufio.Reader) error {
	return m.baseDecode(buf)
}

func (m *modbusTCPStatute) decodeSlaveFrame(frame []byte) error {
	buf := bufio.NewReader(bytes.NewReader(frame))
	return m.decodeSlaveReader(buf)
}

func (m *modbusTCPStatute) decodeSlaveReader(buf *bufio.Reader) error {
	return m.baseDecode(buf)
}

func (m *modbusTCPStatute) encode(frameId uint16, slaveId, functionCode byte, data []byte) ([]byte, error) {
	encode := []byte{byte(frameId >> 8), byte(frameId), 0x00, 0x00}
	data1 := append([]byte{slaveId, functionCode}, data...)
	encode = append(encode, byte(len(data1)>>8), byte(len(data1)))
	return append(encode, data1...), nil
}

func (m *modbusTCPStatute) obtainFuncCode() byte {
	return m.funcCode
}

func (m *modbusTCPStatute) obtainData() []byte {
	return m.data
}

func (m *modbusTCPStatute) identifier() uint16 {
	return m.frameId
}

func (m *modbusTCPStatute) obtainSlaveId() byte {
	return m.slaveId
}

func (m *modbusTCPStatute) obtainFrame() []byte {
	return m.original
}

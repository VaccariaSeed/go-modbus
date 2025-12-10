package go_modbus

import (
	"errors"
	"sync"
	"time"
)

const (
	defaultRwTimeout      = 20 * time.Millisecond
	defaultConnectTimeout = 3 * time.Second
	defaultReadTimeout    = 2 * time.Second
	defaultWriteTimeout   = 2 * time.Second
)

// ModbusPacket 编解码器
type ModbusPacket struct {
	*modbusFrameBuilder //报文构造器
	*modbusDataParser
}

func (m *ModbusPacket) isModbusTCP() (bool, error) {
	if m.modbusFrameBuilder == nil {
		return false, errors.New("modbus packet is nil")
	}
	if m.modbusDataParser == nil || m.codec == nil {
		return false, errors.New("modbus packet codec is nil")
	}
	if _, ok := m.codec.(*modbusTCPStatute); ok {
		return true, nil
	}
	return false, nil
}

func NewModbusRTUPacket() *ModbusPacket {
	return &ModbusPacket{
		modbusFrameBuilder: &modbusFrameBuilder{},
		modbusDataParser:   &modbusDataParser{codec: &modbusRTUStatute{}},
	}
}

func NewModbusTCPPacket() *ModbusPacket {
	return &ModbusPacket{
		modbusFrameBuilder: &modbusFrameBuilder{},
		modbusDataParser:   &modbusDataParser{codec: &modbusTCPStatute{}},
	}
}

// NewModbusClient 创建一个modbus客户端
func NewModbusClient(packet *ModbusPacket, conn Connector, rwInterval time.Duration) (*ModbusClient, error) {
	if rwInterval <= 0 {
		rwInterval = defaultRwTimeout
	}
	if packet == nil || conn == nil {
		return nil, errors.New("invalid parameter")
	}
	isModbusTcp, err := packet.isModbusTCP()
	if err != nil {
		return nil, err
	}
	return &ModbusClient{ModbusPacket: packet, conn: conn, identifierNumber: 1, rwInterval: rwInterval, isModbusTCP: isModbusTcp}, nil
}

// ModbusClient modbus 客户端
type ModbusClient struct {
	*ModbusPacket
	conn Connector

	rwInterval       time.Duration //读写间隔
	isModbusTCP      bool
	identifierNumber uint16
	identLock        sync.Mutex

	SendPrintHandler    func(frame []byte) //发送调用
	ReceivePrintHandler func(frame []byte) //接收调用
}

// Flush 刷新通道数据
func (c *ModbusClient) Flush() error {
	if c.conn == nil {
		return errors.New("connector is nil")
	}
	return c.conn.Flush()
}

func (c *ModbusClient) identifier() uint16 {
	if !c.isModbusTCP {
		return 0
	}
	c.identLock.Lock()
	defer func() {
		c.identifierNumber++
		c.identLock.Unlock()
	}()
	if c.identifierNumber >= 65535 {
		c.identifierNumber = 1
	}
	return c.identifierNumber
}

// Connect 启动连接
func (c *ModbusClient) Connect() error {
	c.identifierNumber = 1
	if err := c.conn.Connect(); err != nil {
		return err
	}
	return nil
}

func (c *ModbusClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ModbusClient) send(slaveId byte, funcCode byte, data []byte) error {
	ident := c.identifier()
	req, _ := c.codec.encode(ident, slaveId, funcCode, data)
	_, err := c.conn.Write(req)
	if err != nil {
		return err
	}
	if c.SendPrintHandler != nil {
		c.SendPrintHandler(req)
	}
	time.Sleep(c.rwInterval)
	err = c.conn.Read(c.codec, ident, slaveId, funcCode)
	if c.ReceivePrintHandler != nil {
		c.ReceivePrintHandler(c.codec.obtainFrame())
	}
	return err
}

// ReadCoils 读线圈
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (c *ModbusClient) ReadCoils(slaveId byte, addr, number uint16) (length uint16, result []bool, err error) {
	data := c.buildReadCoilsRequest(addr, number)
	err = c.send(slaveId, ReadCoils, data)
	if err != nil {
		return
	}
	return c.parseReadCoilsResponse(number)
}

// ReadDiscreteInputs 读离散输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (c *ModbusClient) ReadDiscreteInputs(slaveId byte, address, number uint16) (length uint16, result []bool, err error) {
	data := c.buildReadDiscreteInputsRequest(address, number)
	err = c.send(slaveId, ReadDiscreteInputs, data)
	if err != nil {
		return
	}
	return c.parseReadDiscreteInputsResponse(number)
}

// ReadHoldingRegisters 读保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (c *ModbusClient) ReadHoldingRegisters(slaveId byte, address, number uint16) (length byte, result []uint16, err error) {
	data := c.buildReadHoldingInputsRequest(address, number)
	err = c.send(slaveId, ReadHoldingRegisters, data)
	if err != nil {
		return
	}
	return c.parseReadHoldingRegistersResponse()
}

// ReadInputInputs 读输入寄存器
// slaveId 从站id
// addr 寄存器起始地址
// number 寄存器数量
func (c *ModbusClient) ReadInputInputs(slaveId byte, address, number uint16) (length byte, result []uint16, err error) {
	data := c.buildReadInputInputsRequest(address, number)
	err = c.send(slaveId, ReadInputRegisters, data)
	if err != nil {
		return
	}
	return c.parseReadInputRegistersResponse()
}

// WriteSingleCoil 写单个线圈
// slaveId 从站id
// addr 起始地址
// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
func (c *ModbusClient) WriteSingleCoil(slaveId byte, address uint16, value bool) (addr uint16, status bool, err error) {
	data := c.buildWriteSingleCoil(address, value)
	err = c.send(slaveId, WriteSingleCoil, data)
	if err != nil {
		return
	}
	addr, statusValue, err := c.parseWriteSingleCoil()
	if err != nil {
		return
	}
	if statusValue == 0xFF00 {
		status = true
	} else if statusValue == 0x0000 {
		status = false
	} else {
		return 0, false, errors.New("invalid status")
	}
	return addr, status, nil
}

// WriteSingleRegister 写单个保持寄存器
// slaveId 从站id
// addr 寄存器起始地址
// value 设定值
func (c *ModbusClient) WriteSingleRegister(slaveId byte, address uint16, value uint16) (addr, status uint16, err error) {
	data := c.buildWriteSingleRegister(address, value)
	err = c.send(slaveId, WriteSingleRegister, data)
	if err != nil {
		return
	}
	return c.parseWriteSingleRegister()
}

// WriteMultipleCoils 多个线圈的请求
// slaveId 从站id
// addr 寄存器起始地址
// status 线圈状态
func (c *ModbusClient) WriteMultipleCoils(slaveId byte, address uint16, status ...bool) (addr, size uint16, err error) {
	if status == nil || len(status) == 0 {
		return 0, 0, errors.New("status can not be nil")
	}
	data := c.buildWriteMultipleCoilsRequest(address, uint16(len(status)), status...)
	err = c.send(slaveId, WriteMultipleCoils, data)
	if err != nil {
		return
	}
	return c.parseWriteMultipleCoilsResponse()
}

// WriteMultipleRegisters 写多个保持寄存器
func (c *ModbusClient) WriteMultipleRegisters(slaveId byte, address uint16, value ...uint16) (addr, size uint16, err error) {
	if value == nil || len(value) == 0 {
		return 0, 0, errors.New("value can not be nil")
	}
	data, _ := c.buildWriteMultipleRegistersRequest(address, uint16(len(value)), value...)
	err = c.send(slaveId, WriteMultipleRegisters, data)
	if err != nil {
		return
	}
	return c.parseWriteMultipleRegistersResponse()
}

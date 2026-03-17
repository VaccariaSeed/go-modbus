package statute

import (
	"bufio"
)

type ModbusCodec interface {

	// 生成一条完整的报文
	buildFrame(slaveId byte, funcCode byte, data []byte) []byte

	// BuildReadCoils 读线圈
	// slaveId 从站id
	// addr 寄存器起始地址
	// number 寄存器数量
	BuildReadCoils(slaveId byte, addr, number uint16) []byte

	// BuildReadDiscreteInputs 读离散输入寄存器
	// slaveId 从站id
	// addr 寄存器起始地址
	// number 寄存器数量
	BuildReadDiscreteInputs(slaveId byte, address, number uint16) []byte

	// BuildReadHoldingRegisters 读保持寄存器
	// slaveId 从站id
	// addr 寄存器起始地址
	// number 寄存器数量
	BuildReadHoldingRegisters(slaveId byte, address, number uint16) []byte

	// BuildReadInputRegisters 读输入寄存器
	// slaveId 从站id
	// addr 寄存器起始地址
	// number 寄存器数量
	BuildReadInputRegisters(slaveId byte, address, number uint16) []byte

	// BuildWriteSingleCoil 写单个线圈
	// slaveId 从站id
	// addr 起始地址
	// value 设定值 写0xFF00表示线圈为ON，写0x0000表示线圈为OFF
	BuildWriteSingleCoil(slaveId byte, address uint16, value CoilStatus) []byte

	// BuildWriteSingleRegister 写单个保持寄存器
	// slaveId 从站id
	// addr 寄存器起始地址
	// value 设定值
	BuildWriteSingleRegister(slaveId byte, address uint16, value uint16) []byte

	// BuildWriteMultipleCoils 多个线圈的请求
	// slaveId 从站id
	// addr 寄存器起始地址
	// status 线圈状态
	BuildWriteMultipleCoils(slaveId byte, address uint16, status ...CoilStatus) ([]byte, error)

	// BuildWriteMultipleRegisters 写多个保持寄存器
	BuildWriteMultipleRegisters(slaveId byte, address uint16, value ...uint16) ([]byte, error)

	// Decode 解码
	// result 结果数据集
	// error 解码错误
	Decode(buf *bufio.Reader) ([]byte, error)

	// ObtainIntermediary 获取中介
	ObtainIntermediary() *intermediary
}

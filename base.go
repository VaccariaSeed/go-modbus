package go_modbus

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

var mrFuncCodes []byte

var errFuncCodes []byte

func init() {
	mrFuncCodes = []byte{ReadCoils, ReadDiscreteInputs, ReadHoldingRegisters, ReadInputRegisters, WriteSingleCoil, WriteSingleRegister, WriteMultipleCoils, WriteMultipleRegisters}

	errFuncCodes = []byte{ReadCoils + 0x80, ReadDiscreteInputs + 0x80, ReadHoldingRegisters + 0x80, ReadInputRegisters + 0x80, WriteSingleCoil + 0x80, WriteSingleRegister + 0x80, WriteMultipleCoils + 0x80, WriteMultipleRegisters + 0x80}
}

const (
	ReadCoils              byte = 0x01 //读线圈,位,取得一组逻辑线圈的当前状态(ON/OFF)
	ReadDiscreteInputs     byte = 0x02 //读离散输入寄存器,位,取得一组开关输入的当前状态(ON/OFF)
	ReadHoldingRegisters   byte = 0x03 //读保持寄存器,整型、浮点型、字符型,在一个或多个保持寄存器中取得当前的二进制值
	ReadInputRegisters     byte = 0x04 //读输入寄存器,整型、浮点型,在一个或多个输入寄存器中取得当前的二进制值
	WriteSingleCoil        byte = 0x05 //写单个线圈寄存器,位,强置一个逻辑线圈的通断状态
	WriteSingleRegister    byte = 0x06 //写单个保持寄存器,整型、浮点型、字符型,把具体二进制值装入一个保持寄存器
	WriteMultipleCoils     byte = 0x0F //写多个线圈寄存器,位,强置一串连续逻辑线圈的通断
	WriteMultipleRegisters byte = 0x10 //写多个保持寄存器,整型、浮点型、字符型,把具体的二进制值装入一串连续的保持寄存器
)

type modbusStatute interface {
	baseDecode(buf *bufio.Reader) error
	decodeMasterFrame(frame []byte) error       //解码主站发来的报文
	decodeMasterReader(buf *bufio.Reader) error //解码主站发来的报文
	decodeSlaveFrame(frame []byte) error        //解码从站响应的报文
	decodeSlaveReader(buf *bufio.Reader) error  //解码从站响应的报文
	encode(frameId uint16, slaveId, functionCode byte, data []byte) ([]byte, error)
	obtainFuncCode() byte
	obtainData() []byte
	identifier() uint16
	obtainSlaveId() byte
	obtainFrame() []byte
}

var _ error = (*FuncCodeError)(nil)

type FuncCodeError struct {
	funcCode  byte
	errorCode byte
}

func (f *FuncCodeError) Error() string {
	return fmt.Sprintf("Function code 0x%02x, error code 0x%02x", f.funcCode, f.errorCode)
}

func (f *FuncCodeError) ErrorCode() (errFuncCode, errorCode byte) {
	return f.funcCode, f.errorCode
}

func (f *FuncCodeError) setFuncCode(funcCode byte, errorCode byte) *FuncCodeError {
	f.funcCode, f.errorCode = funcCode, errorCode
	return f
}

var fce = &FuncCodeError{}

var CsError = errors.New("cs error")

var ModbusTCPProtocolFlagError = errors.New("modbus tcp protocol flag error, must be [0x00, 0x00]")

var opErr *net.OpError

// CheckTCPErr 校验错误判定失败
func CheckTCPErr(err error) bool {
	if errors.As(err, &opErr) && opErr.Timeout() {
		//TCP通道超时
		return false
	} else if errors.Is(err, io.EOF) || // 对端正常关闭
		errors.Is(err, syscall.ECONNRESET) || // 对端强制重置
		errors.Is(err, syscall.EPIPE) || errors.As(err, &opErr) {
		return true
	} else {
		return false
	}
}

// CheckSerialErr 校验错误判定失败
func CheckSerialErr(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, syscall.EIO) {
		return true
	}
	return false
}

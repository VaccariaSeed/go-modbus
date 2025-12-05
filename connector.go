package go_modbus

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/tarm/serial"
)

var NoConnectionError = errors.New("no connected")
var FuncCodeOrSlaveIdError = errors.New("response funcCode or slaveId error")

type Parity byte

const (
	ParityNone  Parity = 'N'
	ParityOdd   Parity = 'O'
	ParityEven  Parity = 'E'
	ParityMark  Parity = 'M' // parity bit is always 1
	ParitySpace Parity = 'S' // parity bit is always 0
)

type Connector interface {
	Connect() error                                                            //建立连接
	Close() error                                                              //关闭
	Write(frame []byte) (int, error)                                           //写数据
	Read(codec modbusStatute, ident uint16, slaveId byte, funcCode byte) error //读
	Flush() error
}

var _ Connector = (*TCPConnector)(nil)
var _ Connector = (*SerialConnector)(nil)

// NewTCPConnector 创建一个tcp连接器
func NewTCPConnector(ip string, port int, connectTimeout, readTimeout, writeTimeout time.Duration) *TCPConnector {
	if connectTimeout <= 0 {
		connectTimeout = defaultConnectTimeout
	}
	if readTimeout <= 0 {
		readTimeout = defaultReadTimeout
	}
	if writeTimeout <= 0 {
		writeTimeout = defaultWriteTimeout
	}
	return &TCPConnector{ip: ip, port: port, connectTimeout: connectTimeout, readTimeout: readTimeout, writeTimeout: writeTimeout}
}

type TCPConnector struct {
	ip             string
	port           int
	connectTimeout time.Duration //连接超时
	readTimeout    time.Duration //读超时
	writeTimeout   time.Duration //写超时
	conn           *net.TCPConn
	reader         *bufio.Reader
	lock           sync.Mutex
}

func (T *TCPConnector) Connect() error {
	T.lock.Lock()
	defer T.lock.Unlock()
	conn, err := net.DialTimeout("tcp", T.ip+":"+strconv.Itoa(T.port), T.connectTimeout)
	if err != nil {
		return err
	}
	T.conn = conn.(*net.TCPConn)
	T.reader = bufio.NewReader(conn)
	return nil
}

func (T *TCPConnector) Close() error {
	T.lock.Lock()
	defer func() {
		T.reader = nil
		T.conn = nil
		T.lock.Unlock()
	}()
	if T.conn != nil {
		return T.conn.Close()
	}
	return nil
}

func (T *TCPConnector) Write(frame []byte) (int, error) {
	T.lock.Lock()
	defer T.lock.Unlock()
	if T.conn == nil {
		return 0, NoConnectionError
	}
	err := T.conn.SetWriteDeadline(time.Now().Add(T.writeTimeout))
	if err != nil {
		return 0, err
	}
	return T.conn.Write(frame)
}

func (T *TCPConnector) Read(codec modbusStatute, ident uint16, slaveId byte, funcCode byte) error {
	T.lock.Lock()
	defer T.lock.Unlock()
	if T.conn == nil {
		return NoConnectionError
	}
	err := T.conn.SetReadDeadline(time.Now().Add(T.readTimeout))
	if err != nil {
		return err
	}
	err = codec.decodeSlaveReader(T.reader)
	if err != nil {
		return err
	}
	if ident == codec.identifier() && slaveId == codec.obtainSlaveId() && funcCode == codec.obtainFuncCode() {
		return nil
	}
	return FuncCodeOrSlaveIdError
}

func (T *TCPConnector) Flush() error {
	T.lock.Lock()
	defer T.lock.Unlock()
	if T.conn == nil {
		return NoConnectionError
	}
	_, err := io.ReadAll(T.reader)
	return err
}

// NewSerialConnector 创建一个串口连接器
func NewSerialConnector(port string, baud int, dataBit byte, parity Parity, stopBit byte, readTimeout time.Duration) *SerialConnector {
	if readTimeout <= 0 {
		readTimeout = defaultReadTimeout
	}
	conn := &SerialConnector{port: port, baud: baud, dataBit: dataBit, parity: parity, stopBit: stopBit, readTimeout: readTimeout}
	return conn
}

type SerialConnector struct {
	port        string        //串口号
	baud        int           //波特率
	dataBit     byte          //数据位
	parity      Parity        //校验位
	stopBit     byte          //停止位
	readTimeout time.Duration //读超时
	serialPort  *serial.Port

	reader *bufio.Reader

	lock sync.Mutex
}

func (s *SerialConnector) Connect() error {
	config := &serial.Config{Name: s.port, Baud: s.baud, Size: s.dataBit, Parity: serial.Parity(s.parity), StopBits: serial.StopBits(s.stopBit), ReadTimeout: s.readTimeout}
	port, err := serial.OpenPort(config)
	if err != nil {
		return err
	}
	s.serialPort = port
	s.reader = bufio.NewReader(port)
	return nil
}

func (s *SerialConnector) Close() error {
	s.lock.Lock()
	defer func() {
		s.reader = nil
		s.serialPort = nil
		s.lock.Unlock()
	}()
	if s.serialPort != nil {
		return s.serialPort.Close()
	}
	return nil
}

func (s *SerialConnector) Write(frame []byte) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.serialPort == nil {
		return 0, NoConnectionError
	}
	return s.serialPort.Write(frame)
}

func (s *SerialConnector) Read(codec modbusStatute, ident uint16, slaveId byte, funcCode byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.serialPort == nil {
		return NoConnectionError
	}
	err := codec.decodeSlaveReader(s.reader)
	if err != nil {
		return err
	}
	if ident == codec.identifier() && slaveId == codec.obtainSlaveId() && funcCode == codec.obtainFuncCode() {
		return nil
	}
	return FuncCodeOrSlaveIdError
}

func (s *SerialConnector) Flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.serialPort == nil {
		return NoConnectionError
	}
	return s.serialPort.Flush()
}

package go_modbus

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestTCP(t *testing.T) {
	master, err := NewModbusClient(NewModbusTCPPacket(), NewTCPConnector("127.0.0.1", 502, defaultConnectTimeout, defaultReadTimeout, defaultWriteTimeout), defaultRwTimeout)
	if err != nil {
		fmt.Printf("NewModbusClient err: %v \n", err)
		return
	}
	master.SendPrintHandler = func(frame []byte) {
		fmt.Println("发送数据：", hex.EncodeToString(frame))
	}

	master.ReceivePrintHandler = func(frame []byte) {
		fmt.Println("接收数据：", hex.EncodeToString(frame))
	}

	for {
		err = master.Connect()
		if err != nil {
			fmt.Printf("master.Connect err: %v \n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	for {
		fmt.Println("----------------------")
		// 设置随机种子，使用当前时间纳秒
		length, result, readErr := master.ReadHoldingRegisters(1, 11, 9)
		if readErr != nil {
			fmt.Printf("ReadCoils err: %v \n", readErr)
		} else {
			fmt.Printf("ReadCoils length: %v \n", length)
			fmt.Printf("ReadCoils result: %v \n", result)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestSerial(t *testing.T) {
	master, err := NewModbusClient(NewModbusRTUPacket(), NewSerialConnector("COM3", 9600, 8, ParityNone, 1, defaultReadTimeout), defaultRwTimeout)
	if err != nil {
		fmt.Printf("NewModbusClient err: %v \n", err)
		return
	}
	master.SendPrintHandler = func(frame []byte) {
		fmt.Println("发送数据：", hex.EncodeToString(frame))
	}

	master.ReceivePrintHandler = func(frame []byte) {
		fmt.Println("接收数据：", hex.EncodeToString(frame))
	}

	for {
		err = master.Connect()
		if err != nil {
			fmt.Printf("master.Connect err: %v \n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	for {
		_ = master.Flush()
		fmt.Println("----------------------")
		length, result, readErr := master.WriteMultipleRegisters(1, 0, 22, 33, 44, 55, 66, 77)
		if readErr != nil {
			fmt.Printf("ReadCoils err: %v \n", readErr)
		} else {
			fmt.Printf("ReadCoils length: %v \n", length)
			fmt.Printf("ReadCoils result: %v \n", result)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

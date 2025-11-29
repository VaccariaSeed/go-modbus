package go_modbus

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestModbusRtu(t *testing.T) {
	messageChannel := make(chan *ModbusRTUStatute, 2)
	go func() {
		for msg := range messageChannel {
			fmt.Println(hex.EncodeToString(msg.data))
		}
	}()
	md := NewModbusRTUDecoder(Master, 1, messageChannel)
	frame, _ := hex.DecodeString("0103021998b27E0103021998b27E")
	buf := bufio.NewReader(bytes.NewReader(frame))
	go func() {
		err := md.StreamDecoder(buf)
		if err != nil {
			fmt.Println(err)
		}
	}()
	time.Sleep(10 * time.Minute)
}

# go-modbus
modbus_rtu and modbus TCP protocol

**需要自己维护通道连接状态，需自己根据错误判定是否重连**

**You need to maintain the channel connection status yourself and decide whether to reconnect based on errors**

#### TCP
```go
tcp, err := NewModbusTCPPacket("127.0.0.1", 502, defaultConnectTimeout, defaultReadTimeout, defaultWriteTimeout, defaultRwTimeout, ModbusTCP)
if err != nil {
    return
}
err = tcp.Connect()
if err != nil {
    return
}
```

#### RTU
```go
rtu, err := NewModbusRTUPacket("COM2", 9600, 8, 'N', 1, defaultReadTimeout, defaultRwTimeout, ModbusRTU)
if err != nil {
    t.Error(err)
    return
}
err = rtu.Connect()
if err != nil {
    t.Error(err)
    return
}
```
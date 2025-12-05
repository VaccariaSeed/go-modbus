# go-modbus
modbus_rtu and modbus TCP protocol

**需要自己维护通道连接状态，需自己根据错误判定是否重连**

#### 建立连接通道 connector

###### TCP
```go
NewTCPConnector(IP, 端口, 连接超时， 读超时，写超时)
```

###### 串口
```go
NewSerialConnector(串口号, 波特率, 数据位, 校验位, 停止位, 读超时)
```


#### 建立modbus规约模型 packet

###### modbusRTU
```go
NewModbusRTUPacket()
```

###### modbusTCP
```go
NewModbusTCPPacket()
```

###### 创建读modbus设备信息的客户端
```go
client, err := NewModbusClient(packet,connector, 读写间隔)
```
* 设置报文打印方法
```go
client.SendPrintHandler = func(frame []byte) {
	fmt.Println("发送了一条数据：", hex.EncodeToString(frame))
}

client.ReceivePrintHandler = func(frame []byte) {
	fmt.Println("收到了一条数据：", hex.EncodeToString(frame))
}
```
* 开启连接
```go
err := client.Connect()
```
* 关闭连接
```go
err := client.Close()
```
* 刷新
```go
err := client.Flush()
```

###### 如果设备时返回的异常功能码
读写操作的方法都会返回FuncCodeErr，需对错误进行类型判断


**读写方法在client结构体中，注释详细**
package cli

import (
	"encoding/binary"
	"fmt"
	"go-oak/util"
	"go.bug.st/serial"
	"log"
	"time"
)

const (
	rtuMinSize       = 4
	rtuMaxSize       = 256
	rtuExceptionSize = 5
	numSlavesScan    = 20
)

// NewClientDefault 根据给定的参数创建一个 modbus client.
func NewClientDefault(mode *serial.Mode, salveId byte) Client {
	// 寻找可用串口并连接
	port, err := util.ConnectDefault(mode)
	err = port.SetReadTimeout(time.Second)
	if err != nil {
		log.Fatal("连接失败")
		return nil
	}
	return &client{*mode, port, salveId}
}

func NewClient(slaveId byte) (cli Client, err error) {
	// 定义 Mode
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	// 寻找可用串口并连接
	port, err := util.ConnectDefault(mode)
	if err = port.SetReadTimeout(time.Second); err != nil {
		log.Fatal("连接失败")
		return
	}
	cli = &client{*mode, port, slaveId}
	return
}

func CustomClient(mode *serial.Mode, portName string) Client {
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil
	}
	return &client{*mode, port, 0}
}

func TempHumClient() (client Client, err error) {
	// 定义 Mode
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	ports, err := util.GetPorts()
	if err != nil {
		log.Fatal(err)
	}
	for i := range ports {
		client = CustomClient(mode, ports[i])
		for id := 1; id <= numSlavesScan; id++ {
			client.SetSlaveId(byte(id))
			log.Printf("尝试连接 %02d@%s", id, ports[i])
			_, err = client.ReadInputRegisters(1, 1)
			if err != nil {
				log.Printf("连接 %02d@%s 失败\n", id, ports[i])
				continue
			}
			log.Printf("连接 站号%02d@端口%s 成功", id, ports[i])
			return
		}
	}
	return
}

func ChangeSlaveId() {
	// 定义 Mode
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	ports, err := util.GetPorts()
	if err != nil {
		log.Fatal(err)
	}
	for i := range ports {
		client := CustomClient(mode, ports[i])
		flag := false
		for id := 1; id <= numSlavesScan; id++ {
			client.SetSlaveId(byte(id))
			log.Printf("尝试连接 %02d@%s", id, ports[i])
			res, err := client.ReadInputRegisters(1, 1)
			if err != nil {
				log.Printf("连接 %02d@%s 失败\n", id, ports[i])
				continue
			}
			log.Printf("连接 站号%02d@端口%s 成功", id, ports[i])
			fmt.Println("输入想要更改的站号( 输入`0`不更改继续扫描 ):")
			var to uint16
			_, err = fmt.Scanln(&to)
			if err != nil {
				log.Println(err)
			}
			if to != 0 {
				_, err = client.WriteSingleRegister(257, to)
				log.Println("更改成功，请重新插拔设备，按回车继续.")
				if err = client.Close(); err != nil {
					log.Fatal("Close client failed")
				}
				_, _ = fmt.Scanln(&to)
				client = CustomClient(mode, ports[i])
				client.SetSlaveId(byte(to))
				res, err = client.ReadHoldingRegisters(257, 1)
				change, _ := util.BytesToIntU(res)
				log.Printf("寄存器中站号：%d", change)
				if change == int(to) {
					fmt.Printf("更改成功，是否需要继续更改站号(y/N):  ")
					var yes string
					_, _ = fmt.Scan(&yes)
					if yes == "y" {
						log.Println("继续更改")
					} else {
						log.Println("不更改，程序退出")
						flag = true
						break
					}
				}
			}
			if flag {
				break
			}
		}
		if err := client.Close(); err != nil {
			log.Fatal("Close client failed")
		}
	}
}

// client 实现 Client 接口
type client struct {
	serial.Mode
	port    serial.Port
	slaveId byte
}

func (cli *client) SetSlaveId(id byte) {
	cli.slaveId = id
}

func (cli *client) String() string {
	return "Slave ID " + string(cli.slaveId)
}

// Encode 将 PDU 转换成帧并返回
func (cli *client) Encode(pdu *ProtocolDataUnit) (adu []byte, err error) {
	length := len(pdu.Data) + 4
	if length > rtuMaxSize {
		err = fmt.Errorf("modbus: 数据 '%v' 的长度不能大于 '%v'", length, rtuMaxSize)
		return
	}
	adu = make([]byte, length)

	adu[0] = cli.slaveId      // 从设备ID
	adu[1] = pdu.FunctionCode // 功能码
	copy(adu[2:], pdu.Data)   // 传输数据

	// 添加 CRC 校验码
	checksum := util.CheckSum(adu[0 : length-2])
	adu[length-1] = byte(checksum >> 8)
	adu[length-2] = byte(checksum)
	return
}

// Verify 验证，响应长度的是否合法，请求和响应的从机ID是否一致
func (cli *client) Verify(aduRequest []byte, aduResponse []byte) (err error) {
	length := len(aduResponse)
	// 验证是否达到最小响应长度 len(address + function + CRC)
	if length < rtuMinSize {
		err = fmt.Errorf("modbus: 响应长度 '%v' 低于最小长度 '%v'", length, rtuMinSize)
		return
	}
	// 验证从主机ID 是否匹配
	if aduResponse[0] != aduRequest[0] {
		err = fmt.Errorf("modbus: 响应的从主机ID '%v' 与请求ID '%v' 不匹配", aduResponse[0], aduRequest[0])
		return
	}
	return
}

// Decode 从帧中提取 PDU 并对比 checksum 是否匹配，最后返回 PDU。
func Decode(adu []byte) (pdu *ProtocolDataUnit, err error) {
	length := len(adu)
	// 计算 checksum 是否匹配
	realChecksum := util.CheckSum(adu[0 : length-2])
	checksum := uint16(adu[length-1])<<8 | uint16(adu[length-2])
	if checksum != realChecksum {
		err = fmt.Errorf("modbus: response crc '%v' does not match expected '%v'", checksum, realChecksum)
		return
	}
	// 功能码和数据封装
	pdu = &ProtocolDataUnit{
		FunctionCode: adu[1],
		Data:         adu[2 : length-2],
	}
	return
}

// send 发送 PDU，返回响应的 PDU
func (cli *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	adu, err := cli.Encode(request)
	request.Data[0] = cli.slaveId
	if err != nil {
		return
	}
	aduResponse, err := cli.Send(adu)
	if err != nil {
		//log.Println(err, "From send()")
		return
	}
	response, err = Decode(aduResponse)
	if response.FunctionCode != request.FunctionCode { // 发送与响应功能码不同
		err = responseError(response)
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	if response.Data == nil || len(response.Data) == 0 {
		err = fmt.Errorf("modbus: 无数据响应")
		return
	}
	return
}

// Send 发送帧，返回响应的帧
func (cli *client) Send(aduRequest []byte) (aduResponse []byte, err error) {
	if _, err = cli.port.Write(aduRequest); err != nil {
		return
	}
	//log.Printf("TX:% X \n", aduRequest)
	functionalCode := aduRequest[1]
	functionFail := aduRequest[1] & 0x80
	bytesToRead := calculateResponseLength(aduRequest)
	delay := cli.calculateDelay((len(aduRequest) + bytesToRead) * int(aduRequest[1]))
	time.Sleep(delay)
	data := make([]byte, rtuMaxSize)

	// 先读最小的长度，如果无错再读完
	err = cli.port.SetReadTimeout(delay)
	n, err := cli.port.Read(data[:])
	if n == 0 {
		err = &ModbusError{
			FunctionCode:  aduRequest[1],
			ExceptionCode: ExceptionCodeGatewayTargetDeviceFailedToRespond}
		return
	}
	if data[1] == functionalCode {
		aduResponse = data[:n]
	} else if data[1] == functionFail {
		// 串口返回错误码
		aduRequest = data[:rtuExceptionSize]
	} else {
		err = fmt.Errorf("无响应数据")
		return
	}
	if err != nil {
		return
	}
	return
}

func (cli *client) ReadInputRegisters(address, quantity uint16) (result []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadInputRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := cli.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: 响应长度 '%v' 不匹配实际接收长度 '%v'", length, count)
		return
	}
	result = response.Data[1:]
	return
}

func (cli *client) ReadHoldingRegisters(address, quantity uint16) (result []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeReadHoldingRegisters,
		Data:         dataBlock(address, quantity),
	}
	response, err := cli.send(&request)
	if err != nil {
		return
	}
	count := int(response.Data[0])
	length := len(response.Data) - 1
	if count != length {
		err = fmt.Errorf("modbus: 接收到 '%v' bytes，而实际发送了 '%v' bytes", length, count)
		return
	}
	result = response.Data[1:]
	return
}

func (cli *client) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	request := ProtocolDataUnit{
		FunctionCode: FuncCodeWriteSingleRegister,
		Data:         dataBlock(address, value),
	}
	response, err := cli.send(&request)
	if err != nil {
		return
	}
	// 检查响应是否与请求一致
	// 写入单个值 长度就是4
	if len(response.Data) != 4 {
		err = fmt.Errorf("modbus: 响应长度 '%v' 与预期接收长度 '%v' 不匹配", len(response.Data), 4)
		return
	}
	respValue := binary.BigEndian.Uint16(response.Data)
	if address != respValue {
		err = fmt.Errorf("modbus: 响应 Address '%v' 与实际接收 Address '%v' 不匹配", respValue, address)
		return
	}
	results = response.Data[2:]
	respValue = binary.BigEndian.Uint16(results)
	if value != respValue {
		err = fmt.Errorf("modbus: 响应值 '%v' 与实际接收值 '%v' 不匹配", respValue, value)
		return
	}
	return
}

func (cli *client) ReadCoils(address, quantity uint16) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) WriteMultipleRegisters(address, quantity uint16, value []byte) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (results []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (cli *client) Close() (cErr error) {
	cErr = cli.port.Close()
	if cErr != nil {
		log.Fatal(cErr)
	}
	return
}

// dataBlock 把传入的 uint16 数组转换为 byte 数组.
func dataBlock(value ...uint16) []byte {
	data := make([]byte, 2*len(value))
	for i, v := range value {
		binary.BigEndian.PutUint16(data[i*2:], v)
	}
	return data
}

// 计算应该响应数据的长度
func calculateResponseLength(adu []byte) int {
	length := rtuMinSize
	switch adu[1] {
	case FuncCodeReadDiscreteInputs,
		FuncCodeReadCoils:
		count := int(binary.BigEndian.Uint16(adu[4:]))
		length += 1 + count/8
		if count%8 != 0 {
			length++
		}
	case FuncCodeReadInputRegisters,
		FuncCodeReadHoldingRegisters:
		count := int(binary.BigEndian.Uint16(adu[4:]))
		length += 1 + count*2
	case FuncCodeWriteSingleCoil,
		FuncCodeWriteMultipleCoils,
		FuncCodeWriteSingleRegister,
		FuncCodeWriteMultipleRegisters:
		length += 4
	default:
	}
	return length
}

// calculateDelay 简单计算等待响应的时间
// See MODBUS over Serial Line - Specification and Implementation Guide (page 13).
func (cli *client) calculateDelay(chars int) time.Duration {
	var characterDelay, frameDelay int // us
	if cli.BaudRate <= 0 || cli.BaudRate > 19200 {
		characterDelay = 750
		frameDelay = 1750
	} else {
		characterDelay = 15000000 / cli.BaudRate
		frameDelay = 35000000 / cli.BaudRate
	}
	return time.Duration(characterDelay*chars+frameDelay) * time.Microsecond
}

func responseError(response *ProtocolDataUnit) error {
	mbError := &ModbusError{FunctionCode: response.FunctionCode}
	if response.Data != nil && len(response.Data) > 0 {
		mbError.ExceptionCode = response.Data[0]
	}
	return mbError
}

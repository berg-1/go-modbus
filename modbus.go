package main

import "fmt"

// 功能码
const (
	FuncCodeReadCoils          = 1 // 1Bit 访问
	FuncCodeReadDiscreteInputs = 2
	FuncCodeWriteSingleCoil    = 5
	FuncCodeWriteMultipleCoils = 15

	FuncCodeReadHoldingRegisters   = 3 // 16Bit 访问
	FuncCodeReadInputRegisters     = 4
	FuncCodeWriteSingleRegister    = 6
	FuncCodeWriteMultipleRegisters = 16
)

// 异常码
const (
	ExceptionCodeIllegalFunction                    = 1
	ExceptionCodeIllegalDataAddress                 = 2
	ExceptionCodeIllegalDataValue                   = 3
	ExceptionCodeServerDeviceFailure                = 4
	ExceptionCodeAcknowledge                        = 5
	ExceptionCodeServerDeviceBusy                   = 6
	ExceptionCodeMemoryParityError                  = 8
	ExceptionCodeGatewayPathUnavailable             = 10
	ExceptionCodeGatewayTargetDeviceFailedToRespond = 11
)

// ModbusError 实现错误接口
type ModbusError struct {
	FunctionCode  byte
	ExceptionCode byte
}

// Error 转换已知的 Modbus 错误码为错误信息
func (e *ModbusError) Error() string {
	var name string
	switch e.ExceptionCode {
	case ExceptionCodeIllegalFunction:
		name = "illegal function"
	case ExceptionCodeIllegalDataAddress:
		name = "illegal data address"
	case ExceptionCodeIllegalDataValue:
		name = "illegal data value"
	case ExceptionCodeServerDeviceFailure:
		name = "server device failure"
	case ExceptionCodeAcknowledge:
		name = "acknowledge"
	case ExceptionCodeServerDeviceBusy:
		name = "server device busy"
	case ExceptionCodeMemoryParityError:
		name = "memory parity error"
	case ExceptionCodeGatewayPathUnavailable:
		name = "gateway path unavailable"
	case ExceptionCodeGatewayTargetDeviceFailedToRespond:
		name = "gateway target device failed to respond"
	default:
		name = "unknown"
	}
	return fmt.Sprintf("modbus: exception '%v' (%s), function '%v'", e.ExceptionCode, name, e.FunctionCode)
}

// ProtocolDataUnit (PDU) 独立于底层通信层 -> 功能码 + 数据
type ProtocolDataUnit struct {
	SlaveId      byte
	FunctionCode byte
	Data         []byte
}

// Packager 指定会话层
type Packager interface {
	Encode(pdu *ProtocolDataUnit) (adu []byte, err error)
	Decode(adu []byte) (pdu *ProtocolDataUnit, err error)
	Verify(aduRequest []byte, aduResponse []byte) (err error)
}

// Transporter 指定传输层
type Transporter interface {
	Send(aduRequest []byte) (aduResponse []byte, err error)
}

// Client 实现 Modbus 协议的客户端
type Client interface {
	// 1Bit 访问

	// ReadCoils 读取远程设备中线圈的 1 到 2000 个连续状态并返回线圈状态。
	ReadCoils(address, quantity uint16) (results []byte, err error)
	// ReadDiscreteInputs 读取远程设备中离散输入的 1 到 2000 个连续状态并返回输入状态。
	ReadDiscreteInputs(address, quantity uint16) (results []byte, err error)
	// WriteSingleCoil WriteSingleCoil 将单个输出写入 ON 或 OFF 到远程设备中并返回输出值。
	WriteSingleCoil(address, value uint16) (results []byte, err error)
	// WriteMultipleCoils 强制远程设备中的线圈序列中的每个线圈打开（ON）或关闭（OFF），并返回输出数量。
	WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error)

	// 16Bit 访问

	// ReadInputRegisters 读取 1 到 125 个连续输入寄存器并返回输入寄存器。
	ReadInputRegisters(address, quantity uint16) (results []byte, err error)

	// ReadHoldingRegisters 读取连续保持寄存器块的内容并返回寄存器值。
	ReadHoldingRegisters(address, quantity uint16) (results []byte, err error)

	// WriteSingleRegister 写单个保持寄存器值并返回。
	WriteSingleRegister(address, value uint16) (results []byte, err error)

	// WriteMultipleRegisters 写入一个连续寄存器块（1 到 123 个寄存器）并返回寄存器数量。
	WriteMultipleRegisters(address, quantity uint16, value []byte) (results []byte, err error)

	// ReadWriteMultipleRegisters 执行一次读取操作和一次写入操作的组合。 它返回读取的寄存器值。
	ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (results []byte, err error)

	SetSlaveId(id byte)
	// Close 关闭 Client
	Close() (cErr error)
}

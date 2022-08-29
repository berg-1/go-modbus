package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

// BytesToFloat 转换 []bytes（大端） 至 float32
func BytesToFloat(bigEndianBytes []byte) float32 {
	tempInt, err := BytesToIntU(bigEndianBytes)
	if err != nil {
		log.Println("温度转换错误")
		return 0
	}
	return float32(tempInt) / 10
}

func BytesToNFloat(bigEndianBytes []byte, n int) (res []float32) {
	res = make([]float32, n)
	count := 0
	for i := 0; i < n*2; i += 2 {
		res[count] = BytesToFloat(bigEndianBytes[i : i+2])
		count++
	}
	return
}

// BytesToIntU 字节数(大端)组转成int(无符号)
func BytesToIntU(b []byte) (int, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp uint8
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	case 2:
		var tmp uint16
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	case 4:
		var tmp uint32
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

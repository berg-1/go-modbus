package util

import (
	"go.bug.st/serial"
	"log"
)

// GetPorts 获取可用端口列表
func GetPorts() (ports []string, err error) {
	// Retrieve the port list
	ports, err = serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
		return
	}
	// Print the list of detected ports
	for _, port := range ports {
		log.Printf("Found port: %v\n", port)
	}
	return
}

// ConnectDefault 根据给定的 Mode 连接
func ConnectDefault(mode *serial.Mode) (port serial.Port, err error) {
	ports, err := GetPorts()
	if err != nil {
		log.Fatal(err)
	}
	port, err = serial.Open(ports[0], mode)
	if err != nil {
		log.Fatal(err)
	}
	return
}

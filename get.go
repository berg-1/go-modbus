package main

import (
	"fmt"
	"go-oak/cli"
	"go-oak/util"
	"log"
)

func GetTemperAndHumidity() {
	client, _ := cli.NewClient(6)

	inputRegTemp, e3 := client.ReadInputRegisters(1, 1)
	if e3 == nil {
		log.Printf("RX: % X\n", inputRegTemp)
		tf32 := util.BytesToFloat(inputRegTemp)
		log.Printf("目前温度：%.2f℃ \n", tf32)
	}
	inputRegHumidity, e4 := client.ReadInputRegisters(2, 1)
	if e4 == nil {
		log.Printf("RX: % X\n", inputRegHumidity)
		hf32 := util.BytesToFloat(inputRegHumidity)
		log.Printf("目前湿度：%.2f%% \n", hf32)
	}

	res, _ := client.WriteSingleRegister(257, 9)
	fmt.Printf("% X\n", res)
}

func Auto() {
	cli.ChangeSlaveId()
}

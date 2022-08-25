package main

import (
	"fmt"
	"go-oak/util"
	"log"
)

func main() {
	_ = AutoClient()

	//getTemperAndHumidity()
}

func getTemperAndHumidity() {
	cli := NewClientExample(6)
	inputRegTemp, e3 := cli.ReadInputRegisters(1, 1)
	if e3 == nil {
		log.Printf("RX: % X\n", inputRegTemp)
		tf32 := util.BytesToFloat(inputRegTemp)
		log.Printf("目前温度：%.2f℃ \n", tf32)
	}
	inputRegHumidity, e4 := cli.ReadInputRegisters(2, 1)
	if e4 == nil {
		log.Printf("RX: % X\n", inputRegHumidity)
		hf32 := util.BytesToFloat(inputRegHumidity)
		log.Printf("目前湿度：%.2f%% \n", hf32)
	}

	res, _ := cli.WriteSingleRegister(257, 9)
	fmt.Printf("% X\n", res)
}

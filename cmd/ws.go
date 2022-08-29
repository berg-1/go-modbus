package cmd

import (
	"fmt"
	"go-oak/cli"
	"go-oak/util"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var slaveId uint8

// wsCmd represents the ws command
var wsCmd = &cobra.Command{
	Use:   "ws",
	Short: "持续显示当前温湿度",
	Long: `每隔 1 秒，客户端会向连接到的站请求温湿度信息，
然后将请求到的信息转换成温湿度进行打印。

可能会出现一些意想不到的状况。`,
	Run: func(cmd *cobra.Command, args []string) {
		var client cli.Client
		var err error
		if slaveId != 0 {
			log.Println("SlaveId :", slaveId)
			client, err = cli.NewClient(slaveId)
			if err != nil {
				log.Fatalf("连接站号 %d 失败", slaveId)
			}
		} else {
			client, err = cli.TempHumClient()
			if err != nil {
				log.Println("没有站可以响应温湿度")
			}
		}

		SetupCloseHandler(client)

		//设置要接收的信号
		for {
			inputRegTemp, e3 := client.ReadInputRegisters(1, 2)
			if e3 == nil {
				res := util.BytesToNFloat(inputRegTemp, 2)
				fmt.Printf("\r目前温度：%.2f℃ 湿度：%.2f%%", res[0], res[1])
				time.Sleep(time.Second)
			} else {
				log.Println("无法获取温湿度信息")
				break
			}
		}
	},
}

func SetupCloseHandler(client cli.Client) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n进程终止，程序退出...")
		if err := client.Close(); err != nil {
			log.Fatal("客户端关闭失败")
		}
		os.Exit(0)
	}()
}

func init() {
	rootCmd.AddCommand(wsCmd)
	wsCmd.Flags().Uint8VarP(&slaveId, "slave", "s", 0, "要连接的站号")
}

package cmd

import (
	"go-oak/cli"
	"go-oak/util"
	"log"
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
		for {
			inputRegTemp, e3 := client.ReadInputRegisters(1, 1)
			var tf32, hf32 float32
			if e3 == nil {
				tf32 = util.BytesToFloat(inputRegTemp)
			}
			inputRegHumidity, e4 := client.ReadInputRegisters(2, 1)
			if e4 == nil {
				hf32 = util.BytesToFloat(inputRegHumidity)
			}
			log.Printf("目前温度：%.2f℃ 湿度：%.2f%% \n", tf32, hf32)
			time.Sleep(time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(wsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	wsCmd.Flags().Uint8VarP(&slaveId, "slave", "s", 0, "要连接的站号")
}

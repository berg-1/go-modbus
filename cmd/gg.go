package cmd

import (
	"github.com/spf13/cobra"
	"go-oak/cli"
	"log"
)

// ggCmd represents the gg command
var ggCmd = &cobra.Command{
	Use:   "gg",
	Short: "更改站号",
	Long: `循环遍历可用端口，并遍历前 20 个站号，
如果发现可以响应信息的站，询问用户是否更改站号，
输入 0，不更改站号，输入 1-20 中的数字将更改为指定站号。

更改完成会提示插拔设备，用户插拔完成按回车，程序检验是否更改完成
如果更改完成程序退出，否则程序报错。`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("更改站号...")
		cli.ChangeSlaveId()
	},
}

func init() {
	rootCmd.AddCommand(ggCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ggCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ggCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

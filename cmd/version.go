package cmd

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/app"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print out version info and exit. // 打印版本信息并退出。",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s ( Git:%s ) BuidTime:%s\n", app.Version, app.GitTag, app.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

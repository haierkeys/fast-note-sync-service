package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var frontendFiles embed.FS
var configDefault string
var rootCmd = &cobra.Command{
	Use:   "image-api",
	Short: "obsidian image-api gateway",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpTemplate()
		cmd.Help()
	},
}

func Execute(efs embed.FS, c string) {
	frontendFiles = efs
	configDefault = c
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

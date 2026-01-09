package cmd

import (
	"embed"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var frontendFiles embed.FS
var configDefault string
var rootCmd = &cobra.Command{
	Use:   "fast-note-sync-service",
	Short: "Fast Note Sync Service",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpTemplate()
		cmd.Help()
	},
}

func Execute(efs embed.FS, c string) {
	frontendFiles = efs
	configDefault = c
	if err := rootCmd.Execute(); err != nil {
		BootstrapLogger().Error("command execution failed", zap.Error(err))
		os.Exit(1)
	}
}

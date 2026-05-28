package main

import (
	"github.com/hjhsamuel/itoio/app"
	"github.com/spf13/cobra"
)

var (
	config string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Start(config)
	},
}

func init() {
	serverCmd.Flags().StringVarP(&config, "config", "c", "config.yaml", "config file")
	rootCmd.AddCommand(serverCmd)
}

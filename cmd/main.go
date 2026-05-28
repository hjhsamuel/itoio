package main

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ito",
	Short: "A P2P server let you communicate with others",
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

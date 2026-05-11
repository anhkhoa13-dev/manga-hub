package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mangahub",
	Short: "MangaHub CLI - Track your manga reading progress",
	Long:  `A command-line interface for the MangaHub tracking system, supporting HTTP API, TCP sync, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
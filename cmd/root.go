package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudoff",
	Short: "cloudoff is a cost savging tool for cloud resources",
	Long:  "cloudoff is a cost savging tool for cloud resources",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

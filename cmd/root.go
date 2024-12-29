package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "gimer",
	Short: "A CLI timer tool",
	Long:  "Gimer is a CLI tool for setting and managing timers.",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

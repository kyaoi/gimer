package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the current status of all timers",
	Long:  "Display the current status of all timers including their ID, description, and remaining time.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
		}

		mu.Lock()
		defer mu.Unlock()

		if len(timers) == 0 {
			fmt.Println("No timers currently running.")
			return
		}

		fmt.Println("Current timers:")
		for id, timer := range timers {
			remaining := time.Until(timer.Time)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Printf("- ID: %s, Description: %s, Remaining: %s\n",
				id, timer.Description, remaining)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

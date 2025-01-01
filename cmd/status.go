package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the current status of all timers",
	Long:  "Display the current status of all timers including their ID, description, and remaining time.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
		}

		mu.Lock()
		defer mu.Unlock()

		if len(activeTimers) == 0 {
			fmt.Println("No timers currently running.")
			return
		}

		fmt.Println("Current timers:")
		for _, timer := range activeTimers {
			remaining := time.Until(timer.TriggerTimer)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Printf("Description: %s, Remaining: %s\n",
				timer.Description, formatDuration(remaining))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

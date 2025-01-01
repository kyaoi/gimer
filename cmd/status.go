package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	showSavedTimersFlag bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the current status of all timers",
	Long:  "Display the current status of all timers including their ID, description, and remaining time.",
	Run: func(cmd *cobra.Command, args []string) {
		if showSavedTimersFlag {
			if err := loadSavedTimers(); err != nil {
				fmt.Printf("Error loading saved timers: %v\n", err)
			}
			fmt.Println("Current saved timers:")
			for _, timer := range savedTimers {
				fmt.Printf("Description: %s, Time: %s\n",
					timer.Description, timer.Duration)
			}
			return
		}
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
		}

		mu.Lock()
		defer mu.Unlock()

		if len(activeTimers) == 0 {
			fmt.Println("No timers currently running.")
			return
		}

		fmt.Println("Current running timers:")
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
	statusCmd.Flags().BoolVarP(&showSavedTimersFlag, "saved", "s", false, "Show saved timers")
	rootCmd.AddCommand(statusCmd)
}

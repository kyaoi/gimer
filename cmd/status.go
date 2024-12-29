package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of all timers",
	Long:  "Displays the status of all currently running timers.",
	Run: func(cmd *cobra.Command, args []string) {
		// タイマー状態を読み込む
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		timers := getTimers()
		if len(timers) == 0 {
			fmt.Println("No timers found.")
			return
		}

		fmt.Println("Current timers:")
		for _, timer := range timers {
			status := "Running"
			if !timer.TimerRunning {
				status = "Stopped"
			}
			fmt.Printf("- ID: %s, Description: %s, End Time: %s, Status: %s\n",
				timer.ID, timer.Description, timer.ActiveTimerEnd.Format("15:04:05"), status)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

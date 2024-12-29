package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Flag for the stop command
var timerID string

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a specific timer",
	Long:  "Stop a specific timer by its ID.",
	Run: func(cmd *cobra.Command, args []string) {
		if timerID == "" {
			fmt.Println("Please provide a timer ID using the --id flag.")
			return
		}

		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		mu.Lock()
		_, exists := timers[timerID]
		if exists {
			if err := stopTimer(timerID); err != nil {
				fmt.Printf("Error stopping timer: %v\n", err)
			}
		}
		mu.Unlock()

		if !exists {
			fmt.Printf("Timer with ID %s not found.\n", timerID)
			return
		}

		fmt.Printf("Timer with ID %s has been stopped.\n", timerID)
	},
}

func init() {
	stopCmd.Flags().StringVar(&timerID, "id", "", "Specify the ID of the timer to stop")
	rootCmd.AddCommand(stopCmd)
}

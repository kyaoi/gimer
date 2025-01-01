package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	stopAllFlag bool
	deleteFlag  bool
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a specific timer",
	Long:  "Stop a specific timer by its ID.",
	Run: func(cmd *cobra.Command, args []string) {
		if deleteFlag {
			if err := loadSavedTimers(); err != nil {
				fmt.Printf("Error loading saved timers: %v\n", err)
				return
			}

			ids := getSavedTimerIds()
			if err := generateTimerIndex(ids, savedTimerIndexMap); err != nil {
				fmt.Printf("Error generating timer index: %v\n", err)
				return
			}

			if len(savedTimers) == 0 {
				fmt.Println("No timers currently saved.")
				return
			}

			fmt.Println("Available timers:")
			for i, id := range savedTimerIndexMap {
				timer := savedTimers[id]
				fmt.Printf("[%v] Description: %s, Time: %s\n",
					i, timer.Description, timer.Duration)
			}

			fmt.Print("Enter the number of the timer to stop: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				return
			}

			input = input[:len(input)-1]
			timerIndex, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("Invalid input. Please enter a valid number.")
				return
			}

			mu.Lock()
			id, exists := savedTimerIndexMap[timerIndex]
			if !exists {
				fmt.Printf("Timer with index %d not found.\n", timerIndex)
				mu.Unlock()
				return
			}
			if err := deleteSavedTimer(id); err != nil {
				fmt.Printf("Error deleting timer: %v\n", err)
				mu.Unlock()
				return
			}
			mu.Unlock()
			return
		}

		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		if len(activeTimers) == 0 {
			fmt.Println("No timers currently running.")
			return
		}

		if stopAllFlag {
			if err := stopAllTimer(); err != nil {
				fmt.Printf("Error stopping timer: %v\n", err)
				return
			}
			return
		}

		ids := getActiveTimerIds()
		if err := generateTimerIndex(ids, activeTimerIndexMap); err != nil {
			fmt.Printf("Error generating timer index: %v\n", err)
			return
		}

		fmt.Println("Available timers:")
		for i, id := range activeTimerIndexMap {
			timer := activeTimers[id]
			remaining := time.Until(timer.TriggerTimer)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Printf("[%v] Description: %s, Remaining: %s\n",
				i, timer.Description, formatDuration(remaining))
		}

		fmt.Print("Enter the number of the timer to stop: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}

		input = input[:len(input)-1]
		timerIndex, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a valid number.")
			return
		}

		mu.Lock()
		id, exists := activeTimerIndexMap[timerIndex]
		if !exists {
			fmt.Printf("Timer with index %d not found.\n", timerIndex)
			mu.Unlock()
			return
		}
		if err := stopTimer(id); err != nil {
			fmt.Printf("Error stopping timer: %v\n", err)
			mu.Unlock()
			return
		}
		mu.Unlock()
	},
}

func init() {
	// TODO: 保存したタイマーを削除できるようにする
	stopCmd.Flags().BoolVarP(&stopAllFlag, "all", "a", false, "Stop all timers")
	stopCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete a saved timer")
	rootCmd.AddCommand(stopCmd)
}

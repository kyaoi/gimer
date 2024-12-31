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
	stopAll   bool
	timerList bool
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a specific timer",
	Long:  "Stop a specific timer by its ID.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if stopAll && timerList {
			fmt.Println("Error: --all and --list cannot be used together.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		if timerList || (!stopAll) {
			if err := generateTimerIndex(); err != nil {
				fmt.Printf("Error generating timer index: %v\n", err)
				return
			}

			fmt.Println("Available timers:")
			for i, id := range timerIndexMap {
				timer := activeTimers[id]
				remaining := time.Until(timer.Time)
				if remaining < 0 {
					remaining = 0
				}
				fmt.Printf("[%v] Description: %s, Remaining: %s\n",
					i, timer.Description, remaining)
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
			id, exists := timerIndexMap[timerIndex]
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

			return
		}

		if stopAll {
			if err := stopAllTimer(); err != nil {
				fmt.Printf("Error stopping timer: %v\n", err)
				return
			}
			return
		}
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&stopAll, "all", "a", false, "Stop all timers")
	stopCmd.Flags().BoolVarP(&timerList, "", "", false, "Show timers list")
	rootCmd.AddCommand(stopCmd)
}

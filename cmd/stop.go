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
	stopAll bool
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a specific timer",
	Long:  "Stop a specific timer by its ID.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		if len(activeTimers) == 0 {
			fmt.Println("No timers currently running.")
			return
		}

		if stopAll {
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
	stopCmd.Flags().BoolVarP(&stopAll, "all", "a", false, "Stop all timers")
	rootCmd.AddCommand(stopCmd)
}

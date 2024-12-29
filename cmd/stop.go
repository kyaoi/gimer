package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a specific timer",
	Long:  "Stops a specific timer by its ID and updates the timer state.",
	Run: func(cmd *cobra.Command, args []string) {
		// タイマーIDが提供されているか確認
		if timerID == "" {
			fmt.Println("Please provide a timer ID using the --id flag.")
			return
		}

		// タイマー状態を読み込む
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		// タイマーを停止
		if err := stopTimer(timerID); err != nil {
			fmt.Printf("Error stopping timer: %v\n", err)
			return
		}

		fmt.Printf("Timer with ID %s has been stopped.\n", timerID)
	},
}

// フラグを格納する変数
var timerID string

func init() {
	stopCmd.Flags().StringVar(&timerID, "id", "", "Specify the ID of the timer to stop")
	rootCmd.AddCommand(stopCmd)
}

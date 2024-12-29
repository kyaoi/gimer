package cmd

import (
	"bytes"
	"fmt"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Set a timer and play a sound when the timer ends",
	Long: `Set a timer for a specified duration using flags (-s for seconds, -m for minutes, --hours for hours). 
When the timer ends, a default sound or an embedded sound will play for 5 minutes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// タイマー状態を読み込む
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		// タイマーの合計時間を計算
		totalDuration := time.Duration(seconds+minutes*60+hours*3600) * time.Second
		if totalDuration <= 0 {
			fmt.Println("Please specify a valid timer duration using -s, -m, or --hours.")
			return
		}

		// 新しいタイマーを追加
		id, err := addTimer(description, totalDuration)
		if err != nil {
			fmt.Printf("Error adding timer: %v\n", err)
			return
		}

		fmt.Printf("Timer started: ID=%s, Description=%s, Duration=%s\n", id, description, totalDuration)

		// タイマーの実行
		for i := time.Duration(0); i < totalDuration; i += time.Second {
			time.Sleep(time.Second)

			// タイマーが停止されているか確認
			if err := loadState(); err != nil {
				fmt.Printf("Error loading timer state: %v\n", err)
				break
			}

			mu.Lock()
			for _, timer := range timers {
				if timer.ID == id && !timer.TimerRunning {
					fmt.Println("Timer was stopped manually.")
					mu.Unlock()
					return
				}
			}
			mu.Unlock()
		}

		// タイマー終了
		fmt.Printf("Timer ended: ID=%s, Description=%s\n", id, description)
		go func() {
			err := playEmbeddedSoundRepeatedly(id)
			if err != nil {
				fmt.Printf("Error playing sound: %v\n", err)
			}
		}()

		// サウンドを5分後に停止
		time.Sleep(5 * time.Minute)
		if err := stopTimer(id); err != nil {
			fmt.Printf("Error stopping timer sound: %v\n", err)
		}
	},
}

// playEmbeddedSoundRepeatedly 再生を繰り返す関数
func playEmbeddedSoundRepeatedly(id string) error {
	mu.Lock()
	var stopChannel chan struct{}
	for i, timer := range timers {
		if timer.ID == id {
			stopChannel = timer.StopChannel
			timers[i].TimerRunning = true
			saveState()
			break
		}
	}
	mu.Unlock()

	if stopChannel == nil {
		return fmt.Errorf("stop channel not found for timer ID: %s", id)
	}

	streamer, format, err := wav.Decode(bytes.NewReader(embeddedSound))
	if err != nil {
		return fmt.Errorf("could not decode embedded sound: %v", err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("could not initialize speaker: %v", err)
	}

	loop := beep.Loop(-1, streamer)
	speaker.Play(beep.Seq(loop))

	select {
	case <-stopChannel:
		fmt.Println("Stop signal received. Stopping playback.")
	}

	speaker.Close()
	stopTimer(id)

	return nil
}

func init() {
	startCmd.Flags().IntVarP(&seconds, "seconds", "s", 0, "Set timer duration in seconds")
	startCmd.Flags().IntVarP(&minutes, "minutes", "m", 0, "Set timer duration in minutes")
	startCmd.Flags().IntVar(&hours, "hours", 0, "Set timer duration in hours")
	startCmd.Flags().StringVarP(&description, "description", "d", "Timer", "Set a description for the timer")
	rootCmd.AddCommand(startCmd)
}

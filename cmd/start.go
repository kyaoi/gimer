package cmd

import (
	"bytes"
	"fmt"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	seconds     int
	minutes     int
	hours       int
	description string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new timer",
	Long:  "Start a new timer with a specified duration and description.",
	Run: func(cmd *cobra.Command, args []string) {
		duration := time.Duration(seconds+minutes*60+hours*3600) * time.Second
		if duration <= 0 {
			fmt.Println("Please specify a valid duration using --seconds, --minutes, or --hours.")
			return
		}
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
		}

		id := uuid.New().String()
		timer := &Timer{
			ID:          id,
			Description: description,
			Time:        time.Now().Add(duration),
		}

		if err := saveTimer(timer); err != nil {
			fmt.Printf("Error saving timer state: %v\n", err)
			return
		}

		fmt.Printf("Timer started: ID=%s, Description=%s, Duration=%s\n", id, description, duration)

		runTimer(timer)
	},
}

func runTimer(timer *Timer) {
	duration := time.Until(timer.Time)
	time.Sleep(duration)

	fmt.Printf("Timer ended: ID=%s, Description=%s\n", timer.ID, timer.Description)
	playSound(timer)
}

func playSound(timer *Timer) {
	streamer, format, err := wav.Decode(bytes.NewReader(embeddedSound))
	if err != nil {
		fmt.Printf("Error decoding sound: %v\n", err)
		return
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		fmt.Printf("Error initializing speaker: %v\n", err)
		return
	}

	loop := beep.Loop(-1, streamer)
	speaker.Play(beep.Seq(loop))

	duration := 5 * time.Minute
	for remaining := duration; remaining > 0; remaining -= time.Second {
		if err := loadState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
		}

		exists := getSoundRunningStateById(timer.ID)

		if !exists {
			fmt.Println("Sound stopped")
			speaker.Clear()
			break
		}
		time.Sleep(time.Second)
	}

	mu.Lock()
	_, exists := timers[timerID]
	if exists {
		if err := stopTimer(timerID); err != nil {
			fmt.Printf("Error stopping timer: %v\n", err)
		}
	}
	mu.Unlock()
}

func init() {
	startCmd.Flags().IntVarP(&seconds, "seconds", "s", 0, "Set timer duration in seconds")
	startCmd.Flags().IntVarP(&minutes, "minutes", "m", 0, "Set timer duration in minutes")
	startCmd.Flags().IntVar(&hours, "hours", 0, "Set timer duration in hours")
	startCmd.Flags().StringVarP(&description, "description", "d", "Timer", "Set a description for the timer")
	rootCmd.AddCommand(startCmd)
}

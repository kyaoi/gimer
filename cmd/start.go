package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	seconds      int
	minutes      int
	hours        int
	description  string
	save         bool
	useSaveTimer bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new timer",
	Long:  "Start a new timer with a specified duration and description.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if useSaveTimer && seconds != 0 || useSaveTimer && minutes != 0 || useSaveTimer && hours != 0 {
			fmt.Println("Error: --use cannot be used with --seconds, --minutes, or --hours.")
			os.Exit(1)
		}
		if useSaveTimer && description != "Timer" {
			fmt.Println("Error: --use cannot be used with --description.")
			os.Exit(1)
		}
		if useSaveTimer && save {
			fmt.Println("Error: --use and --save cannot be used together.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		if useSaveTimer {
			if err := loadSavedTimers(); err != nil {
				fmt.Printf("Error loading saved timers: %v\n", err)
				return
			}

			if len(savedTimers) == 0 {
				fmt.Println("No timers saved.")
				return
			}

			ids := getSavedTimerIds()
			if err := generateTimerIndex(ids, savedTimerIndexMap); err != nil {
				fmt.Printf("Error generating timer index: %v\n", err)
				return
			}

			fmt.Println("Available timers:")
			for i, id := range savedTimerIndexMap {
				timer := savedTimers[id]
				fmt.Printf("[%v] Description: %s, Duration: %s\n",
					i, timer.Description, timer.Duration)
			}

			fmt.Print("Enter the number of the timer to start: ")
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

			savedTimer := savedTimers[id]
			timer := &ActiveTimer{
				ID:           savedTimer.ID,
				Description:  savedTimer.Description,
				Duration:     savedTimer.Duration,
				TriggerTimer: time.Now().Add(savedTimer.Duration),
			}

			if err := saveActiveTimer(timer); err != nil {
				fmt.Printf("Error start timer: %v\n", err)
				mu.Unlock()
				return
			}
			mu.Unlock()
			fmt.Printf("Timer started: Description=%s, Duration=%s\n", timer.Description, timer.Duration)
			runTimer(timer)
			return
		}

		duration := time.Duration(seconds+minutes*60+hours*3600) * time.Second
		if duration <= 0 {
			fmt.Println("Please specify a valid duration using --seconds, --minutes, or --hours.")
			return
		}

		id := uuid.New().String()
		timer := &ActiveTimer{
			ID:           id,
			Description:  description,
			Duration:     duration,
			TriggerTimer: time.Now().Add(duration),
		}

		if err := saveActiveTimer(timer); err != nil {
			fmt.Printf("Error saving timer state: %v\n", err)
			return
		}

		if save {
			if err := loadSavedTimers(); err != nil {
				fmt.Printf("Error loading saved timers: %v\n", err)
				return
			}
			if err := saveSavedTimers(timer); err != nil {
				fmt.Printf("Error saving timers: %v\n", err)
				return
			}
		}

		fmt.Printf("Timer started: Description=%s, Duration=%s\n", description, duration)

		runTimer(timer)
	},
}

func runTimer(timer *ActiveTimer) {
	duration := time.Until(timer.TriggerTimer)
	time.Sleep(duration)

	fmt.Printf("Timer ended: Description=%s\n", timer.Description)
	playSound(timer)
}

func playSound(timer *ActiveTimer) {
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
		if err := loadActiveTimerState(); err != nil {
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
	_, exists := activeTimers[timerID]
	if exists {
		if err := stopTimer(timerID); err != nil {
			fmt.Printf("Error stopping timer: %v\n", err)
		}
	}
	mu.Unlock()
}

func init() {
	startCmd.Flags().IntVarP(&seconds, "seconds", "S", 0, "Set timer duration in seconds")
	startCmd.Flags().IntVarP(&minutes, "minutes", "M", 0, "Set timer duration in minutes")
	startCmd.Flags().IntVarP(&hours, "hours", "H", 0, "Set timer duration in hours")
	startCmd.Flags().StringVarP(&description, "description", "d", "Timer", "Set a description for the timer")
	startCmd.Flags().BoolVarP(&save, "save", "s", false, "Save timer state to file")
	startCmd.Flags().BoolVarP(&useSaveTimer, "use", "u", false, "Use a saved timer")
	rootCmd.AddCommand(startCmd)
}

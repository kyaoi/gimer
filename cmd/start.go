package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/google/uuid"
	"github.com/jroimartin/gocui"
	"github.com/spf13/cobra"
)

var (
	seconds          int
	minutes          int
	hours            int
	description      string
	saveFlag         bool
	useSaveTimerFlag bool
)

var running int32 = 1 // カウントダウンの実行状態 (1: 実行中, 0: 停止中)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new timer",
	Long:  "Start a new timer with a specified duration and description.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if useSaveTimerFlag && seconds != 0 || useSaveTimerFlag && minutes != 0 || useSaveTimerFlag && hours != 0 {
			fmt.Println("Error: --use cannot be used with --seconds, --minutes, or --hours.")
			os.Exit(1)
		}
		if useSaveTimerFlag && description != "Timer" {
			fmt.Println("Error: --use cannot be used with --description.")
			os.Exit(1)
		}
		if useSaveTimerFlag && saveFlag {
			fmt.Println("Error: --use and --save cannot be used together.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadActiveTimerState(); err != nil {
			fmt.Printf("Error loading timer state: %v\n", err)
			return
		}

		if useSaveTimerFlag {
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

		if saveFlag {
			if err := loadSavedTimers(); err != nil {
				fmt.Printf("Error loading saved timers: %v\n", err)
				return
			}
			if err := saveSavedTimers(timer); err != nil {
				fmt.Printf("Error saving timers: %v\n", err)
				return
			}
		}

		runTimer(timer)
	},
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	v, err := g.SetView("timer", maxX/2-35, maxY/2-9, maxX/2+35, maxY/2+9)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	v.Frame = false
	return nil
}

func displayTime(g *gocui.Gui, t *ActiveTimer, d time.Duration) {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("timer")
		if err != nil {
			return err
		}
		v.Clear()
		timeStr := fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
		fig := figure.NewFigure(timeStr, "big", true)
		fmt.Fprint(v, fig.String())
		if err := loadActiveTimerState(); err != nil {
			return fmt.Errorf("Error loading timer state: %v\n", err)
		}
		exists := getActiveTimerById(t.ID)
		if !exists {
			return gocui.ErrQuit
		}
		return nil
	})
}

func updateTimer(g *gocui.Gui, t *ActiveTimer, d time.Duration) {
	for r := d; r > 0; r -= time.Second {
		for atomic.LoadInt32(&running) == 0 {
			time.Sleep(100 * time.Millisecond) // 停止中は待機
			exists := getActiveTimerById(t.ID)
			if !exists {
				os.Exit(0)
			}
		}
		g.Update(func(g *gocui.Gui) error {
			displayTime(g, t, (r+time.Second-1)/time.Second*time.Second) // 小数秒を切り上げて表示
			return nil
		})
		time.Sleep(1 * time.Second)
	}
	displayTime(g, t, 0) // 完全に0になったときに表示を更新
	playSound(t)         // 0秒になった瞬間にサウンドを再生
}

func togglePause(g *gocui.Gui, v *gocui.View) error {
	if atomic.LoadInt32(&running) == 1 {
		atomic.StoreInt32(&running, 0) // 停止
	} else {
		atomic.StoreInt32(&running, 1) // 再開
	}
	return nil
}

func runTimer(timer *ActiveTimer) {
	duration := time.Until(timer.TriggerTimer)
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	go updateTimer(g, timer, duration)

	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone, togglePause); err != nil { // Enterキーで開始・停止
		panic(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
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

		exists := getActiveTimerById(timer.ID)

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
	startCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, "Save timer state to file")
	startCmd.Flags().BoolVarP(&useSaveTimerFlag, "use", "u", false, "Use a saved timer")
	rootCmd.AddCommand(startCmd)
}

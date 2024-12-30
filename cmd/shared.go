package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Timer struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Time        time.Time `json:"active_timer_end"`
}

//go:embed resources/sound.wav
var embeddedSound []byte

var (
	mu     sync.Mutex
	timers = make(map[string]*Timer)
)

var timerID string

var timerIndexMap = make(map[int]string)

func getStateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	stateDir := filepath.Join(homeDir, ".gimer")
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		if err := os.Mkdir(stateDir, 0755); err != nil {
			return "", err
		}
	}

	return filepath.Join(stateDir, "timer_state.json"), nil
}

func loadState() error {
	filePath, err := getStateFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var loadedTimers map[string]*Timer
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&loadedTimers); err != nil {
		return fmt.Errorf("Failed to decode timer state: %v\n", err)
	}

	timers = loadedTimers

	return nil
}

func saveTimer(timer *Timer) error {
	timers[timer.ID] = timer
	filePath, err := getStateFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(timers)
}

func stopTimer(id string) error {
	if timer, exists := timers[id]; exists {
		filePath, err := getStateFilePath()
		if err != nil {
			return fmt.Errorf("Error loading timer state json file: %v\n", err)
		}

		fmt.Println("Stopping timer with ID:", id)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		delete(timers, id)
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(timers); err != nil {
			if err := saveTimer(timer); err != nil {
				return fmt.Errorf("Error saving timer state: %v\n", err)
			}
		}
		fmt.Println("Ended timer with description:", timer.Description)
		return nil
	}
	return fmt.Errorf("Timer not found")
}

func stopAllTimer() error {
	filePath, err := getStateFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	timerBackup := timers
	timers = make(map[string]*Timer)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(timers); err != nil {
		timers = timerBackup
		return fmt.Errorf("Error saving timer state: %v\n", err)
	}
	return nil
}

func getSoundRunningStateById(id string) bool {
	_, exists := timers[id]
	return exists
}

func generateTimerIndex() error {
	number := 1
	for _, timer := range timers {
		timerIndexMap[number] = timer.ID
		number++
	}
	return nil
}

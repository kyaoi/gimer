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

// Timer represents a single timer's state
type Timer struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Time        time.Time `json:"active_timer_end"`
}

//go:embed resources/sound.wav
var embeddedSound []byte

// Global variables for managing timers
var (
	mu     sync.Mutex
	timers = make(map[string]*Timer)
)

// Get the file path for the state file
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

	timers = loadedTimers // メモリ内のtimersマップを更新

	return nil
}

func getSoundRunningStateById(id string) bool {
	for _, timer := range timers {
		fmt.Println("timer: ", timer)
	}
	_, exists := timers[id]
	return exists
}

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

var activeTimerFile = "active_timer.json"
var savedTimersFile = "timers.json"

var (
	mu           sync.Mutex
	activeTimers = make(map[string]*Timer)
	savedTimers  = make(map[string]*Timer)
)

var timerID string

var timerIndexMap = make(map[int]string)

func getGimerDir() (string, error) {
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

	return stateDir, nil

}

func getSavedTimersFilePath() (string, error) {
	d, err := getGimerDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, savedTimersFile), nil
}

func getActiveTimerFilePath() (string, error) {
	d, err := getGimerDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, activeTimerFile), nil
}

func openOrCreateFile(p string) (*os.File, error) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return os.Create(p)
	} else if err != nil {
		return nil, fmt.Errorf("error checking file existence: %v", err)
	}

	return os.Open(p)
}

func loadSavedTimers() error {
	p, err := getSavedTimersFilePath()
	if err != nil {
		return fmt.Errorf("Error loading saved timers json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := openOrCreateFile(p)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Error getting file info: %v", err)
	}

	if stat.Size() == 0 {
		savedTimers = make(map[string]*Timer)
		return nil
	}

	var loadedTimers map[string]*Timer
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&loadedTimers); err != nil {
		return fmt.Errorf("Failed to decode saved timers: %v\n", err)
	}

	savedTimers = loadedTimers

	return nil
}

func loadActiveTimerState() error {
	p, err := getActiveTimerFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := openOrCreateFile(p)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Error getting file info: %v", err)
	}

	if stat.Size() == 0 {
		activeTimers = make(map[string]*Timer)
		return nil
	}

	var loadedTimers map[string]*Timer
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&loadedTimers); err != nil {
		return fmt.Errorf("Failed to decode timer state: %v\n", err)
	}

	activeTimers = loadedTimers

	return nil
}

func saveSavedTimers(timer *Timer) error {
	savedTimers[timer.ID] = timer
	p, err := getSavedTimersFilePath()
	if err != nil {
		return fmt.Errorf("Error loading saved timers json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(savedTimers)

}

func saveActiveTimer(timer *Timer) error {
	activeTimers[timer.ID] = timer
	p, err := getActiveTimerFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	mu.Lock()
	defer mu.Unlock()

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(activeTimers)
}

func stopTimer(id string) error {
	if timer, exists := activeTimers[id]; exists {
		p, err := getActiveTimerFilePath()
		if err != nil {
			return fmt.Errorf("Error loading timer state json file: %v\n", err)
		}

		fmt.Println("Stopping timer with ID:", id)
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		defer f.Close()

		delete(activeTimers, id)
		encoder := json.NewEncoder(f)
		if err := encoder.Encode(activeTimers); err != nil {
			if err := saveActiveTimer(timer); err != nil {
				return fmt.Errorf("Error saving timer state: %v\n", err)
			}
		}
		fmt.Println("Ended timer with description:", timer.Description)
		return nil
	}
	return fmt.Errorf("Timer not found")
}

func stopAllTimer() error {
	p, err := getActiveTimerFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	activeTimersBackup := activeTimers
	activeTimers = make(map[string]*Timer)
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(activeTimers); err != nil {
		activeTimers = activeTimersBackup
		return fmt.Errorf("Error saving timer state: %v\n", err)
	}
	return nil
}

func getSoundRunningStateById(id string) bool {
	_, exists := activeTimers[id]
	return exists
}

func generateTimerIndex() error {
	number := 1
	for _, timer := range activeTimers {
		timerIndexMap[number] = timer.ID
		number++
	}
	return nil
}

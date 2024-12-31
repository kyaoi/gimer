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

type SavedTimer struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
}
type ActiveTimer struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`
	Duration     time.Duration `json:"duration"`
	TriggerTimer time.Time     `json:"triggerTimer"`
}

//go:embed resources/sound.wav
var embeddedSound []byte

var activeTimersFile = "active_timers.json"
var savedTimersFile = "timers.json"

var (
	mu           sync.Mutex
	activeTimers = make(map[string]*ActiveTimer)
	savedTimers  = make(map[string]*SavedTimer)
)

var timerID string

var activeTimerIndexMap = make(map[int]string)
var savedTimerIndexMap = make(map[int]string)

var activeTimerIds []string
var savedTimerIds []string

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
	return filepath.Join(d, activeTimersFile), nil
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
		savedTimers = make(map[string]*SavedTimer)
		return nil
	}

	var loadedTimers map[string]*SavedTimer
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
		activeTimers = make(map[string]*ActiveTimer)
		return nil
	}

	var loadedTimers map[string]*ActiveTimer
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&loadedTimers); err != nil {
		return fmt.Errorf("Failed to decode timer state: %v\n", err)
	}

	activeTimers = loadedTimers

	return nil
}

func saveSavedTimers(timer *ActiveTimer) error {
	savedTimer := &SavedTimer{
		ID:          timer.ID,
		Description: timer.Description,
		Duration:    timer.Duration,
	}
	savedTimers[savedTimer.ID] = savedTimer
	p, err := getSavedTimersFilePath()
	if err != nil {
		return fmt.Errorf("Error loading saved timers json file: %v\n", err)
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	return encoder.Encode(savedTimers)

}

func saveActiveTimer(timer *ActiveTimer) error {
	activeTimers[timer.ID] = timer
	p, err := getActiveTimerFilePath()
	if err != nil {
		return fmt.Errorf("Error loading timer state json file: %v\n", err)
	}

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
	activeTimers = make(map[string]*ActiveTimer)
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

func getSavedTimerIds() []string {
	savedTimerIds = make([]string, 0)
	for id := range savedTimers {
		savedTimerIds = append(savedTimerIds, id)
	}
	return savedTimerIds
}

func getActiveTimerIds() []string {
	activeTimerIds = make([]string, 0)
	for id := range activeTimers {
		activeTimerIds = append(activeTimerIds, id)
	}
	return activeTimerIds
}

func generateTimerIndex(ids []string, m map[int]string) error {
	number := 1
	for _, id := range ids {
		m[number] = id
		number++
	}
	return nil
}

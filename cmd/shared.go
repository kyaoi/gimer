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

var (
	seconds     int
	minutes     int
	hours       int
	description string
)

// タイマーの共有状態
var (
	mu     sync.Mutex
	timers []TimerState // 複数のタイマー状態
)

//go:embed resources/sound.wav
var embeddedSound []byte

// 状態構造体
type TimerState struct {
	ID             string        `json:"id"`
	Description    string        `json:"description"`
	TimerRunning   bool          `json:"timer_running"`
	ActiveTimerEnd time.Time     `json:"active_timer_end"`
	StopChannel    chan struct{} `json:"-"` // JSONには保存しない
}

// 状態ファイルのパスを取得
func getStateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// ディレクトリパス
	gimerDir := filepath.Join(homeDir, ".gimer")
	// ディレクトリが存在しない場合は作成
	if _, err := os.Stat(gimerDir); os.IsNotExist(err) {
		if err := os.Mkdir(gimerDir, 0755); err != nil {
			return "", err
		}
	}
	// 状態ファイルのフルパス
	return filepath.Join(gimerDir, "timer_state.json"), nil
}

// 状態を保存
func saveState() error {
	stateFilePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(stateFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(timers)
}

// 状態を読み取り
func loadState() error {
	stateFilePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	file, err := os.Open(stateFilePath)
	if os.IsNotExist(err) {
		// ファイルが存在しない場合は初期化
		timers = []TimerState{}
		return nil
	} else if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&timers); err != nil {
		// 不正な形式の場合、初期化して上書き
		timers = []TimerState{}
		if err := saveState(); err != nil {
			return err
		}
	}

	// チャネルの初期化
	for i := range timers {
		if timers[i].TimerRunning {
			timers[i].StopChannel = make(chan struct{})
		}
	}

	return nil
}

// 新しいタイマーを追加
func addTimer(description string, duration time.Duration) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	id := generateID()
	newTimer := TimerState{
		ID:             id,
		Description:    description,
		TimerRunning:   true,
		ActiveTimerEnd: time.Now().Add(duration),
		StopChannel:    make(chan struct{}),
	}

	timers = append(timers, newTimer)
	return id, saveState()
}

// タイマーを停止
func stopTimer(id string) error {
	mu.Lock()
	defer mu.Unlock()

	for i, timer := range timers {
		if timer.ID == id {
			// タイマーを停止
			timers[i].TimerRunning = false

			// サウンド再生中なら停止
			if timers[i].StopChannel != nil {
				close(timers[i].StopChannel) // チャネルを閉じることで再生を停止
				timers[i].StopChannel = nil  // チャネルをクリア
			}

			// 状態を保存
			return saveState()
		}
	}
	return fmt.Errorf("timer not found")
}

// 全タイマーの状態を取得
func getTimers() []TimerState {
	mu.Lock()
	defer mu.Unlock()

	return timers
}

// タイマーIDを生成
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

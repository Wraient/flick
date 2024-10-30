package internal

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

    "github.com/Microsoft/go-winio"
)


func PlayWithMPV(url string) (string, error) {	
	// Create a unique socket path in /tmp
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("mpv-socket-flick-%d", time.Now().UnixNano()))

	// Start mpv with IPC socket and fullscreen
	cmd := exec.Command("mpv", "--fs", "--input-ipc-server="+socketPath, url)
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start mpv: %w", err)
	}

	return socketPath, nil
}


func MPVSendCommand(ipcSocketPath string, command []interface{}) (interface{}, error) {
    var conn net.Conn
    var err error

    if runtime.GOOS == "windows" {
        // Use named pipe for Windows
        conn, err = winio.DialPipe(ipcSocketPath, nil)
    } else {
        conn, err = net.Dial("unix", ipcSocketPath)
    }
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    commandStr, err := json.Marshal(map[string]interface{}{
        "command": command,
    })
    if err != nil {
        return nil, err
    }

    // Send the command
    _, err = conn.Write(append(commandStr, '\n'))
    if err != nil {
        return nil, err
    }

    // Receive the response
    buf := make([]byte, 4096)
    n, err := conn.Read(buf)
    if err != nil {
        return nil, err
    }

    var response map[string]interface{}
    if err := json.Unmarshal(buf[:n], &response); err != nil {
        return nil, err
    }

    if data, exists := response["data"]; exists {
        return data, nil
    }

    return nil, nil
}

func SeekMPV(ipcSocketPath string, time int) (interface{}, error) {
    command := []interface{}{"seek", time, "absolute"}
    return MPVSendCommand(ipcSocketPath, command)
}

func GetMPVPausedStatus(ipcSocketPath string) (bool, error) {
    status, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "pause"})
    if err != nil || status == nil {
        return false, err
    }

    paused, ok := status.(bool)
    if ok {
        return paused, nil
    }
    return false, nil
}

func GetMPVPlaybackSpeed(ipcSocketPath string) (float64, error) {
	userFlickConfig := GetGlobalConfig()
	if userFlickConfig == nil || userFlickConfig.StoragePath == "" {
		var homeDir string
		if runtime.GOOS == "windows" {
			homeDir = os.Getenv("USERPROFILE")
		} else {
			homeDir = os.Getenv("HOME")
		}	
		userFlickConfig.StoragePath = filepath.Join(homeDir, ".local", "share", "flick")
	}
	logFile := filepath.Join(os.ExpandEnv(userFlickConfig.StoragePath), "debug.log")

    speed, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "speed"})
    if err != nil || speed == nil {
        Log("Failed to get playback speed.", logFile)
        return 0, err
    }

    currentSpeed, ok := speed.(float64)
    if ok {
        return currentSpeed, nil
    }

    return 0, nil
}

func GetPercentageWatched(ipcSocketPath string) (float64, error) {
    currentTime, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "time-pos"})
    if err != nil || currentTime == nil {
        return 0, err
    }

    duration, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "duration"})
    if err != nil || duration == nil {
        return 0, err
    }

    currTime, ok1 := currentTime.(float64)
    dur, ok2 := duration.(float64)

    if ok1 && ok2 && dur > 0 {
        percentageWatched := (currTime / dur) * 100
        return percentageWatched, nil
    }

    return 0, nil
}

func PercentageWatched(playbackTime int, duration int) float64 {
    if duration > 0 {
        percentage := (float64(playbackTime) / float64(duration)) * 100
        return percentage
    }
    return float64(0)
}

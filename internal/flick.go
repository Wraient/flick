package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// ClearScreen clears the terminal screen and saves the state
func ClearScreen() {
    fmt.Print("\033[?1049h") // Switch to alternate screen buffer
    fmt.Print("\033[2J")     // Clear the entire screen
    fmt.Print("\033[H")      // Move cursor to the top left
}

// RestoreScreen restores the original terminal state
func RestoreScreen() {
    fmt.Print("\033[?1049l") // Switch back to the main screen buffer
}

func ExitOcto(msg string, err error) {
	RestoreScreen()
	OctoOut("Have a great day!")
	if err != nil {
		OctoOut(err)
		if runtime.GOOS == "windows" {
			fmt.Println("Press Enter to exit")
			var wait string
			fmt.Scanln(&wait)
			os.Exit(1)
		} else {
			OctoOut(fmt.Sprintf("Error: %v", err))
			os.Exit(1)
		}
	}
	if msg != "" {
		OctoOut(msg)
	}
	os.Exit(0)
}

func OctoOut(data interface{}) {
	userOctoConfig := GetGlobalConfig()
	if userOctoConfig == nil || userOctoConfig.StoragePath == "" {
		var homeDir string
		if runtime.GOOS == "windows" {
			homeDir = os.Getenv("USERPROFILE")
		} else {
			homeDir = os.Getenv("HOME")
		}	
		userOctoConfig.StoragePath = filepath.Join(homeDir, ".local", "share", "octo")
	}
	logFile := filepath.Join(os.ExpandEnv(userOctoConfig.StoragePath), "debug.log")

	userConfig := GetGlobalConfig()
	dataStr := fmt.Sprintf("%v", data)
	if userConfig.RofiSelection && runtime.GOOS != "windows" {
		cmd := exec.Command("notify-send", dataStr)
		err := cmd.Run()
		if err != nil {
			Log(fmt.Sprintf("%v", cmd), logFile)
			Log(fmt.Sprintf("Failed to send notification: %v", err), logFile)
		}
	} else {
		fmt.Println(data)
	}
}

// LogData logs the input data into a specified log file with the format [LOG] time lineNumber: logData
func Log(data interface{}, logFile string) error {
	// Open or create the log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close() // Ensure the file is closed when done

	// Attempt to marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Get the caller information
	_, filename, lineNumber, ok := runtime.Caller(1) // Caller 1 gives the caller of LogData
	if !ok {
		return fmt.Errorf("unable to get caller information")
	}

	// Log the current time and the JSON representation along with caller info
	currentTime := time.Now().Format("2006/01/02 15:04:05")
	logMessage := fmt.Sprintf("[LOG] %s %s:%d: %s\n", currentTime, filename, lineNumber, jsonData)
	_, err = fmt.Fprint(file, logMessage) // Write to the file
	if err != nil {
		return err
	}

	return nil
}

func ClearLog(logFile string) error {
	os.Remove(logFile)
	return nil
}

func CheckAndDownloadFiles(storagePath string, filesToCheck []string) error {
	// Create storage directory if it doesn't exist
	storagePath = os.ExpandEnv(storagePath)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Base URL for downloading config files
	baseURL := "https://raw.githubusercontent.com/Wraient/octo/refs/heads/main/rofi/"

	// Check each file
	for _, fileName := range filesToCheck {
		filePath := filepath.Join(os.ExpandEnv(storagePath), fileName)

		// Skip if file already exists
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		// Download file if it doesn't exist
		resp, err := http.Get(baseURL + fileName)
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", fileName, err)
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", fileName, err)
		}
		defer out.Close()

		// Write the content
		if _, err := io.Copy(out, resp.Body); err != nil {
			return fmt.Errorf("failed to write file %s: %v", fileName, err)
		}
	}

	return nil
}

func UpdateOcto(repo, fileName string) error {
    // Get the path of the currently running executable
    executablePath, err := os.Executable()
    if err != nil {
        return fmt.Errorf("unable to find current executable: %v", err)
    }

    // Adjust file name for Windows
    if runtime.GOOS == "windows" {
        fileName += ".exe"
    }

    // GitHub release URL for curd
    url := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", repo, fileName)

    // Temporary path for the downloaded curd executable
    tmpPath := executablePath + ".tmp"

    // Download the curd executable
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("failed to download file: %v", err)
    }
    defer resp.Body.Close()

    // Check if the download was successful
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to download file: received status code %d", resp.StatusCode)
    }

    // Create a new temporary file
    out, err := os.Create(tmpPath)
    if err != nil {
        return fmt.Errorf("failed to create temporary file: %v", err)
    }
    defer out.Close()

    // Write the downloaded content to the temporary file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return fmt.Errorf("failed to save downloaded file: %v", err)
    }

    // Close and rename the temporary file to replace the current executable
    out.Close()

    // Replace the current executable with the downloaded curd
    if err := os.Rename(tmpPath, executablePath); err != nil {
        return fmt.Errorf("failed to replace the current executable: %v", err)
    }
    OctoOut(fmt.Sprintf("Downloaded curd executable to %v", executablePath))

	if runtime.GOOS != "windows" {
		// Ensure the new file has executable permissions
		if err := os.Chmod(executablePath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions on the new file: %v", err)
		}
	}
	
    return nil
}

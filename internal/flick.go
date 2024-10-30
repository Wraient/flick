package internal

import (
	"encoding/json"
	"fmt"
	"os"
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

func ExitFlick(err error) {
	RestoreScreen()
	FlickOut("Have a great day!")
	if err != nil {
		FlickOut(err)
		if runtime.GOOS == "windows" {
			fmt.Println("Press Enter to exit")
			var wait string
			fmt.Scanln(&wait)
			os.Exit(1)
		} else {
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func FlickOut(data interface{}) {
	fmt.Printf("%v\n", data)
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
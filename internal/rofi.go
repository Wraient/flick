package internal

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func RofiSelect(options map[string]string, addanimeopt bool) (SelectionOption, error) {
	userCurdConfig := GetGlobalConfig()
	if userCurdConfig.StoragePath == "" {
		userCurdConfig.StoragePath = os.ExpandEnv("${HOME}/.local/share/flick")
	}

	// Create a slice to store the options in the order we want
	var optionsList []string
	for _, value := range options {
		optionsList = append(optionsList, value)
	}
	
	// Sort the options alphabetically
	sort.Strings(optionsList)
	
	// Add "Add new anime" and "Quit" options
	if addanimeopt {
		optionsList = append(optionsList, "Add new show", "Quit")
	} else {
		optionsList = append(optionsList, "Quit")
	}
	
	// Join all options into a single string, separated by newlines
	optionsString := strings.Join(optionsList, "\n")
	
	// Prepare the Rofi command
	cmd := exec.Command("rofi", "-dmenu", "-theme", filepath.Join(os.ExpandEnv(userCurdConfig.StoragePath), "selectanime.rasi"), "-i", "-p", "Select a show")
	
	// Set up pipes for input and output
	cmd.Stdin = strings.NewReader(optionsString)
	var out bytes.Buffer
	cmd.Stdout = &out
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		return SelectionOption{}, fmt.Errorf("failed to run Rofi: %v", err)
	}
	
	// Get the selected option
	selected := strings.TrimSpace(out.String())
	
	// Handle special cases
	switch selected {
	case "":
		return SelectionOption{}, fmt.Errorf("no selection made")
	case "Add new show":
		return SelectionOption{Label: "Add new show", Key: "add_new"}, nil
	case "Quit":
		return SelectionOption{Label: "Quit", Key: "-1"}, nil
	}
	
	// Find the key for the selected value
	for key, value := range options {
		if value == selected {
			return SelectionOption{Label: value, Key: key}, nil
		}
	}
	
	// If we get here, the selected option wasn't found in the original map
	return SelectionOption{}, fmt.Errorf("selected option not found in original list")
}


// GetUserInputFromRofi prompts the user for input using Rofi with a custom message
func GetUserInputFromRofi(message string) (string, error) {
	userCurdConfig := GetGlobalConfig()
	if userCurdConfig.StoragePath == "" {
		userCurdConfig.StoragePath = os.ExpandEnv("${HOME}/.local/share/curd")
	}
	// Create the Rofi command
	cmd := exec.Command("rofi", "-dmenu", "-theme", filepath.Join(os.ExpandEnv(userCurdConfig.StoragePath), "userinput.rasi"), "-p", "Input", "-mesg", message)
	
	// Set up pipes for output
	var out bytes.Buffer
	cmd.Stdout = &out
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run Rofi: %w", err)
	}
	
	// Get the entered input
	userInput := strings.TrimSpace(out.String())
	
	return userInput, nil
}
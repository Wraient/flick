package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Wraient/flick/internal"
)

// Replace the directory navigation code in main() with:
func browseDirectory(dirID string) (string, error) {
    for {
        dir, err := internal.GetVadapav(dirID)
        if err != nil {
            return "", err
        }

        fileOptions := make(map[string]string)
        
        // Add parent directory option if not at root
        if dir.Parent != "" {
            fileOptions[".."] = "📁 .."
        }

        for _, file := range dir.Files {
            if file.Dir {
                fileOptions[file.Id] = fmt.Sprintf("📁 %s", file.Name)
            } else if strings.HasSuffix(strings.ToLower(file.Name), ".mkv") || 
                      strings.HasSuffix(strings.ToLower(file.Name), ".mp4") {
                fileOptions[file.Id] = fmt.Sprintf("🎬 %s", file.Name)
            }
        }

        if len(fileOptions) == 0 {
            return "", fmt.Errorf("no files found in directory")
        }

        selectedFile, err := internal.DynamicSelect(fileOptions)
        if err != nil || selectedFile.Key == "-1" {
            return "", err
        }

        // Handle parent directory navigation
        if selectedFile.Key == ".." {
            dirID = dir.Parent
            continue
        }

        // Check if selected item is a directory
        for _, file := range dir.Files {
            if file.Id == selectedFile.Key {
                if file.Dir {
                    dirID = file.Id
                    continue
                } else {
                    return file.Id, nil
                }
            }
        }
    }
}


func main() {
	var user internal.User
	var show internal.TVShow
	vadapavPlaybackUrl := "https://dl2.vadapav.mov/f/"

    var homeDir string
	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
	} else {
		homeDir = os.Getenv("HOME")
	}

	configFilePath := filepath.Join(homeDir, ".config", "flick", "flick.conf")

	// load flick userFlickConfig
	userFlickConfig, err := internal.LoadConfig(configFilePath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	internal.SetGlobalConfig(&userFlickConfig)

	databaseFile := filepath.Join(os.ExpandEnv(userFlickConfig.StoragePath), "shows.db")
	logFile := filepath.Join(os.ExpandEnv(userFlickConfig.StoragePath), "debug.log")

	// Flags configured here cause userconfig needs to be changed.
	flag.StringVar(&userFlickConfig.Player, "player", userFlickConfig.Player, "Player to use for playback (Only mpv supported currently)")
	flag.StringVar(&userFlickConfig.StoragePath, "storage-path", userFlickConfig.StoragePath, "Path to the storage directory")
	flag.IntVar(&userFlickConfig.PercentageToMarkComplete, "percentage-to-mark-complete", userFlickConfig.PercentageToMarkComplete, "Percentage to mark episode as complete")
	flag.BoolVar(&userFlickConfig.SaveMpvSpeed, "save-mpv-speed", userFlickConfig.SaveMpvSpeed, "Save MPV speed setting (true/false)")
	flag.BoolVar(&userFlickConfig.NextEpisodePrompt, "next-episode-prompt", userFlickConfig.NextEpisodePrompt, "Prompt for the next episode (true/false)")

	// Boolean flags that accept true/false
	rofiSelection := flag.Bool("rofi", false, "Open selection in rofi")
	editConfig := flag.Bool("e", false, "Edit config file")
	noRofi := flag.Bool("no-rofi", false, "No rofi")
	updateScript := flag.Bool("update", false, "Update the script")

	// Custom help/usage function
	flag.Usage = func() {
		internal.RestoreScreen()
		fmt.Fprintf(os.Stderr, "Flick is a CLI tool to manage and watch TV shows/Movies playback.\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults() // This prints the default flag information
	}

	flag.Parse()

	if *updateScript {
		repo := "wraient/flick"
		fileName := "flick"

		if err := internal.UpdateFlick(repo, fileName); err != nil {
			internal.ExitFlick(fmt.Sprintf("Error updating executable: %v", err), err)
		} else {
			internal.ExitFlick("Program Updated!", nil)
		}
	}

    if *editConfig {
        editor := os.Getenv("EDITOR")
        if editor == "" {
            if runtime.GOOS == "windows" {
                editor = "notepad"
            } else {
                editor = "vim" 
            }
        }
        cmd := exec.Command(editor, os.ExpandEnv(configFilePath))
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            fmt.Printf("Error opening editor: %v\n", err)
        }
        internal.Log(fmt.Sprintf("Editing config file: %s", os.ExpandEnv(configFilePath)), logFile)
        internal.ExitFlick("Config file updated!", nil)
    }

	if *rofiSelection {
		userFlickConfig.RofiSelection = true
	}

	if *noRofi || runtime.GOOS == "windows" {
		userFlickConfig.RofiSelection = false
	}

    // Download Rofi config's if not already present
	if userFlickConfig.RofiSelection {
		// Define a slice of file names to check and download
		filesToCheck := []string{
			"selectanimepreview.rasi",
			"selectanime.rasi",
			"userinput.rasi",
		}

		// Call the function to check and download files
		err := internal.CheckAndDownloadFiles(os.ExpandEnv(userFlickConfig.StoragePath), filesToCheck)
		if err != nil {
			internal.Log(fmt.Sprintf("Error checking and downloading files: %v\n", err), logFile)
			internal.FlickOut(fmt.Sprintf("Error checking and downloading files: %v\n", err))
			internal.ExitFlick("", err)
		}
	}

	internal.ClearLog(logFile)
	// Get all shows from database
	shows := internal.LocalGetAllShows(databaseFile)
	
	if len(shows) > 0 {
		// Create options for continue watching prompt
		continueOptions := map[string]string{
			"y": "Continue watching",
			"n": "Search for a new show",
		}

		selectedOption, err := internal.DynamicSelect(continueOptions)
		if err != nil {
			internal.Log(fmt.Sprintf("Error selecting continue option: %v", err), logFile)
			return
		}

        if selectedOption.Key == "-1" {
            internal.ExitFlick("", nil)
        }

		if selectedOption.Key == "y" {
			// Create options map for database shows
			options := make(map[string]string)
			for _, s := range shows {
				showName, err := internal.GetShowNameFromID(s.ID)
				if err != nil {
					showName = s.ID
				}
				options[s.ID] = fmt.Sprintf("%s (Episode ID: %s)", showName, s.EpisodeID)
			}

			// Show selection menu
			selectedShow, err := internal.DynamicSelect(options)
			if err != nil {
				internal.Log(fmt.Sprintf("Error selecting show: %v", err), logFile)
				return
			}

			if selectedShow.Key == "-1" {
                internal.ExitFlick("", nil)
			}

			// Find selected show and store in show variable
			for _, s := range shows {
				if s.ID == selectedShow.Key {
					show = s
					user.Resume = true
					break
				}
			}
		}
	}

    var query string

	// If not using database, search for show
	if show.ID == "" {
        if userFlickConfig.RofiSelection {
            query, err = internal.GetUserInputFromRofi("Enter name: ")
            if err != nil {
                internal.Log(fmt.Sprintf("Error getting user input: %v", err), logFile)
                internal.ExitFlick("", err)
                return
            }
        } else {
            reader := bufio.NewReader(os.Stdin)
            fmt.Print("Enter name: ")
            query, _ = reader.ReadString('\n')
            query = strings.TrimSpace(query)
        }
		// Search and show selection code
		searchResults, err := internal.SearchShow(query)
		if err != nil {
			internal.Log(fmt.Sprintf("Error searching for show: %v", err), logFile)
			internal.ExitFlick("", err)
			return
		}

		options := make(map[string]string)
		for _, s := range searchResults {
			options[s.Id] = s.Name
		}

		selectedShow, err := internal.DynamicSelect(options)
		if err != nil || selectedShow.Key == "-1" {
			return
		}

		// Replace the directory navigation section with a recursive approach
		selectedEpisodeID, err := browseDirectory(selectedShow.Key)
		if err != nil {
			internal.Log(fmt.Sprintf("Error browsing directory: %v", err), logFile)
			internal.ExitFlick("", err)
			return
		}

		show = internal.TVShow{
			ID:           selectedShow.Key,
			EpisodeID:    selectedEpisodeID,
			PlaybackTime: 0,
		}
	}

    internal.FlickOut(fmt.Sprintf("Playing %s", show.EpisodeID))
	// Start MPV with show data
	user.Player.SocketPath, err = internal.PlayWithMPV(fmt.Sprintf("%s%s", vadapavPlaybackUrl, show.EpisodeID))
	if err != nil {
		internal.Log(fmt.Sprintf("Error starting MPV: %v", err), logFile)
		internal.ExitFlick("", err)
		return
	}

    for {

        // Get video duration
        go func() {
            for {
                if user.Player.Started {
                    if user.Player.Duration == 0 {
                        // Get video duration
                        durationPos, err := internal.MPVSendCommand(user.Player.SocketPath, []interface{}{"get_property", "duration"})
                        if err != nil {
                            internal.Log("Error getting video duration: "+err.Error(), logFile)
                        } else if durationPos != nil {
                            if duration, ok := durationPos.(float64); ok {
                                user.Player.Duration = int(duration + 0.5) // Round to nearest integer
                                internal.Log(fmt.Sprintf("Video duration: %d seconds", user.Player.Duration), logFile)
                            } else {
                                internal.Log("Error: duration is not a float64", logFile)
                            }
                        }
                        break
                    }
                }
                time.Sleep(1 * time.Second)
            }
        }()

        // Playback monitoring and database updates
        skipLoop:
        for {
            time.Sleep(1 * time.Second)

            timePos, err := internal.MPVSendCommand(user.Player.SocketPath, []interface{}{"get_property", "time-pos"})
            if err != nil {
                internal.Log("Error getting time position: "+err.Error(), logFile)
                // MPV closed or error occurred
                // Check if we reached completion percentage before starting next episode
                if user.Player.Started { 
                    percentage := internal.PercentageWatched(show.PlaybackTime, user.Player.Duration)
                    if err != nil {
                        internal.Log("Error getting percentage watched: "+err.Error(), logFile)
                    }
                    internal.Log(fmt.Sprintf("Percentage watched: %f", percentage), logFile)
                    internal.Log(fmt.Sprintf("Percentage to mark complete: %d", userFlickConfig.PercentageToMarkComplete), logFile)
                    if percentage >= float64(userFlickConfig.PercentageToMarkComplete) {
                        showDetails, err := internal.GetShow(show.ID)
                        if err != nil {
                            internal.Log(fmt.Sprintf("Error getting show details: %v", err), logFile)
                            break skipLoop
                        }

                        nextEp := internal.GetNextEpisode(showDetails, show.EpisodeID)
                        if nextEp != nil {
                            internal.FlickOut(fmt.Sprintf("Starting next episode: S%02dE%02d", nextEp.Season, nextEp.Episode))
                            show.EpisodeID = nextEp.ID
                            show.PlaybackTime = 0
                            // Remove MPV start from here and break the loop
                            break skipLoop
                        }
                        if nextEp == nil {
                            internal.FlickOut("No more episodes found")
                            internal.ExitFlick("", nil)
                        }
                    } else {
                        internal.ExitFlick("", nil)
                    }
                }
                break skipLoop  // Add this to ensure we break the loop on any MPV error
            }

            // Episode started
            if timePos != nil {
                if !user.Player.Started {
                    user.Player.Started = true
                    // Set the playback speed
                    if userFlickConfig.SaveMpvSpeed {
                        speedCmd := []interface{}{"set_property", "speed", user.Player.Speed}
                        _, err := internal.MPVSendCommand(user.Player.SocketPath, speedCmd)
                        if err != nil {
                            internal.Log("Error setting playback speed: "+err.Error(), logFile)
                        }
                    }
                }


                if user.Resume {
                    internal.SeekMPV(user.Player.SocketPath, show.PlaybackTime)
                    user.Resume = false
                }

                showPosition, ok := timePos.(float64)
                if !ok {
                    continue
                }

                // Update playback time
                show.PlaybackTime = int(showPosition + 0.5)
                user.Player.Speed, err = internal.GetMPVPlaybackSpeed(user.Player.SocketPath)
                if err != nil {
                    internal.Log(fmt.Sprintf("Error getting playback speed: %v", err), logFile)
                }
                // Save to database
                err = internal.LocalUpdateShow(databaseFile, show)
                if err != nil {
                    internal.Log(fmt.Sprintf("Error updating database: %v", err), logFile)
                }
            }
        }

        // Start the next episode after the skipLoop if we have one
        if show.PlaybackTime == 0 {  // This indicates we're ready for next episode
            var err error
            user.Player.Duration = 0  // Reset duration for new episode
            user.Player.Started = false  // Reset started flag
            user.Player.SocketPath, err = internal.PlayWithMPV(fmt.Sprintf("%s%s", vadapavPlaybackUrl, show.EpisodeID))
            if err != nil {
                internal.Log(fmt.Sprintf("Error starting next episode: %v", err), logFile)
                internal.ExitFlick("", err)
            }
        }
    }
}


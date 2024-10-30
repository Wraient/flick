package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Wraient/flick/internal"
)

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

	configFilePath := filepath.Join(homeDir, "Projects", "flick", ".config", "flick", "flick.conf")

	// load flick userFlickConfig
	userFlickConfig, err := internal.LoadConfig(configFilePath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	internal.SetGlobalConfig(&userFlickConfig)

	databaseFile := filepath.Join(os.ExpandEnv(userFlickConfig.StoragePath), "shows.db")
	logFile := "debug.log"

	internal.ClearLog(logFile)
	// Get all shows from database
	shows := internal.LocalGetAllShows(databaseFile)
	
	if len(shows) > 0 {
		fmt.Println("\nWould you like to continue watching a show? (y/n)")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" {
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
				fmt.Println("Error:", err)
				return
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

	// If not using database, search for show
	if show.ID == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter show name: ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)

		// Search and show selection code
		searchResults, err := internal.SearchShow(query)
		if err != nil {
			fmt.Println("Error:", err)
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

		// Get episodes for selected show
		showDetails, err := internal.GetShow(selectedShow.Key)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Episode selection
		episodeOptions := make(map[string]string)
		for _, episode := range showDetails.EpisodesList {
			key := fmt.Sprintf("%d_%d", episode.Season, episode.Episode)
			episodeOptions[key] = fmt.Sprintf("S%02dE%02d: %s", episode.Season, episode.Episode, episode.Name)
		}

		selectedEp, err := internal.DynamicSelect(episodeOptions)
		if err != nil || selectedEp.Key == "-1" {
			return
		}

		// Parse selected episode
		parts := strings.Split(selectedEp.Key, "_")
		episode, _ := strconv.Atoi(parts[1])

		// Store in show variable
		show = internal.TVShow{
			ID:           selectedShow.Key,
			EpisodeID:    showDetails.EpisodesList[episode-1].ID,
			PlaybackTime: 0,
		}
	}

	// Start MPV with show data
	user.Player.SocketPath, err = internal.PlayWithMPV(fmt.Sprintf("%s%s", vadapavPlaybackUrl, show.EpisodeID))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Playing with MPV at socket path:", user.Player.SocketPath)

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
	for {
		time.Sleep(1 * time.Second)

		timePos, err := internal.MPVSendCommand(user.Player.SocketPath, []interface{}{"get_property", "time-pos"})
		if err != nil {
            internal.Log("Error getting time position: "+err.Error(), logFile)
			// MPV closed or error occurred
			// Check if we reached completion percentage before starting next episode
            if user.Player.Started { 
                percentage := internal.PercentageWatched(show.PlaybackTime, user.Player.Duration)
                fmt.Println(percentage)
                if err != nil {
                    internal.Log("Error getting percentage watched: "+err.Error(), logFile)
                }
                internal.Log(fmt.Sprintf("Percentage watched: %f", percentage), logFile)
                internal.Log(fmt.Sprintf("Percentage to mark complete: %d", userFlickConfig.PercentageToMarkComplete), logFile)
				if percentage >= float64(userFlickConfig.PercentageToMarkComplete) {
                    showDetails, err := internal.GetShow(show.ID)
                    if err != nil {
                        fmt.Println("Error getting show details:", err)
                        break
                    }

                    nextEp := internal.GetNextEpisode(showDetails, show.EpisodeID)
                    if nextEp != nil {
                        fmt.Printf("\nStarting next episode: S%02dE%02d\n", nextEp.Season, nextEp.Episode)
                        show.EpisodeID = nextEp.ID
                        show.PlaybackTime = 0
                        user.Player.SocketPath, err = internal.PlayWithMPV(fmt.Sprintf("%s%s", vadapavPlaybackUrl, nextEp.ID))
                        if err != nil {
                            fmt.Println("Error starting next episode:", err)
                        }
						continue
					}
				} else {
					internal.FlickOut("Have a great day.")
                    break
				}
			}
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

			// Save to database
			err = internal.LocalUpdateShow(databaseFile, show)
			if err != nil {
				fmt.Println("Error updating database:", err)
			}
		}
	}
}

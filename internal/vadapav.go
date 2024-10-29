package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type MediaItem struct {
	Name      string
	Link      string
	IsDir     bool
	Quality   string
	Type      string // "movie" or "show"
}

type SearchResult struct {
	Movies map[string][]MediaItem
	Shows  map[string][]MediaItem
}

type VadapavFile struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Dir    bool      `json:"dir"`
	Parent string    `json:"parent"`
	MTime  time.Time `json:"mtime"`
}

type DirectoryResponse struct {
	Message string `json:"message"`
	Data    struct {
		ID    string        `json:"id"`
		Name  string        `json:"name"`
		Parent string        `json:"parent"`
		Dir   bool          `json:"dir"`
		MTime time.Time     `json:"mtime"`
		Files []VadapavFile `json:"files"`
	} `json:"data"`
}

type SearchResponse struct {
	Message string        `json:"message"`
	Data    []VadapavFile `json:"data"`
}

const apiBaseURL = "https://vadapav.mov/api"

func BrowseAndPlay(dirID string) error {
	for {
		// Fetch directory contents
		dirURL := fmt.Sprintf("%s/d/%s", apiBaseURL, dirID)
		
		var dirResp DirectoryResponse
		if err := fetchJSON(dirURL, &dirResp); err != nil {
			return fmt.Errorf("failed to fetch directory: %w", err)
		}

		// Convert entries to selection options
		options := make(map[string]string)
		
		// Add back option if not at root and has parent
		if dirResp.Data.Parent != "" {
			options[".."] = "← Back"
		}

		// Add all entries
		for _, file := range dirResp.Data.Files {
			displayName := file.Name
			if !file.Dir {
				displayName = "▶ " + displayName // Add play symbol for files
			}
			options[file.ID] = displayName
		}

		// Show selection menu
		selected, err := DynamicSelect(options, false)
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}

		// Handle quit
		if selected.Key == "-1" {
			return nil
		}

		// Handle back
		if selected.Key == ".." {
			dirID = dirResp.Data.Parent
			continue
		}

		// Find selected file
		var selectedFile VadapavFile
		for _, file := range dirResp.Data.Files {
			if file.ID == selected.Key {
				selectedFile = file
				break
			}
		}

		// If it's a directory, navigate into it
		if selectedFile.Dir {
			dirID = selectedFile.ID
			continue
		}

		// If it's a file, play it with MPV
		playURL := fmt.Sprintf("https://vadapav.mov/f/%s", selectedFile.ID)
		err = playWithMPV(playURL)
		if err != nil {
			return fmt.Errorf("failed to play media: %w", err)
		}
		return nil
	}
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check rate limiting headers
	remaining := resp.Header.Get("x-ratelimit-remaining")
	if remaining == "0" {
		reset := resp.Header.Get("x-ratelimit-reset")
		return fmt.Errorf("rate limit exceeded, resets at: %s", reset)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func playWithMPV(url string) error {
	cmd := exec.Command("mpv", url)
	return cmd.Run()
}

func SearchMedia(query string) error {
	// Search for the media
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s/s/%s", apiBaseURL, encodedQuery)

	fmt.Printf("Searching for: %s\n", searchURL)
	
	var searchResp SearchResponse
	if err := fetchJSON(searchURL, &searchResp); err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Convert search results to selection options
	options := make(map[string]string)
	for _, file := range searchResp.Data {
		mediaType := "Movie"
		if file.Dir {
			mediaType = "Show"
		}
		options[file.ID] = fmt.Sprintf("[%s] %s", mediaType, file.Name)
	}

	// Show selection menu
	selected, err := DynamicSelect(options, false)
	if err != nil {
		return fmt.Errorf("selection failed: %w", err)
	}

	// Handle quit
	if selected.Key == "-1" {
		return nil
	}

	// Browse the selected item
	return BrowseAndPlay(selected.Key)
}

// Helper function to clean up titles
func cleanTitle(title string) string {
	// Remove quality indicators
	title = strings.ReplaceAll(strings.ToLower(title), "1080p", "")
	title = strings.ReplaceAll(title, "720p", "")
	title = strings.ReplaceAll(title, "2160p", "")
	
	// Remove common brackets content
	title = regexp.MustCompile(`\[.*?\]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(title, "")
	
	// Remove season/episode indicators
	title = regexp.MustCompile(`s\d{2}e\d{2}`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`season \d+`).ReplaceAllString(title, "")
	
	// Clean up remaining artifacts
	title = strings.TrimSpace(title)
	title = strings.Trim(title, ".-_ ")
	
	return strings.Title(strings.ToLower(title))
}



package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Simplified struct without LastWatched
type TVShow struct {
	ID           string `json:"id"`           // Vadapav show directory ID
	PlaybackTime int    `json:"playback_time"`// Current playback time
	EpisodeID    string `json:"episode_id"`   // Current episode ID
}

// Function to add or update a TV show entry
func LocalUpdateShow(databaseFile string, show TVShow) error {
	shows := LocalGetAllShows(os.ExpandEnv(databaseFile))
	
	// Find and update existing entry or add new one
	updated := false
	for i, s := range shows {
		if s.ID == show.ID {
			shows[i] = show
			updated = true
			break
		}
	}

	if !updated {
		shows = append(shows, show)
	}

	// Write updated list back to file
	file, err := os.Create(databaseFile)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if file is new
	if !updated && len(shows) == 1 {
		header := []string{"ShowID", "EpisodeID", "PlaybackTime"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}
	}

	// Write all shows without LastWatched
	for _, s := range shows {
		record := []string{
			s.ID,
			s.EpisodeID,
			strconv.Itoa(s.PlaybackTime),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
	}
	
	return nil
}

// Function to get all TV shows from the database
func LocalGetAllShows(databaseFile string) []TVShow {
	var shows []TVShow

	// Ensure the directory exists
	dir := filepath.Dir(databaseFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		OctoOut(fmt.Sprintf("Error creating directory: %v", err))
		return shows
	}

	// Open the file, create if it doesn't exist
	file, err := os.OpenFile(databaseFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		OctoOut(fmt.Sprintf("Error opening or creating file: %v", err))
		return shows
	}
	defer file.Close()

	// If the file was just created, return empty list
	fileInfo, err := file.Stat()
	if err != nil {
		OctoOut(fmt.Sprintf("Error getting file info: %v", err))
		return shows
	}
	if fileInfo.Size() == 0 {
		return shows
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		OctoOut(fmt.Sprintf("Error reading file: %v", err))
		return shows
	}

	// Skip header row if it exists
	startIdx := 0
	if len(records) > 0 && records[0][0] == "ShowID" {
		startIdx = 1
	}

	// Parse records
	for _, row := range records[startIdx:] {
		show := parseShowRow(row)
		if show != nil {
			shows = append(shows, *show)
		}
	}

	return shows
}

// Function to parse a single row of show data
func parseShowRow(row []string) *TVShow {
	if len(row) < 3 {
		OctoOut(fmt.Sprintf("Invalid row format: %v", row))
		return nil
	}

	playbackTime, _ := strconv.Atoi(row[2])

	return &TVShow{
		ID:           row[0],
		EpisodeID:    row[1],
		PlaybackTime: playbackTime,
	}
}

// Function to find a show by ID
func LocalFindShow(shows []TVShow, showID string) *TVShow {
	for _, show := range shows {
		if show.ID == showID {
			return &show
		}
	}
	return nil
}

// Function to update show progress
func UpdateShowProgress(databaseFile string, showID string, episodeID string, playbackTime int) error {
	shows := LocalGetAllShows(databaseFile)
	show := LocalFindShow(shows, showID)
	
	if show == nil {
		// New show
		show = &TVShow{
			ID: showID,
		}
	}
	
	show.EpisodeID = episodeID
	show.PlaybackTime = playbackTime
	
	return LocalUpdateShow(databaseFile, *show)
}

// Function to delete a show by ID
func LocalDeleteShow(databaseFile string, showID string) error {
	shows := LocalGetAllShows(databaseFile)
	
	// Filter out the show with matching ID
	var filteredShows []TVShow
	found := false
	for _, show := range shows {
		if show.ID != showID {
			filteredShows = append(filteredShows, show)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("show with ID %s not found", showID)
	}

	// Write the filtered list back to file
	file, err := os.Create(databaseFile)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ShowID", "EpisodeID", "PlaybackTime"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write remaining shows
	for _, show := range filteredShows {
		record := []string{
			show.ID,
			show.EpisodeID,
			strconv.Itoa(show.PlaybackTime),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
	}

	return nil
}

// Function to delete multiple shows by IDs
func LocalDeleteShows(databaseFile string, showIDs []string) error {
	shows := LocalGetAllShows(databaseFile)
	
	// Create a map of IDs to delete for O(1) lookup
	toDelete := make(map[string]bool)
	for _, id := range showIDs {
		toDelete[id] = true
	}

	// Filter out shows with matching IDs
	var filteredShows []TVShow
	for _, show := range shows {
		if !toDelete[show.ID] {
			filteredShows = append(filteredShows, show)
		}
	}

	// Write the filtered list back to file
	file, err := os.Create(databaseFile)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ShowID", "EpisodeID", "PlaybackTime"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write remaining shows
	for _, show := range filteredShows {
		record := []string{
			show.ID,
			show.EpisodeID,
			strconv.Itoa(show.PlaybackTime),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
	}

	return nil
}

// Function to clear all shows from the database
func LocalClearShows(databaseFile string) error {
	// Create empty file (clearing all content)
	file, err := os.Create(databaseFile)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write only the header
	header := []string{"ShowID", "EpisodeID", "PlaybackTime"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	return nil
}

// Function to get show name from ID
func GetShowNameFromID(showID string) (string, error) {
	show, err := GetShow(showID)
	if err != nil {
		return "", fmt.Errorf("error getting show name: %w", err)
	}
	return show.Name, nil
}

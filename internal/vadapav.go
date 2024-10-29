package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

func BrowseAndPlay(initialURL string) error {
	currentURL := initialURL
	for {
		// Fetch and parse the current directory
		entries, err := fetchDirectoryContents(currentURL)
		if err != nil {
			return fmt.Errorf("failed to fetch directory: %w", err)
		}

		// Convert entries to selection options
		options := make(map[string]string)
		mediaItems := make(map[string]MediaItem)
		
		// Add back option if not at root
		if currentURL != initialURL {
			options[".."] = "← Back"
			mediaItems[".."] = MediaItem{
				Name:  "Back",
				Link:  getParentURL(currentURL),
				IsDir: true,
			}
		}

		// Add all entries (both directories and files)
		for _, entry := range entries {
			displayName := entry.Name
			options[entry.Link] = displayName
			mediaItems[entry.Link] = entry
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

		// Get the selected media item
		selectedItem := mediaItems[selected.Key]

		// If it's a directory, navigate into it
		if selectedItem.IsDir {
			ClearScreen()
			currentURL = selectedItem.Link
			continue
		}

		// If it's a file, play it with MPV
		err = playWithMPV(selectedItem.Link)
		if err != nil {
			return fmt.Errorf("failed to play media: %w", err)
		}
		return nil
	}
}

func fetchDirectoryContents(url string) ([]MediaItem, error) {
	doc, err := fetchAndParse(url)
	if err != nil {
		return nil, err
	}

	var items []MediaItem

	// Find directory entries
	doc.Find(".directory-entry").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		fullLink := "https://vadapav.mov" + link
		
		quality := "unknown"
		if strings.Contains(strings.ToLower(name), "1080p") {
			quality = "1080p"
		} else if strings.Contains(strings.ToLower(name), "720p") {
			quality = "720p"
		} else if strings.Contains(strings.ToLower(name), "2160p") {
			quality = "4K"
		}

		mediaType := "movie"
		if strings.Contains(strings.ToLower(name), "season") || 
		   strings.Contains(strings.ToLower(name), "episode") {
			mediaType = "show"
		}

		items = append(items, MediaItem{
			Name:    name,
			Link:    fullLink,
			IsDir:   true,
			Quality: quality,
			Type:    mediaType,
		})
	})

	// Find file entries
	doc.Find(".file-entry").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		fullLink := "https://vadapav.mov" + link
		
		quality := "unknown"
		if strings.Contains(strings.ToLower(name), "1080p") {
			quality = "1080p"
		} else if strings.Contains(strings.ToLower(name), "720p") {
			quality = "720p"
		} else if strings.Contains(strings.ToLower(name), "2160p") {
			quality = "4K"
		}

		mediaType := "movie"
		if strings.Contains(strings.ToLower(name), "season") || 
		   strings.Contains(strings.ToLower(name), "episode") {
			mediaType = "show"
		}

		items = append(items, MediaItem{
			Name:    name,
			Link:    fullLink,
			IsDir:   false,  // This is a file
			Quality: quality,
			Type:    mediaType,
		})
	})

	return items, nil
}

func fetchAndParse(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

func playWithMPV(url string) error {
	cmd := exec.Command("mpv", url)
	return cmd.Run()
}

func getParentURL(url string) string {
	// Remove trailing slash if present
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	
	// Find the last slash
	lastSlash := strings.LastIndex(url, "/")
	if lastSlash == -1 {
		return url
	}
	
	return url[:lastSlash+1]
}

func SearchMedia(query string) (*SearchResult, error) {
	baseURL := "https://vadapav.mov/s/"
	encodedQuery := url.PathEscape(query)
	searchURL := baseURL + encodedQuery

	// Make the HTTP request
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	results := &SearchResult{
		Movies: make(map[string][]MediaItem),
		Shows:  make(map[string][]MediaItem),
	}

	// Find all directory entries
	doc.Find(".directory-entry").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		fullLink := "https://vadapav.mov" + link
		isDir := strings.HasSuffix(link, "/")

		// Parse quality
		quality := "unknown"
		if strings.Contains(strings.ToLower(name), "1080p") {
			quality = "1080p"
		} else if strings.Contains(strings.ToLower(name), "720p") {
			quality = "720p"
		} else if strings.Contains(strings.ToLower(name), "2160p") {
			quality = "4K"
		}

		// Create MediaItem
		item := MediaItem{
			Name:    name,
			Link:    fullLink,
			IsDir:   isDir,
			Quality: quality,
		}

		// Determine if it's a show or movie
		isShow := strings.Contains(strings.ToLower(name), "season") ||
			strings.Contains(strings.ToLower(name), "episode") ||
			strings.Contains(strings.ToLower(name), "s01") ||
			strings.Contains(strings.ToLower(name), "complete")

		// Clean the title to get the base name
		baseName := cleanTitle(name)

		if isShow {
			item.Type = "show"
			if _, exists := results.Shows[baseName]; !exists {
				results.Shows[baseName] = []MediaItem{}
			}
			results.Shows[baseName] = append(results.Shows[baseName], item)
		} else {
			item.Type = "movie"
			if _, exists := results.Movies[baseName]; !exists {
				results.Movies[baseName] = []MediaItem{}
			}
			results.Movies[baseName] = append(results.Movies[baseName], item)
		}
	})

	// Create selection options from search results
	options := make(map[string]string)
	mediaItems := make(map[string]MediaItem)

	// Add movies
	for title, items := range results.Movies {
		for _, item := range items {
			displayName := fmt.Sprintf("[Movie] %s (%s)", title, item.Quality)
			options[item.Link] = displayName
			mediaItems[item.Link] = item
		}
	}

	// Add shows
	for title, items := range results.Shows {
		for _, item := range items {
			displayName := fmt.Sprintf("[Show] %s (%s)", title, item.Quality)
			options[item.Link] = displayName
			mediaItems[item.Link] = item
		}
	}

	// Show selection menu
	selected, err := DynamicSelect(options, false)
	if err != nil {
		return nil, fmt.Errorf("selection failed: %w", err)
	}

	// Handle quit
	if selected.Key == "-1" {
		return nil, nil
	}

	// Browse and play the selected item
	err = BrowseAndPlay(selected.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to browse and play: %w", err)
	}

	return results, nil
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



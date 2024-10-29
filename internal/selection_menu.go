package internal

import (
	"fmt"
	"sort"
	"strings"
	"github.com/charmbracelet/bubbletea"
)

// SelectionOption holds the label and the internal key
type SelectionOption struct {
	Label string
	Key   string
	Type  string // "movie" or "show"
	Quality string // e.g., "1080p", "720p"
}

// Model represents the application state for the selection prompt
type Model struct {
	options        map[string]string
	filter         string
	filteredKeys   []SelectionOption
	selected       int
	terminalWidth  int
	terminalHeight int
	scrollOffset   int
	showAddNew     bool
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.terminalWidth = wsm.Width
		m.terminalHeight = wsm.Height
	}

	updateFilter := false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.filteredKeys[m.selected] = SelectionOption{Label: "quit", Key: "-1"}
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				updateFilter = true
			}
		case "down":
			if m.selected < len(m.filteredKeys)-1 {
				m.selected++
			}
			if m.selected >= m.scrollOffset+m.visibleItemsCount() {
				m.scrollOffset++
			}
		case "up":
			if m.selected > 0 {
				m.selected--
			}
			if m.selected < m.scrollOffset {
				m.scrollOffset--
			}
		default:
			if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
				m.filter += msg.String()
				updateFilter = true
			}
		}
	}

	if updateFilter {
		m.filterOptions()
		m.selected = 0
		m.scrollOffset = 0
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	var b strings.Builder

	b.WriteString("Search (Press Ctrl+C to quit):\n")
	b.WriteString("Filter: " + m.filter + "\n\n")

	if len(m.filteredKeys) == 0 {
		b.WriteString("No matches found.\n")
	} else {
		visibleItems := m.visibleItemsCount()
		start := m.scrollOffset
		end := start + visibleItems
		if end > len(m.filteredKeys) {
			end = len(m.filteredKeys)
		}

		for i := start; i < end; i++ {
			entry := m.filteredKeys[i]
			prefix := "  "
			if i == m.selected {
				prefix = "▶ "
			}

			// Format the entry based on type and quality
			label := entry.Label
			if entry.Type != "" && entry.Quality != "" {
				label = fmt.Sprintf("%s (%s, %s)", entry.Label, entry.Type, entry.Quality)
			}

			b.WriteString(fmt.Sprintf("%s%s\n", prefix, label))
		}
	}

	return b.String()
}

func (m Model) visibleItemsCount() int {
	return m.terminalHeight - 4
}

func (m *Model) filterOptions() {
	m.filteredKeys = []SelectionOption{}

	for key, value := range m.options {
		if strings.Contains(strings.ToLower(value), strings.ToLower(m.filter)) {
			// Try to detect if it's a movie or show and its quality
			entryType := "movie"
			if strings.Contains(strings.ToLower(value), "season") || 
			   strings.Contains(strings.ToLower(value), "episode") {
				entryType = "show"
			}

			quality := "unknown"
			if strings.Contains(strings.ToLower(value), "1080p") {
				quality = "1080p"
			} else if strings.Contains(strings.ToLower(value), "720p") {
				quality = "720p"
			} else if strings.Contains(strings.ToLower(value), "2160p") {
				quality = "4K"
			}

			m.filteredKeys = append(m.filteredKeys, SelectionOption{
				Label:   value,
				Key:     key,
				Type:    entryType,
				Quality: quality,
			})
		}
	}

	sort.Slice(m.filteredKeys, func(i, j int) bool {
		// Sort by type first (movies before shows)
		if m.filteredKeys[i].Type != m.filteredKeys[j].Type {
			return m.filteredKeys[i].Type < m.filteredKeys[j].Type
		}
		// Then by label
		return m.filteredKeys[i].Label < m.filteredKeys[j].Label
	})

	if m.showAddNew {
		m.filteredKeys = append(m.filteredKeys, SelectionOption{
			Label: "Add new media",
			Key:   "add_new",
		})
	}

	m.filteredKeys = append(m.filteredKeys, SelectionOption{
		Label: "Quit",
		Key:   "-1",
	})
}

func DynamicSelect(options map[string]string, showAddNew bool) (SelectionOption, error) {
	model := &Model{
		options:    options,
		filteredKeys: make([]SelectionOption, 0),
		showAddNew: showAddNew,
	}

	model.filterOptions()
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return SelectionOption{}, err
	}

	finalSelectionModel, ok := finalModel.(*Model)
	if !ok {
		return SelectionOption{}, fmt.Errorf("unexpected model type")
	}

	if finalSelectionModel.selected < len(finalSelectionModel.filteredKeys) {
		return finalSelectionModel.filteredKeys[finalSelectionModel.selected], nil
	}
	return SelectionOption{}, nil
}

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pretty-output/formatter"
	"pretty-output/parser"
	"pretty-output/store"
)

// Focus states
type focus int

const (
	focusList focus = iota
	focusLogs
	focusFilter
)

// Styles
var (
	// Left panel styles
	containerListStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	containerListFocusedStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("212")).
					Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)

	// Right panel styles
	logPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	logPanelFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("212")).
				Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	filterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	filterInputStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
)

// NewEntryMsg signals a new log entry was added
type NewEntryMsg struct {
	Entry parser.Entry
}

// Model is the main TUI model
type Model struct {
	store             *store.Store
	selectedIndex     int
	selectedContainer string
	viewport          viewport.Model
	width             int
	height            int
	ready             bool
	focus             focus
	filter            string
}

// New creates a new Model
func New(s *store.Store) Model {
	return Model{
		store:             s,
		selectedIndex:     0,
		selectedContainer: "",
		focus:             focusList,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle filter mode input
		if m.focus == focusFilter {
			switch msg.String() {
			case "esc":
				// Clear filter and go back to logs
				m.filter = ""
				m.focus = focusLogs
				m = m.updateViewportContent()
			case "enter":
				// Confirm filter and go back to logs
				m.focus = focusLogs
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m = m.updateViewportContent()
				}
			case "ctrl+c":
				return m, tea.Quit
			default:
				// Add character to filter
				if len(msg.String()) == 1 {
					m.filter += msg.String()
					m = m.updateViewportContent()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			// Toggle focus between panels
			if m.focus == focusList {
				m.focus = focusLogs
			} else {
				m.focus = focusList
			}
		case "enter", "right", "l":
			// Enter log view
			if m.focus == focusList {
				m.focus = focusLogs
			}
		case "left", "h":
			// Back to container list
			if m.focus == focusLogs {
				m.focus = focusList
			}
		case "esc":
			// Back to container list or clear filter
			if m.focus == focusLogs {
				if m.filter != "" {
					m.filter = ""
					m = m.updateViewportContent()
				} else {
					m.focus = focusList
				}
			}
		case "/":
			// Enter filter mode
			if m.focus == focusLogs {
				m.focus = focusFilter
			}
		case "up", "k":
			if m.focus == focusList {
				if m.selectedIndex > 0 {
					m.selectedIndex--
					m = m.updateSelectedContainer()
				}
			} else {
				m.viewport.LineUp(1)
			}
		case "down", "j":
			if m.focus == focusList {
				containers := m.store.Containers()
				if m.selectedIndex < len(containers)-1 {
					m.selectedIndex++
					m = m.updateSelectedContainer()
				}
			} else {
				m.viewport.LineDown(1)
			}
		case "pgup":
			m.viewport.LineUp(m.viewport.Height / 2)
		case "pgdown":
			m.viewport.LineDown(m.viewport.Height / 2)
		case "home", "g":
			if m.focus == focusLogs {
				m.viewport.GotoTop()
			}
		case "end", "G":
			if m.focus == focusLogs {
				m.viewport.GotoBottom()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate panel widths
		leftWidth := 25
		rightWidth := m.width - leftWidth - 4 // account for borders

		if !m.ready {
			m.viewport = viewport.New(rightWidth-2, m.height-4)
			m.viewport.Style = lipgloss.NewStyle()
			m.ready = true
		} else {
			m.viewport.Width = rightWidth - 2
			m.viewport.Height = m.height - 4
		}

		m = m.updateViewportContent()

	case NewEntryMsg:
		m.store.Add(msg.Entry)

		// Auto-select first container
		if m.selectedContainer == "" {
			containers := m.store.Containers()
			if len(containers) > 0 {
				m.selectedContainer = containers[0]
			}
		}

		// Update viewport if this entry belongs to selected container
		if msg.Entry.Container == m.selectedContainer {
			m = m.updateViewportContent()
			// Auto-scroll to bottom for new entries
			m.viewport.GotoBottom()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) updateSelectedContainer() Model {
	containers := m.store.Containers()
	if m.selectedIndex < len(containers) {
		m.selectedContainer = containers[m.selectedIndex]
		m.filter = "" // Clear filter when changing containers
		m = m.updateViewportContent()
		m.viewport.GotoBottom()
	}
	return m
}

func (m Model) updateViewportContent() Model {
	if m.selectedContainer == "" {
		m.viewport.SetContent("No container selected")
		return m
	}

	entries := m.store.Entries(m.selectedContainer)
	var lines []string
	filterLower := strings.ToLower(m.filter)

	for _, entry := range entries {
		formatted := formatter.Format(entry)
		// Apply filter if set
		if m.filter == "" || strings.Contains(strings.ToLower(entry.Content), filterLower) {
			lines = append(lines, formatted)
		}
	}

	if len(lines) == 0 && m.filter != "" {
		lines = append(lines, helpStyle.Render("No matches for: "+m.filter))
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
	return m
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Left panel: container list
	leftWidth := 25
	rightWidth := m.width - leftWidth - 4

	containers := m.store.Containers()
	var containerList strings.Builder
	containerList.WriteString(titleStyle.Render("Containers"))
	containerList.WriteString("\n\n")

	for i, container := range containers {
		count := m.store.EntryCount(container)
		line := fmt.Sprintf("%s (%d)", container, count)

		if i == m.selectedIndex {
			containerList.WriteString(selectedStyle.Render(line))
		} else {
			containerList.WriteString(normalStyle.Render(line))
		}
		containerList.WriteString("\n")
	}

	if len(containers) == 0 {
		containerList.WriteString(helpStyle.Render("Waiting for input..."))
	}

	// Apply focus style to left panel
	leftPanelStyle := containerListStyle
	if m.focus == focusList {
		leftPanelStyle = containerListFocusedStyle
	}
	leftPanel := leftPanelStyle.
		Width(leftWidth).
		Height(m.height - 2).
		Render(containerList.String())

	// Right panel: log output
	title := titleStyle.Render(fmt.Sprintf("Logs: %s", m.selectedContainer))
	scrollInfo := helpStyle.Render(fmt.Sprintf(" %d%%", int(m.viewport.ScrollPercent()*100)))

	// Show filter if active
	var filterLine string
	if m.focus == focusFilter {
		filterLine = "\n" + filterStyle.Render("/") + filterInputStyle.Render(m.filter+"▏")
	} else if m.filter != "" {
		filterLine = "\n" + filterStyle.Render("filter: ") + helpStyle.Render(m.filter)
	}

	rightContent := title + scrollInfo + filterLine + "\n" + m.viewport.View()

	// Apply focus style to right panel
	rightPanelStyle := logPanelStyle
	if m.focus == focusLogs {
		rightPanelStyle = logPanelFocusedStyle
	}
	rightPanel := rightPanelStyle.
		Width(rightWidth).
		Height(m.height - 2).
		Render(rightContent)

	// Help bar - context sensitive
	var help string
	switch m.focus {
	case focusList:
		help = helpStyle.Render("↑/↓: select • enter/→: view logs • tab: switch • q: quit")
	case focusLogs:
		help = helpStyle.Render("↑/↓: scroll • /: filter • ←/esc: back • g/G: top/bottom • q: quit")
	case focusFilter:
		help = helpStyle.Render("type to filter • enter: confirm • esc: clear & back")
	}

	// Combine panels
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	return panels + "\n" + help
}

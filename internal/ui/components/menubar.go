package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/ui/theme"
)

// MenuItem represents a single menu in the menu bar
type MenuItem struct {
	Label   string // Display text (e.g., "File")
	Hotkey  rune   // Alt+key (e.g., 'F' for Alt+F)
	Active  bool   // Is this menu currently open?
	Items   []string // Submenu items (for future dropdown implementation)
}

// MenuBar represents the top menu bar
type MenuBar struct {
	items        []MenuItem
	activeIndex  int  // Which menu is highlighted (-1 = none)
	menuActive   bool // Is the menu bar active (F10 or Alt pressed)?
	width        int
}

// NewMenuBar creates a new TP7-style menu bar
func NewMenuBar() MenuBar {
	return MenuBar{
		items: []MenuItem{
			{
				Label:  "File",
				Hotkey: 'f',
				Items:  []string{"Open", "Export Patterns", "Preferences", "Exit"},
			},
			{
				Label:  "View",
				Hotkey: 'v',
				Items:  []string{"Dashboard", "Transactions", "Accounts", "Reports"},
			},
			{
				Label:  "Reports",
				Hotkey: 'r',
				Items:  []string{"Monthly", "Yearly", "By Category", "Export"},
			},
			{
				Label:  "Help",
				Hotkey: 'h',
				Items:  []string{"Keyboard Shortcuts", "About Lima"},
			},
		},
		activeIndex: -1,
		menuActive:  false,
	}
}

// Update handles messages for the menu bar
func (m MenuBar) Update(msg tea.Msg) (MenuBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// F10 toggles menu activation
		if msg.String() == "f10" {
			m.menuActive = !m.menuActive
			if m.menuActive && m.activeIndex == -1 {
				m.activeIndex = 0 // Activate first menu
			} else if !m.menuActive {
				m.activeIndex = -1 // Deactivate
			}
			return m, nil
		}

		// Escape deactivates menu
		if msg.String() == "esc" && m.menuActive {
			m.menuActive = false
			m.activeIndex = -1
			return m, nil
		}

		// Alt+key activates specific menu
		if strings.HasPrefix(msg.String(), "alt+") {
			key := rune(msg.String()[4]) // Get character after "alt+"
			for i, item := range m.items {
				if item.Hotkey == key {
					m.menuActive = true
					m.activeIndex = i
					return m, nil
				}
			}
		}

		// If menu is active, handle arrow keys
		if m.menuActive {
			switch msg.String() {
			case "left":
				m.activeIndex--
				if m.activeIndex < 0 {
					m.activeIndex = len(m.items) - 1
				}
				return m, nil
			case "right":
				m.activeIndex++
				if m.activeIndex >= len(m.items) {
					m.activeIndex = 0
				}
				return m, nil
			}
		}
	}

	return m, nil
}

// View renders the menu bar
func (m MenuBar) View() string {
	var parts []string

	// Add "Lima" application name
	appName := theme.MenuBarStyle.Render(" Lima ")
	parts = append(parts, appName)

	// Render each menu item
	for i, item := range m.items {
		// Separate menus with spaces
		parts = append(parts, theme.MenuBarStyle.Render(" "))

		// Render the menu label with hotkey underlined
		menuText := renderMenuWithHotkey(item.Label, item.Hotkey, i == m.activeIndex)
		parts = append(parts, menuText)
	}

	// Fill remaining space with menu bar background
	rendered := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	renderedWidth := lipgloss.Width(rendered)
	if m.width > renderedWidth {
		padding := theme.MenuBarStyle.Render(strings.Repeat(" ", m.width-renderedWidth))
		rendered = lipgloss.JoinHorizontal(lipgloss.Top, rendered, padding)
	}

	return rendered
}

// IsActive returns whether the menu bar is currently active
func (m MenuBar) IsActive() bool {
	return m.menuActive
}

// Deactivate deactivates the menu bar
func (m MenuBar) Deactivate() MenuBar {
	m.menuActive = false
	m.activeIndex = -1
	return m
}

// SetWidth sets the menu bar width
func (m MenuBar) SetWidth(width int) MenuBar {
	m.width = width
	return m
}

// renderMenuWithHotkey renders a menu item with its hotkey highlighted
func renderMenuWithHotkey(label string, hotkey rune, active bool) string {
	var result strings.Builder

	// Find the hotkey position in the label
	hotkeyIndex := -1
	labelLower := strings.ToLower(label)
	hotkeyLower := strings.ToLower(string(hotkey))

	for i, ch := range labelLower {
		if string(ch) == hotkeyLower {
			hotkeyIndex = i
			break
		}
	}

	// Build the styled string
	baseStyle := theme.MenuBarStyle
	if active {
		baseStyle = theme.MenuItemActiveStyle
	}

	for i, ch := range label {
		if i == hotkeyIndex {
			// Render hotkey with special style
			hotkeyStyle := theme.MenuHotkeyStyle
			if active {
				// When menu is active, still show yellow on the inverted background
				hotkeyStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.TP7Yellow)).
					Background(lipgloss.Color(theme.TP7Cyan)).
					Underline(true)
			} else {
				hotkeyStyle = hotkeyStyle.Underline(true)
			}
			result.WriteString(hotkeyStyle.Render(string(ch)))
		} else {
			result.WriteString(baseStyle.Render(string(ch)))
		}
	}

	return result.String()
}

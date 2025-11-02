package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/ui/theme"
)

// StatusBarItem represents a single F-key shortcut in the status bar
type StatusBarItem struct {
	Key   string // e.g., "F1", "F2"
	Label string // e.g., "Help", "Dashboard"
}

// StatusBar represents the bottom status bar with F-key hints
type StatusBar struct {
	items []StatusBarItem
	width int
}

// NewStatusBar creates a new TP7-style status bar
func NewStatusBar() StatusBar {
	return StatusBar{
		items: []StatusBarItem{
			{Key: "F1", Label: "Help"},
			{Key: "F2", Label: "Dashboard"},
			{Key: "F3", Label: "Trans"},
			{Key: "F4", Label: "Accounts"},
			{Key: "F5", Label: "Reports"},
			{Key: "F10", Label: "Menu"},
		},
		width: 80,
	}
}

// NewContextStatusBar creates a status bar with context-specific items
func NewContextStatusBar(items []StatusBarItem) StatusBar {
	return StatusBar{
		items: items,
		width: 80,
	}
}

// SetWidth sets the status bar width
func (s StatusBar) SetWidth(width int) StatusBar {
	s.width = width
	return s
}

// SetItems updates the status bar items (for context-sensitive display)
func (s StatusBar) SetItems(items []StatusBarItem) StatusBar {
	s.items = items
	return s
}

// View renders the status bar
func (s StatusBar) View() string {
	var parts []string

	for i, item := range s.items {
		// Render F-key in regular status bar style
		keyPart := theme.StatusBarStyle.Render(item.Key)

		// Render label in status bar style
		labelPart := theme.StatusBarStyle.Render(item.Label)

		// Combine key and label
		combined := keyPart + theme.StatusBarStyle.Render(" ") + labelPart

		parts = append(parts, combined)

		// Add separator between items (except last)
		if i < len(s.items)-1 {
			parts = append(parts, theme.StatusBarStyle.Render("  "))
		}
	}

	// Join all parts
	rendered := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	// Fill remaining space with status bar background
	renderedWidth := lipgloss.Width(rendered)
	if s.width > renderedWidth {
		padding := theme.StatusBarStyle.Render(strings.Repeat(" ", s.width-renderedWidth))
		rendered = lipgloss.JoinHorizontal(lipgloss.Top, rendered, padding)
	}

	return rendered
}

// Common status bar configurations for different views

// DashboardStatusBar returns status bar items for dashboard view
func DashboardStatusBar() []StatusBarItem {
	return []StatusBarItem{
		{Key: "F1", Label: "Help"},
		{Key: "F2", Label: "Dashboard"},
		{Key: "F3", Label: "Trans"},
		{Key: "F4", Label: "Accounts"},
		{Key: "F5", Label: "Reports"},
		{Key: "F10", Label: "Menu"},
	}
}

// TransactionsStatusBar returns status bar items for transactions view
func TransactionsStatusBar() []StatusBarItem {
	return []StatusBarItem{
		{Key: "F1", Label: "Help"},
		{Key: "F3", Label: "Trans"},
		{Key: "Enter", Label: "Categorize"},
		{Key: "j/k", Label: "Navigate"},
		{Key: "g/G", Label: "Top/Bot"},
		{Key: "F10", Label: "Menu"},
	}
}

// AccountsStatusBar returns status bar items for accounts view
func AccountsStatusBar() []StatusBarItem {
	return []StatusBarItem{
		{Key: "F1", Label: "Help"},
		{Key: "F4", Label: "Accounts"},
		{Key: "j/k", Label: "Navigate"},
		{Key: "Enter", Label: "Expand"},
		{Key: "F10", Label: "Menu"},
	}
}

// ReportsStatusBar returns status bar items for reports view
func ReportsStatusBar() []StatusBarItem {
	return []StatusBarItem{
		{Key: "F1", Label: "Help"},
		{Key: "F5", Label: "Reports"},
		{Key: "j/k", Label: "Navigate"},
		{Key: "Enter", Label: "Select"},
		{Key: "F10", Label: "Menu"},
	}
}

// HelpStatusBar returns status bar items for help view
func HelpStatusBar() []StatusBarItem {
	return []StatusBarItem{
		{Key: "F1", Label: "Help"},
		{Key: "Esc", Label: "Close"},
		{Key: "j/k", Label: "Scroll"},
		{Key: "q", Label: "Quit"},
	}
}

// FormatStatusMessage creates a status message for one-off notifications
func FormatStatusMessage(message string, width int) string {
	// Center or left-align the message
	messageStyled := theme.StatusBarStyle.Render(message)
	renderedWidth := lipgloss.Width(messageStyled)

	if width > renderedWidth {
		padding := theme.StatusBarStyle.Render(strings.Repeat(" ", width-renderedWidth))
		return lipgloss.JoinHorizontal(lipgloss.Top, messageStyled, padding)
	}

	return messageStyled
}

// RenderWithMessage renders status bar with a temporary message on the right side
func (s StatusBar) RenderWithMessage(message string) string {
	leftSide := ""
	var parts []string

	// Render left side (F-keys)
	for i, item := range s.items {
		keyPart := theme.StatusBarStyle.Render(item.Key)
		labelPart := theme.StatusBarStyle.Render(item.Label)
		combined := keyPart + theme.StatusBarStyle.Render(" ") + labelPart
		parts = append(parts, combined)

		if i < len(s.items)-1 {
			parts = append(parts, theme.StatusBarStyle.Render("  "))
		}
	}
	leftSide = lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	// Render right side (message)
	messagePart := theme.StatusBarStyle.Render(fmt.Sprintf(" %s ", message))

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	messageWidth := lipgloss.Width(messagePart)
	spacingWidth := s.width - leftWidth - messageWidth

	if spacingWidth < 0 {
		spacingWidth = 0
	}

	spacing := theme.StatusBarStyle.Render(strings.Repeat(" ", spacingWidth))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftSide, spacing, messagePart)
}

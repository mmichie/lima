package theme

import "github.com/charmbracelet/lipgloss"

// TP7 Theme - Inspired by Borland Turbo Pascal 7
// Classic blue background with cyan highlights and yellow accents

// Color Constants - Classic TP7 Palette
const (
	TP7Blue       = "#0000AA" // Classic TP7 background blue
	TP7DarkBlue   = "#000088" // Slightly darker blue for alternating rows
	TP7Cyan       = "#00FFFF" // Menu bar and highlights
	TP7Black      = "#000000" // Selection background (inverted)
	TP7White      = "#FFFFFF" // Primary text
	TP7LightGray  = "#AAAAAA" // Secondary text and borders
	TP7DarkGray   = "#555555" // Muted text
	TP7Yellow     = "#FFFF00" // Warnings and highlights
	TP7Green      = "#00FF00" // Success/positive
	TP7Red        = "#FF0000" // Errors/negative
	TP7DarkCyan   = "#008888" // Subtle highlights
)

// Box Drawing Characters - TP7 style double-line borders
const (
	BoxTopLeft     = "╔"
	BoxTopRight    = "╗"
	BoxBottomLeft  = "╚"
	BoxBottomRight = "╝"
	BoxHorizontal  = "═"
	BoxVertical    = "║"
	BoxTeeLeft     = "╠"
	BoxTeeRight    = "╣"
	BoxTeeTop      = "╦"
	BoxTeeBottom   = "╩"
	BoxCross       = "╬"
)

// Base Styles
var (
	// Screen background
	ScreenStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(TP7Blue)).
			Foreground(lipgloss.Color(TP7White))

	// Menu bar style (top of screen) - TP7 style: light gray background, black text
	MenuBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(TP7LightGray)).
			Foreground(lipgloss.Color(TP7Black)).
			Bold(false)

	// Active menu item (when selected/hovered) - inverted: black background, white text
	MenuItemActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(TP7Black)).
				Foreground(lipgloss.Color(TP7White)).
				Bold(false)

	// Inactive menu item - same as menu bar
	MenuItemInactiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(TP7LightGray)).
				Foreground(lipgloss.Color(TP7Black)).
				Bold(false)

	// Hotkey letter in menu (underlined) - normal text, just underlined
	MenuHotkeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Black)).
			Background(lipgloss.Color(TP7LightGray)).
			Bold(false)

	// Status bar style (bottom of screen)
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(TP7Cyan)).
			Foreground(lipgloss.Color(TP7Black)).
			Bold(false)

	// Border style for dialogs and panels
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(TP7LightGray)).
			Background(lipgloss.Color(TP7Blue))

	// Title style for dialogs and sections
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Cyan)).
			Background(lipgloss.Color(TP7Blue)).
			Bold(true)

	// Normal text
	NormalTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7White)).
			Background(lipgloss.Color(TP7Blue))

	// Muted/secondary text
	MutedTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7LightGray)).
			Background(lipgloss.Color(TP7Blue))

	// Selected item in a list (inverted colors)
	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(TP7Cyan)).
				Foreground(lipgloss.Color(TP7Black)).
				Bold(false)

	// Normal list item
	ListItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7White)).
			Background(lipgloss.Color(TP7Blue))

	// Alternate list item (for striping/alternating rows)
	AlternateItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TP7White)).
				Background(lipgloss.Color(TP7DarkBlue))

	// Highlighted/focused element
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Yellow)).
			Background(lipgloss.Color(TP7Blue)).
			Bold(true)

	// Success/positive indicators
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Green)).
			Background(lipgloss.Color(TP7Blue))

	// Warning indicators
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Yellow)).
			Background(lipgloss.Color(TP7Blue))

	// Error/negative indicators
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Red)).
			Background(lipgloss.Color(TP7Blue))

	// Date/timestamp style
	DateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7DarkCyan)).
			Background(lipgloss.Color(TP7Blue))

	// Amount style (neutral)
	AmountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7White)).
			Background(lipgloss.Color(TP7Blue))

	// Positive amount
	AmountPositiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TP7Green)).
				Background(lipgloss.Color(TP7Blue))

	// Negative amount
	AmountNegativeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TP7Red)).
				Background(lipgloss.Color(TP7Blue))

	// Input field style
	InputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Black)).
			Background(lipgloss.Color(TP7Cyan))

	// Button style (normal)
	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(TP7Black)).
			Background(lipgloss.Color(TP7LightGray)).
			Padding(0, 2)

	// Button style (focused)
	ButtonFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(TP7Black)).
				Background(lipgloss.Color(TP7Cyan)).
				Padding(0, 2).
				Bold(true)
)

// RenderBox renders a TP7-style double-line box around content
func RenderBox(title string, content string, width int) string {
	// Create top border
	topBorder := BoxTopLeft + title
	remainingWidth := width - len(title) - 2
	if remainingWidth > 0 {
		topBorder += lipgloss.NewStyle().Render(repeatString(BoxHorizontal, remainingWidth))
	}
	topBorder += BoxTopRight

	// Split content into lines and wrap in borders
	lines := lipgloss.NewStyle().Width(width - 4).Render(content)
	wrappedLines := ""
	for _, line := range lipgloss.NewStyle().Width(width - 4).Render(lines) {
		wrappedLines += BoxVertical + " " + string(line) + " " + BoxVertical + "\n"
	}

	// Create bottom border
	bottomBorder := BoxBottomLeft + repeatString(BoxHorizontal, width-2) + BoxBottomRight

	return BorderStyle.Render(topBorder + "\n" + wrappedLines + bottomBorder)
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

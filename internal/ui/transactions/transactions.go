package transactions

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/internal/categorizer"
	"github.com/mmichie/lima/internal/ui/theme"
)

// keyMap defines key bindings for the transactions view
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Enter    key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+f"),
			key.WithHelp("pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "details"),
		),
	}
}

// Model represents the transactions view model
type Model struct {
	file        *beancount.File
	categorizer *categorizer.Categorizer
	width       int
	height      int

	// List state for manual rendering
	cursor int
	offset int

	keys keyMap

	// Category picker state
	showingPicker      bool
	pickerCursor       int
	currentSuggestions []*categorizer.Suggestion

	// Cached data
	totalTransactions int
}

// New creates a new transactions model
func New(file *beancount.File, cat *categorizer.Categorizer) Model {
	return Model{
		file:              file,
		categorizer:       cat,
		cursor:            0,
		offset:            0,
		keys:              newKeyMap(),
		totalTransactions: file.TransactionCount(),
		showingPicker:     false,
		pickerCursor:      0,
	}
}

// Init initializes the transactions view
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If category picker is showing, handle picker navigation
		if m.showingPicker {
			switch msg.String() {
			case "esc", "q":
				m.showingPicker = false
				m.pickerCursor = 0
				return m, nil

			case "up", "k":
				if m.pickerCursor > 0 {
					m.pickerCursor--
				}
				return m, nil

			case "down", "j":
				if m.pickerCursor < len(m.currentSuggestions)-1 {
					m.pickerCursor++
				}
				return m, nil

			case "enter":
				// TODO: Apply selected category to transaction
				m.showingPicker = false
				m.pickerCursor = 0
				return m, nil
			}
			return m, nil
		}

		// Handle navigation keys
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				// Adjust offset if cursor moves above visible area
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < m.totalTransactions-1 {
				m.cursor++
				// Adjust offset if cursor moves below visible area
				visibleRows := m.height - 4 // Account for title and padding
				if m.cursor >= m.offset+visibleRows {
					m.offset = m.cursor - visibleRows + 1
				}
			}

		case key.Matches(msg, m.keys.Top):
			m.cursor = 0
			m.offset = 0

		case key.Matches(msg, m.keys.Bottom):
			m.cursor = m.totalTransactions - 1
			visibleRows := m.height - 4
			m.offset = m.cursor - visibleRows + 1
			if m.offset < 0 {
				m.offset = 0
			}

		case key.Matches(msg, m.keys.Enter):
			// Get categorization suggestions for current transaction
			if m.categorizer != nil && m.totalTransactions > 0 {
				tx, err := m.file.GetTransaction(m.cursor)
				if err == nil {
					suggestions, err := m.categorizer.SuggestAll(tx)
					if err == nil && len(suggestions) > 0 {
						m.currentSuggestions = suggestions
						m.showingPicker = true
						m.pickerCursor = 0
						return m, nil
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the transactions view
func (m Model) View() string {
	if m.totalTransactions == 0 {
		titleText := "Transactions (0 total)"
		titlePadded := titleText
		if m.width > len(titleText) {
			titlePadded = titleText + strings.Repeat(" ", m.width-len(titleText))
		}
		title := theme.TitleStyle.Width(m.width).Render(titlePadded)
		return title + "\n" + theme.NormalTextStyle.Render("No transactions found")
	}

	var lines []string

	// Title with count and cursor position
	titleText := fmt.Sprintf("Transactions (%d total) - Row %d/%d", m.totalTransactions, m.cursor+1, m.totalTransactions)
	titlePadded := titleText
	if m.width > len(titleText) {
		titlePadded = titleText + strings.Repeat(" ", m.width-len(titleText))
	}
	title := theme.TitleStyle.Width(m.width).Render(titlePadded)
	lines = append(lines, title)
	lines = append(lines, "")

	// Table header
	headerLine := fmt.Sprintf("%-12s  %1s  %-40s  %-45s  %15s", "Date", "", "Description", "Account", "Amount")
	if m.width > len(headerLine) {
		headerLine = headerLine + strings.Repeat(" ", m.width-len(headerLine))
	}
	lines = append(lines, theme.TitleStyle.Width(m.width).Render(headerLine))

	// Separator
	separator := strings.Repeat("─", m.width)
	lines = append(lines, theme.MutedTextStyle.Width(m.width).Render(separator))

	// Calculate visible range
	visibleRows := m.height - 6 // Account for title, header, separator, padding
	if visibleRows < 1 {
		visibleRows = 1
	}
	end := m.offset + visibleRows
	if end > m.totalTransactions {
		end = m.totalTransactions
	}

	// Render visible transactions
	for i := m.offset; i < end; i++ {
		tx, err := m.file.GetTransaction(i)
		if err != nil {
			continue
		}

		// Format date
		dateStr := tx.Date.Format("2006-01-02")

		// Format flag
		flagStr := tx.Flag

		// Format description
		description := tx.Narration
		if tx.Payee != "" {
			description = tx.Payee
		}
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		// Format account
		account := ""
		if len(tx.Postings) > 0 {
			account = tx.Postings[0].Account
			if len(account) > 45 {
				account = "..." + account[len(account)-42:]
			}
		}

		// Format amount
		amount := ""
		if len(tx.Postings) > 0 && tx.Postings[0].Amount != nil {
			amt := tx.Postings[0].Amount.Number.StringFixed(2)
			commodity := tx.Postings[0].Amount.Commodity
			amount = fmt.Sprintf("%s %s", amt, commodity)
		}

		// Build the row line
		line := fmt.Sprintf("%-12s  %1s  %-40s  %-45s  %15s", dateStr, flagStr, description, account, amount)

		// Pad to full width
		if m.width > len(line) {
			line = line + strings.Repeat(" ", m.width-len(line))
		}

		// Apply highlighting for selected row
		if i == m.cursor {
			line = theme.SelectedItemStyle.Width(m.width).Render(line)
		} else {
			line = theme.ListItemStyle.Width(m.width).Render(line)
		}

		lines = append(lines, line)
	}

	view := strings.Join(lines, "\n")

	// Show category picker overlay if active
	if m.showingPicker {
		return view + "\n\n" + m.renderCategoryPicker()
	}

	return view
}

// SetSize updates the transactions view size
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// renderCategoryPicker renders the category picker overlay with TP7 styling
func (m Model) renderCategoryPicker() string {
	// Use TP7 double-line box drawing characters
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(theme.TP7Cyan)).
		BorderBackground(lipgloss.Color(theme.TP7Blue)).
		Background(lipgloss.Color(theme.TP7Blue)).
		Padding(1, 2).
		Width(m.width - 4)

	var lines []string
	lines = append(lines, theme.TitleStyle.Render("Category Suggestions"))
	lines = append(lines, "")

	if len(m.currentSuggestions) == 0 {
		lines = append(lines, theme.NormalTextStyle.Render("No categorization suggestions available"))
	} else {
		for i, suggestion := range m.currentSuggestions {
			confidence := fmt.Sprintf("%.0f%%", suggestion.Confidence*100)

			// Show indicator based on confidence using TP7 colors
			indicator := "+"
			indicatorStyle := theme.SuccessStyle
			if suggestion.Confidence < 0.8 {
				indicator = "~"
				indicatorStyle = theme.WarningStyle
			} else if suggestion.Confidence >= 0.95 {
				indicator = "*"
				indicatorStyle = theme.HighlightStyle
			}

			line := fmt.Sprintf("%s %s (%s)",
				indicatorStyle.Render(indicator),
				suggestion.Category,
				confidence)

			if i == m.pickerCursor {
				line = theme.SelectedItemStyle.Render(" > " + line)
			} else {
				line = theme.ListItemStyle.Render("   " + line)
			}

			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, theme.MutedTextStyle.Render("j/k:navigate   enter:select   esc:cancel"))

	content := strings.Join(lines, "\n")
	return pickerStyle.Render(content)
}

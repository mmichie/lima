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

	// List state
	cursor int
	offset int
	keys   keyMap

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

		// Normal navigation
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				// Scroll up if cursor moves above viewport
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < m.totalTransactions-1 {
				m.cursor++
				// Scroll down if cursor moves below viewport
				maxVisible := m.height - 3 // Account for title and padding
				if m.cursor >= m.offset+maxVisible {
					m.offset = m.cursor - maxVisible + 1
				}
			}

		case key.Matches(msg, m.keys.PageUp):
			pageSize := m.height - 3
			m.cursor -= pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.offset = m.cursor

		case key.Matches(msg, m.keys.PageDown):
			pageSize := m.height - 3
			m.cursor += pageSize
			if m.cursor >= m.totalTransactions {
				m.cursor = m.totalTransactions - 1
			}
			m.offset = m.cursor - (m.height - 4)
			if m.offset < 0 {
				m.offset = 0
			}

		case key.Matches(msg, m.keys.Top):
			m.cursor = 0
			m.offset = 0

		case key.Matches(msg, m.keys.Bottom):
			m.cursor = m.totalTransactions - 1
			m.offset = m.cursor - (m.height - 4)
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
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the transactions view
func (m Model) View() string {
	var lines []string

	// Title with count using TP7 style - fill full width
	titleText := fmt.Sprintf("Transactions (%d total)", m.totalTransactions)
	titlePadded := titleText
	if m.width > len(titleText) {
		titlePadded = titleText + strings.Repeat(" ", m.width-len(titleText))
	}
	title := theme.TitleStyle.Width(m.width).Render(titlePadded)
	lines = append(lines, title)
	lines = append(lines, "")

	if m.totalTransactions == 0 {
		lines = append(lines, theme.NormalTextStyle.Render("No transactions found"))
		return strings.Join(lines, "\n")
	}

	// Add table header
	header := m.renderTableHeader()
	lines = append(lines, header)

	// Add separator line
	separator := m.renderSeparatorLine()
	lines = append(lines, separator)

	// Calculate visible range
	// Account for: title + blank line + header + separator + data lines
	// Default to showing at least 10 transactions if height not set
	maxVisible := m.height - 5 // -5 for title, blank, header, separator, and padding
	if maxVisible <= 0 {
		maxVisible = 10
	}
	endIdx := m.offset + maxVisible
	if endIdx > m.totalTransactions {
		endIdx = m.totalTransactions
	}

	// Render visible transactions
	for i := m.offset; i < endIdx; i++ {
		tx, err := m.file.GetTransaction(i)
		if err != nil {
			// Show error instead of silently skipping
			errMsg := fmt.Sprintf("Error loading transaction %d: %v", i, err)
			lines = append(lines, theme.ErrorStyle.Render(errMsg))
			continue
		}

		line := m.renderTransactionLine(tx, i == m.cursor)
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

// renderTransactionLine renders a single transaction line with TP7 styling
func (m Model) renderTransactionLine(tx *beancount.Transaction, selected bool) string {
	// Date
	dateStr := tx.Date.Format("2006-01-02")

	// Flag
	var flagStr string
	if tx.Flag == "*" {
		flagStr = "*"
	} else {
		flagStr = "!"
	}

	// Description (payee)
	description := tx.Narration
	if tx.Payee != "" {
		description = tx.Payee
	}
	// Calculate available space for description
	descWidth := 40
	if len(description) > descWidth {
		description = description[:descWidth-3] + "..."
	}

	// Account (first posting)
	account := ""
	if len(tx.Postings) > 0 {
		account = tx.Postings[0].Account
	}
	// Calculate available space for account
	accountWidth := 45
	if len(account) > accountWidth {
		account = "..." + account[len(account)-accountWidth+3:]
	}

	// Get first posting amount for summary
	var amountStr string
	var amountRaw string
	if len(tx.Postings) > 0 && tx.Postings[0].Amount != nil {
		amount := tx.Postings[0].Amount.Number.StringFixed(2)
		commodity := tx.Postings[0].Amount.Commodity
		amountRaw = fmt.Sprintf("%s %s", amount, commodity)
		amountStr = amountRaw
	}

	// Build line with proper spacing and column separators
	line := fmt.Sprintf("%s │ %s │ %-*s │ %-*s │ %15s",
		dateStr,
		flagStr,
		descWidth,
		description,
		accountWidth,
		account,
		amountStr,
	)

	// Pad line to full width
	lineLen := len(line)
	if m.width > lineLen {
		line = line + strings.Repeat(" ", m.width-lineLen)
	}

	// Apply styling
	if selected {
		// Selected line: black on cyan (TP7 style) - fill entire width
		return theme.SelectedItemStyle.Width(m.width).Render(line)
	}

	// Normal line: styled with colors, full width
	styledLine := fmt.Sprintf("%s │ %s │ %-*s │ %-*s │ %15s",
		theme.DateStyle.Render(dateStr),
		theme.WarningStyle.Render(flagStr),
		descWidth,
		theme.NormalTextStyle.Render(description),
		accountWidth,
		theme.NormalTextStyle.Render(account),
		formatTransactionAmount(amountRaw),
	)

	// Pad to full width
	styledLen := lipgloss.Width(styledLine)
	if m.width > styledLen {
		styledLine = styledLine + theme.ListItemStyle.Render(strings.Repeat(" ", m.width-styledLen))
	}

	// Use alternating background for visual structure (every other row)
	style := theme.ListItemStyle
	// We could add alternating here if desired, but let's keep it simple for now
	return style.Width(m.width).Render(styledLine)
}

// formatTransactionAmount formats an amount with proper coloring
func formatTransactionAmount(amount string) string {
	if amount == "" {
		return ""
	}
	if strings.HasPrefix(amount, "-") {
		return theme.AmountNegativeStyle.Render(amount)
	}
	return theme.AmountPositiveStyle.Render(amount)
}

// renderTableHeader renders the column headers for the transaction table
func (m Model) renderTableHeader() string {
	descWidth := 40
	accountWidth := 45

	header := fmt.Sprintf("%-12s │ %-1s │ %-*s │ %-*s │ %15s",
		"Date",
		"",
		descWidth,
		"Description",
		accountWidth,
		"Account",
		"Amount",
	)

	// Pad to full width
	headerLen := lipgloss.Width(header)
	if m.width > headerLen {
		header = header + strings.Repeat(" ", m.width-headerLen)
	}

	return theme.HighlightStyle.Width(m.width).Render(header)
}

// renderSeparatorLine renders a separator line for the table
func (m Model) renderSeparatorLine() string {
	// Use ─ character for horizontal line
	line := strings.Repeat("─", m.width)
	return theme.MutedTextStyle.Width(m.width).Render(line)
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

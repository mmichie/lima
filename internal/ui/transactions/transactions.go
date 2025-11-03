package transactions

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
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

	// Native table component
	table table.Model
	keys  keyMap

	// Category picker state
	showingPicker      bool
	pickerCursor       int
	currentSuggestions []*categorizer.Suggestion

	// Cached data
	totalTransactions int
}

// New creates a new transactions model
func New(file *beancount.File, cat *categorizer.Categorizer) Model {
	totalTransactions := file.TransactionCount()

	// Define table columns
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "", Width: 1},
		{Title: "Description", Width: 40},
		{Title: "Account", Width: 45},
		{Title: "Amount", Width: 15},
	}

	// Build rows from transactions
	rows := []table.Row{}
	for i := 0; i < totalTransactions; i++ {
		tx, err := file.GetTransaction(i)
		if err != nil {
			continue
		}

		// Format transaction data
		dateStr := tx.Date.Format("2006-01-02")
		flag := tx.Flag

		description := tx.Narration
		if tx.Payee != "" {
			description = tx.Payee
		}
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		account := ""
		if len(tx.Postings) > 0 {
			account = tx.Postings[0].Account
			if len(account) > 45 {
				account = "..." + account[len(account)-42:]
			}
		}

		amount := ""
		if len(tx.Postings) > 0 && tx.Postings[0].Amount != nil {
			amt := tx.Postings[0].Amount.Number.StringFixed(2)
			commodity := tx.Postings[0].Amount.Commodity
			amount = fmt.Sprintf("%s %s", amt, commodity)
		}

		rows = append(rows, table.Row{dateStr, flag, description, account, amount})
	}

	// Create table with TP7 styling
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply TP7 table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(theme.TP7Cyan)).
		BorderBottom(true).
		Foreground(lipgloss.Color(theme.TP7Yellow)).
		Background(lipgloss.Color(theme.TP7Blue)).
		Bold(true)
	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TP7Black)).
		Background(lipgloss.Color(theme.TP7Cyan)).
		Bold(false)
	s.Cell = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TP7White)).
		Background(lipgloss.Color(theme.TP7Blue))

	t.SetStyles(s)

	// Enable focus to show selection
	t.Focus()

	return Model{
		file:              file,
		categorizer:       cat,
		table:             t,
		keys:              newKeyMap(),
		totalTransactions: totalTransactions,
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
	var cmd tea.Cmd

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

		// Handle enter key for categorization (only when NOT showing picker)
		if msg.String() == "enter" {
			// Get categorization suggestions for current transaction
			if m.categorizer != nil && m.totalTransactions > 0 {
				cursor := m.table.Cursor()
				tx, err := m.file.GetTransaction(cursor)
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
			// If no suggestions, still consume the enter key
			return m, nil
		}

		// Map j/k to arrow keys for vim-style navigation
		switch msg.String() {
		case "j":
			m.table, cmd = m.table.Update(tea.KeyMsg{Type: tea.KeyDown})
			return m, cmd
		case "k":
			m.table, cmd = m.table.Update(tea.KeyMsg{Type: tea.KeyUp})
			return m, cmd
		}

		// For all other keys, delegate to table (arrows/pgup/pgdn/home/end/g/G)
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	default:
		// For non-key messages, still delegate to table
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
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

	// Title with count
	titleText := fmt.Sprintf("Transactions (%d total)", m.totalTransactions)
	titlePadded := titleText
	if m.width > len(titleText) {
		titlePadded = titleText + strings.Repeat(" ", m.width-len(titleText))
	}
	title := theme.TitleStyle.Width(m.width).Render(titlePadded)

	// Render table
	tableView := m.table.View()

	view := title + "\n\n" + tableView

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

	// Update table height (account for title and padding)
	tableHeight := height - 4
	if tableHeight < 5 {
		tableHeight = 5
	}
	m.table.SetHeight(tableHeight)

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

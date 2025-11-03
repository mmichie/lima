package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/internal/ui/theme"
)

// Model represents the dashboard view model
type Model struct {
	file   *beancount.File
	width  int
	height int

	// Cached statistics
	totalTransactions int
	totalAccounts     int
	totalCommodities  int
	recentCount       int
}

// New creates a new dashboard model
func New(file *beancount.File) Model {
	return Model{
		file:              file,
		totalTransactions: file.TransactionCount(),
		totalAccounts:     len(file.GetAccounts()),
		totalCommodities:  len(file.GetCommodities()),
		recentCount:       5,
	}
}

// Init initializes the dashboard
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View renders the dashboard with TP7 styling
func (m Model) View() string {
	if m.width == 0 {
		return theme.NormalTextStyle.Render("Loading dashboard...")
	}

	// Title
	title := theme.TitleStyle.Render("Dashboard")

	// Statistics boxes
	stats := m.renderStats()

	// Recent transactions
	recent := m.renderRecentTransactions()

	// Combine all sections
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		stats,
		"",
		recent,
	)

	return content
}

// SetSize updates the dashboard size
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// renderStats renders the statistics boxes with TP7 styling
func (m Model) renderStats() string {
	// Use TP7 double-line box drawing characters
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(theme.TP7Cyan)).
		BorderBackground(lipgloss.Color(theme.TP7Blue)).
		Background(lipgloss.Color(theme.TP7Blue)).
		Padding(1, 2).
		Width(30)

	// Transaction stats box
	transactionsBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		theme.MutedTextStyle.Render("Total Transactions"),
		theme.HighlightStyle.Render(fmt.Sprintf("%d", m.totalTransactions)),
	))

	// Accounts box
	accountsBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		theme.MutedTextStyle.Render("Accounts"),
		theme.HighlightStyle.Render(fmt.Sprintf("%d", m.totalAccounts)),
	))

	// Commodities box
	commoditiesBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		theme.MutedTextStyle.Render("Commodities"),
		theme.HighlightStyle.Render(fmt.Sprintf("%d", m.totalCommodities)),
	))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		transactionsBox,
		"  ",
		accountsBox,
		"  ",
		commoditiesBox,
	)
}

// renderRecentTransactions renders the most recent transactions with TP7 styling
func (m Model) renderRecentTransactions() string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, theme.TitleStyle.Render("Recent Transactions"))
	lines = append(lines, "")

	count := m.recentCount
	if count > m.totalTransactions {
		count = m.totalTransactions
	}

	for i := 0; i < count; i++ {
		tx, err := m.file.GetTransaction(i)
		if err != nil {
			continue
		}

		// Format date
		dateStr := tx.Date.Format("2006-01-02")

		// Format payee/narration
		description := tx.Narration
		if tx.Payee != "" {
			description = tx.Payee + " - " + tx.Narration
		}
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		// Format flag with TP7 colors
		flagStr := tx.Flag
		if tx.Flag == "*" {
			flagStr = theme.SuccessStyle.Render("*")
		} else {
			flagStr = theme.WarningStyle.Render("!")
		}

		// Format the line - simple text on blue background
		line := fmt.Sprintf("  %s  %-50s  %s",
			theme.DateStyle.Render(dateStr),
			theme.NormalTextStyle.Render(description),
			flagStr,
		)

		lines = append(lines, line)
	}

	if m.totalTransactions == 0 {
		lines = append(lines, theme.MutedTextStyle.Render("  No transactions found"))
	}

	return strings.Join(lines, "\n")
}

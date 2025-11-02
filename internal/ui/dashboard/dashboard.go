package dashboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/beancount"
	"github.com/shopspring/decimal"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF")).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(30)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	statValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF"))
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

// View renders the dashboard
func (m Model) View() string {
	if m.width == 0 {
		return "Loading dashboard..."
	}

	// Title
	title := titleStyle.Render("Dashboard")

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

// renderStats renders the statistics boxes
func (m Model) renderStats() string {
	// Transaction stats box
	transactionsBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		statLabelStyle.Render("Total Transactions"),
		statValueStyle.Render(fmt.Sprintf("%d", m.totalTransactions)),
	))

	// Accounts box
	accountsBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		statLabelStyle.Render("Accounts"),
		statValueStyle.Render(fmt.Sprintf("%d", m.totalAccounts)),
	))

	// Commodities box
	commoditiesBox := boxStyle.Render(fmt.Sprintf(
		"%s\n%s",
		statLabelStyle.Render("Commodities"),
		statValueStyle.Render(fmt.Sprintf("%d", m.totalCommodities)),
	))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		transactionsBox,
		"  ",
		accountsBox,
		"  ",
		commoditiesBox,
	)
}

// renderRecentTransactions renders the most recent transactions
func (m Model) renderRecentTransactions() string {
	var lines []string
	lines = append(lines, titleStyle.Render("Recent Transactions"))

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

		// Get total amount (sum of all postings)
		var total decimal.Decimal
		var commodity string
		for _, posting := range tx.Postings {
			if posting.Amount != nil {
				total = total.Add(posting.Amount.Number)
				if commodity == "" {
					commodity = posting.Amount.Commodity
				}
			}
		}

		// Format the line
		line := fmt.Sprintf("  %s  %s  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(dateStr),
			description,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Render(tx.Flag),
		)

		lines = append(lines, line)
	}

	if m.totalTransactions == 0 {
		lines = append(lines, "  No transactions found")
	}

	return strings.Join(lines, "\n")
}

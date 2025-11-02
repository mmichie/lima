package accounts

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/beancount"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D9FF")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D9FF")).
			Background(lipgloss.Color("#333333")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	categoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1)
)

// keyMap defines key bindings for the accounts view
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
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
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "bottom"),
		),
	}
}

// Model represents the accounts view model
type Model struct {
	file   *beancount.File
	width  int
	height int

	// List state
	cursor   int
	accounts []string
	keys     keyMap

	// Grouped accounts
	assets     []string
	liabilities []string
	equity     []string
	income     []string
	expenses   []string
}

// New creates a new accounts model
func New(file *beancount.File) Model {
	accounts := file.GetAccounts()

	m := Model{
		file:     file,
		cursor:   0,
		accounts: accounts,
		keys:     newKeyMap(),
	}

	// Group accounts by type
	for _, acc := range accounts {
		if strings.HasPrefix(acc, "Assets:") {
			m.assets = append(m.assets, acc)
		} else if strings.HasPrefix(acc, "Liabilities:") {
			m.liabilities = append(m.liabilities, acc)
		} else if strings.HasPrefix(acc, "Equity:") {
			m.equity = append(m.equity, acc)
		} else if strings.HasPrefix(acc, "Income:") {
			m.income = append(m.income, acc)
		} else if strings.HasPrefix(acc, "Expenses:") {
			m.expenses = append(m.expenses, acc)
		}
	}

	return m
}

// Init initializes the accounts view
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.accounts)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Top):
			m.cursor = 0

		case key.Matches(msg, m.keys.Bottom):
			m.cursor = len(m.accounts) - 1
		}
	}

	return m, nil
}

// View renders the accounts view
func (m Model) View() string {
	if m.width == 0 {
		return "Loading accounts..."
	}

	var lines []string

	// Title
	title := titleStyle.Render(fmt.Sprintf("Accounts (%d total)", len(m.accounts)))
	lines = append(lines, title)

	if len(m.accounts) == 0 {
		lines = append(lines, "No accounts found")
		return strings.Join(lines, "\n")
	}

	// Render grouped accounts
	currentIdx := 0

	if len(m.assets) > 0 {
		lines = append(lines, categoryStyle.Render("Assets"))
		currentIdx = m.renderAccountGroup(m.assets, currentIdx, &lines)
	}

	if len(m.liabilities) > 0 {
		lines = append(lines, categoryStyle.Render("Liabilities"))
		currentIdx = m.renderAccountGroup(m.liabilities, currentIdx, &lines)
	}

	if len(m.equity) > 0 {
		lines = append(lines, categoryStyle.Render("Equity"))
		currentIdx = m.renderAccountGroup(m.equity, currentIdx, &lines)
	}

	if len(m.income) > 0 {
		lines = append(lines, categoryStyle.Render("Income"))
		currentIdx = m.renderAccountGroup(m.income, currentIdx, &lines)
	}

	if len(m.expenses) > 0 {
		lines = append(lines, categoryStyle.Render("Expenses"))
		currentIdx = m.renderAccountGroup(m.expenses, currentIdx, &lines)
	}

	// Add navigation help
	help := fmt.Sprintf("  j/k:navigate • g/G:top/bottom • 1-4:switch view")
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(help))

	return strings.Join(lines, "\n")
}

// SetSize updates the accounts view size
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// renderAccountGroup renders a group of accounts
func (m Model) renderAccountGroup(accounts []string, startIdx int, lines *[]string) int {
	idx := startIdx
	for _, acc := range accounts {
		if idx == m.cursor {
			*lines = append(*lines, selectedStyle.Render("  > "+acc))
		} else {
			*lines = append(*lines, normalStyle.Render("    "+acc))
		}
		idx++
	}
	return idx
}

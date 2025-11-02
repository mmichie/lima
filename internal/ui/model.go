package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mmichie/lima/internal/beancount"
	"github.com/mmichie/lima/internal/ui/accounts"
	"github.com/mmichie/lima/internal/ui/dashboard"
	"github.com/mmichie/lima/internal/ui/transactions"
	"github.com/mmichie/lima/pkg/config"
)

// ViewType represents the different views in the application
type ViewType int

const (
	DashboardView ViewType = iota
	TransactionsView
	AccountsView
	ReportsView
)

// Model is the main application model
type Model struct {
	// Current view
	currentView ViewType

	// Beancount file
	file *beancount.File

	// Configuration
	config *config.Config

	// View models
	dashboard    dashboard.Model
	transactions transactions.Model
	accounts     accounts.Model

	// UI state
	width  int
	height int
	ready  bool

	// Key bindings
	keys keyMap
}

// keyMap defines the key bindings for navigation
type keyMap struct {
	Dashboard    key.Binding
	Transactions key.Binding
	Accounts     key.Binding
	Reports      key.Binding
	Quit         key.Binding
	Help         key.Binding
}

// keyMapFromConfig creates key bindings from config
func keyMapFromConfig(cfg *config.Config) keyMap {
	return keyMap{
		Dashboard: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Dashboard...),
			key.WithHelp(cfg.Keybindings.Dashboard[0], "dashboard"),
		),
		Transactions: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Transactions...),
			key.WithHelp(cfg.Keybindings.Transactions[0], "transactions"),
		),
		Accounts: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Accounts...),
			key.WithHelp(cfg.Keybindings.Accounts[0], "accounts"),
		),
		Reports: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Reports...),
			key.WithHelp(cfg.Keybindings.Reports[0], "reports"),
		),
		Quit: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Quit...),
			key.WithHelp(cfg.Keybindings.Quit[0], "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(cfg.Keybindings.Help...),
			key.WithHelp(cfg.Keybindings.Help[0], "help"),
		),
	}
}

// New creates a new main application model
func New(file *beancount.File, cfg *config.Config) Model {
	// Parse initial view from config
	var initialView ViewType
	switch cfg.UI.DefaultView {
	case "transactions":
		initialView = TransactionsView
	case "accounts":
		initialView = AccountsView
	case "reports":
		initialView = ReportsView
	default:
		initialView = DashboardView
	}

	return Model{
		currentView:  initialView,
		file:         file,
		config:       cfg,
		keys:         keyMapFromConfig(cfg),
		dashboard:    dashboard.New(file),
		transactions: transactions.New(file),
		accounts:     accounts.New(file),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update all view sizes
		m.dashboard = m.dashboard.SetSize(msg.Width, msg.Height-4) // -4 for header/footer
		m.transactions = m.transactions.SetSize(msg.Width, msg.Height-4)
		m.accounts = m.accounts.SetSize(msg.Width, msg.Height-4)

		return m, nil

	case tea.KeyMsg:
		// Global navigation keys
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Dashboard):
			m.currentView = DashboardView
			return m, nil

		case key.Matches(msg, m.keys.Transactions):
			m.currentView = TransactionsView
			return m, nil

		case key.Matches(msg, m.keys.Accounts):
			m.currentView = AccountsView
			return m, nil

		case key.Matches(msg, m.keys.Reports):
			m.currentView = ReportsView
			return m, nil
		}
	}

	// Route to current view
	switch m.currentView {
	case DashboardView:
		newDashboard, cmd := m.dashboard.Update(msg)
		m.dashboard = newDashboard.(dashboard.Model)
		cmds = append(cmds, cmd)

	case TransactionsView:
		newTransactions, cmd := m.transactions.Update(msg)
		m.transactions = newTransactions.(transactions.Model)
		cmds = append(cmds, cmd)

	case AccountsView:
		newAccounts, cmd := m.accounts.Update(msg)
		m.accounts = newAccounts.(accounts.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Render header
	header := renderHeader(m.currentView)

	// Render current view
	var content string
	switch m.currentView {
	case DashboardView:
		content = m.dashboard.View()
	case TransactionsView:
		content = m.transactions.View()
	case AccountsView:
		content = m.accounts.View()
	case ReportsView:
		content = "Reports view coming soon..."
	}

	// Render footer
	footer := renderFooter(m.keys)

	return header + "\n" + content + "\n" + footer
}

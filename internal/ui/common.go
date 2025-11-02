package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette
	primaryColor   = lipgloss.Color("#00D9FF")
	secondaryColor = lipgloss.Color("#7D56F4")
	successColor   = lipgloss.Color("#00FF00")
	warningColor   = lipgloss.Color("#FFFF00")
	errorColor     = lipgloss.Color("#FF0000")
	mutedColor     = lipgloss.Color("#666666")
	textColor      = lipgloss.Color("#FFFFFF")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#333333")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Background(lipgloss.Color("#1a1a1a")).
				Padding(0, 2)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	// Content styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			MarginBottom(1)

	// List styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(textColor)
)

// renderHeader renders the application header with navigation tabs
func renderHeader(currentView ViewType) string {
	var tabs []string

	// Dashboard tab
	if currentView == DashboardView {
		tabs = append(tabs, activeTabStyle.Render("1: Dashboard"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("1: Dashboard"))
	}

	// Transactions tab
	if currentView == TransactionsView {
		tabs = append(tabs, activeTabStyle.Render("2: Transactions"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("2: Transactions"))
	}

	// Accounts tab
	if currentView == AccountsView {
		tabs = append(tabs, activeTabStyle.Render("3: Accounts"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("3: Accounts"))
	}

	// Reports tab
	if currentView == ReportsView {
		tabs = append(tabs, activeTabStyle.Render("4: Reports"))
	} else {
		tabs = append(tabs, inactiveTabStyle.Render("4: Reports"))
	}

	header := headerStyle.Render("Lima - Beancount TUI")
	tabsRendered := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabsRendered,
	)
}

// renderFooter renders the application footer with key bindings
func renderFooter(keys keyMap) string {
	helpText := fmt.Sprintf(
		"%s %s • %s %s • %s %s • %s %s • %s %s",
		keys.Dashboard.Help().Key, keys.Dashboard.Help().Desc,
		keys.Transactions.Help().Key, keys.Transactions.Help().Desc,
		keys.Accounts.Help().Key, keys.Accounts.Help().Desc,
		keys.Reports.Help().Key, keys.Reports.Help().Desc,
		keys.Quit.Help().Key, keys.Quit.Help().Desc,
	)

	return footerStyle.Render(helpText)
}

// formatAmount formats a decimal amount with commodity
func formatAmount(amount string, commodity string) string {
	amountStyle := lipgloss.NewStyle().Foreground(successColor)
	if strings.HasPrefix(amount, "-") {
		amountStyle = lipgloss.NewStyle().Foreground(errorColor)
	}
	return amountStyle.Render(amount + " " + commodity)
}

// formatDate formats a date string
func formatDate(date string) string {
	style := lipgloss.NewStyle().Foreground(mutedColor)
	return style.Render(date)
}

// formatAccount formats an account name
func formatAccount(account string) string {
	style := lipgloss.NewStyle().Foreground(textColor)
	return style.Render(account)
}

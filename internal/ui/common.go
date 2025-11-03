package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mmichie/lima/internal/ui/components"
	"github.com/mmichie/lima/internal/ui/theme"
)

// renderHeader renders the TP7-style menu bar
func renderHeader(menuBar components.MenuBar) string {
	return menuBar.View()
}

// renderFooter renders the TP7-style status bar based on current view
func renderFooter(currentView ViewType, statusBar components.StatusBar) string {
	// Set context-specific status bar items based on view
	var items []components.StatusBarItem
	switch currentView {
	case DashboardView:
		items = components.DashboardStatusBar()
	case TransactionsView:
		items = components.TransactionsStatusBar()
	case AccountsView:
		items = components.AccountsStatusBar()
	case ReportsView:
		items = components.ReportsStatusBar()
	default:
		items = components.DashboardStatusBar()
	}

	statusBar = statusBar.SetItems(items)
	return statusBar.View()
}

// formatAmount formats a decimal amount with commodity using TP7 theme
func formatAmount(amount string, commodity string) string {
	amountStyle := theme.AmountPositiveStyle
	if strings.HasPrefix(amount, "-") {
		amountStyle = theme.AmountNegativeStyle
	}
	return amountStyle.Render(amount + " " + commodity)
}

// formatDate formats a date string using TP7 theme
func formatDate(date string) string {
	return theme.DateStyle.Render(date)
}

// formatAccount formats an account name using TP7 theme
func formatAccount(account string) string {
	return theme.NormalTextStyle.Render(account)
}

// renderFullScreenContent fills the content area with TP7 blue background
func renderFullScreenContent(content string, width, height int) string {
	// Create a style that fills the entire content area with blue background
	fullScreenStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(theme.TP7Blue)).
		Width(width).
		Height(height)

	return fullScreenStyle.Render(content)
}

// renderReportsPlaceholder renders a placeholder for the reports view
func renderReportsPlaceholder() string {
	title := theme.TitleStyle.Render("Reports")
	message := theme.NormalTextStyle.Render("\n  Reports view coming soon...")

	return title + message
}

// renderLoadingScreen renders a TP7-styled loading screen
func renderLoadingScreen() string {
	return theme.NormalTextStyle.Render("Loading...")
}

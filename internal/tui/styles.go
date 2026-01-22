package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#06B6D4") // Cyan
	successColor   = lipgloss.Color("#10B981") // Green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F9FAFB") // White
	dimTextColor   = lipgloss.Color("#9CA3AF") // Light gray
	bgColor        = lipgloss.Color("#1F2937") // Dark gray
	highlightBg    = lipgloss.Color("#374151") // Lighter gray
)

// Styles contains all the lipgloss styles for the TUI
type Styles struct {
	// App container
	App lipgloss.Style

	// Header styles
	Header      lipgloss.Style
	HeaderTitle lipgloss.Style
	HeaderHelp  lipgloss.Style

	// Navigation tabs
	Tab          lipgloss.Style
	ActiveTab    lipgloss.Style
	TabContainer lipgloss.Style

	// Form styles
	FormLabel       lipgloss.Style
	FormValue       lipgloss.Style
	FormInput       lipgloss.Style
	FormInputActive lipgloss.Style
	FormHelp        lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemDim      lipgloss.Style

	// File browser styles
	Directory    lipgloss.Style
	File         lipgloss.Style
	SelectedItem lipgloss.Style
	CurrentPath  lipgloss.Style

	// Progress styles
	ProgressBar     lipgloss.Style
	ProgressText    lipgloss.Style
	ProgressPercent lipgloss.Style

	// Status styles
	StatusSuccess lipgloss.Style
	StatusError   lipgloss.Style
	StatusWarning lipgloss.Style
	StatusInfo    lipgloss.Style

	// Log styles
	LogContainer lipgloss.Style
	LogEntry     lipgloss.Style
	LogTimestamp lipgloss.Style
	LogSuccess   lipgloss.Style
	LogError     lipgloss.Style
	LogInfo      lipgloss.Style

	// Button styles
	Button       lipgloss.Style
	ButtonActive lipgloss.Style

	// Box styles
	Box         lipgloss.Style
	BoxTitle    lipgloss.Style
	BoxSelected lipgloss.Style

	// Footer
	Footer     lipgloss.Style
	FooterKey  lipgloss.Style
	FooterDesc lipgloss.Style

	// Spinner
	Spinner lipgloss.Style

	// General text styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Text        lipgloss.Style
	TextMuted   lipgloss.Style
	TextBold    lipgloss.Style
	TextSuccess lipgloss.Style
	TextError   lipgloss.Style
}

// DefaultStyles returns the default styling for the TUI
func DefaultStyles() Styles {
	return Styles{
		// App container
		App: lipgloss.NewStyle().
			Padding(1, 2),

		// Header styles
		Header: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 2).
			MarginBottom(1),

		HeaderTitle: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginRight(2),

		HeaderHelp: lipgloss.NewStyle().
			Foreground(dimTextColor),

		// Navigation tabs
		Tab: lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(dimTextColor).
			MarginRight(1),

		ActiveTab: lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(textColor).
			Background(primaryColor).
			Bold(true).
			MarginRight(1),

		TabContainer: lipgloss.NewStyle().
			MarginBottom(1),

		// Form styles
		FormLabel: lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Width(20),

		FormValue: lipgloss.NewStyle().
			Foreground(textColor),

		FormInput: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			Width(30),

		FormInputActive: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1).
			Width(30),

		FormHelp: lipgloss.NewStyle().
			Foreground(dimTextColor).
			Italic(true).
			MarginLeft(2),

		// List styles
		ListItem: lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(2),

		ListItemSelected: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			PaddingLeft(2),

		ListItemDim: lipgloss.NewStyle().
			Foreground(mutedColor).
			PaddingLeft(2),

		// File browser styles
		Directory: lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true),

		File: lipgloss.NewStyle().
			Foreground(textColor),

		SelectedItem: lipgloss.NewStyle().
			Background(highlightBg).
			Foreground(primaryColor).
			Bold(true).
			Padding(0, 1),

		CurrentPath: lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			MarginBottom(1),

		// Progress styles
		ProgressBar: lipgloss.NewStyle().
			Foreground(primaryColor),

		ProgressText: lipgloss.NewStyle().
			Foreground(textColor),

		ProgressPercent: lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true),

		// Status styles
		StatusSuccess: lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true),

		StatusError: lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true),

		StatusWarning: lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true),

		StatusInfo: lipgloss.NewStyle().
			Foreground(secondaryColor),

		// Log styles
		LogContainer: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1).
			MarginTop(1),

		LogEntry: lipgloss.NewStyle().
			Foreground(textColor),

		LogTimestamp: lipgloss.NewStyle().
			Foreground(dimTextColor).
			Width(10),

		LogSuccess: lipgloss.NewStyle().
			Foreground(successColor),

		LogError: lipgloss.NewStyle().
			Foreground(errorColor),

		LogInfo: lipgloss.NewStyle().
			Foreground(dimTextColor),

		// Button styles
		Button: lipgloss.NewStyle().
			Foreground(textColor).
			Background(mutedColor).
			Padding(0, 3).
			MarginRight(1),

		ButtonActive: lipgloss.NewStyle().
			Foreground(textColor).
			Background(primaryColor).
			Bold(true).
			Padding(0, 3).
			MarginRight(1),

		// Box styles
		Box: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2),

		BoxTitle: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1),

		BoxSelected: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2),

		// Footer
		Footer: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(mutedColor).
			BorderTop(true).
			MarginTop(1).
			Padding(0, 1),

		FooterKey: lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true),

		FooterDesc: lipgloss.NewStyle().
			Foreground(dimTextColor),

		// Spinner
		Spinner: lipgloss.NewStyle().
			Foreground(primaryColor),

		// General text styles
		Title: lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(dimTextColor).
			MarginBottom(1),

		Text: lipgloss.NewStyle().
			Foreground(textColor),

		TextMuted: lipgloss.NewStyle().
			Foreground(mutedColor),

		TextBold: lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true),

		TextSuccess: lipgloss.NewStyle().
			Foreground(successColor),

		TextError: lipgloss.NewStyle().
			Foreground(errorColor),
	}
}

// Helper functions for common style operations

// RenderKeyHelp renders a key binding help item
func (s Styles) RenderKeyHelp(key, description string) string {
	return s.FooterKey.Render(key) + " " + s.FooterDesc.Render(description)
}

// RenderStatus renders a status message with appropriate styling
func (s Styles) RenderStatus(status, message string) string {
	switch status {
	case "success":
		return s.StatusSuccess.Render("✓ " + message)
	case "error":
		return s.StatusError.Render("✗ " + message)
	case "warning":
		return s.StatusWarning.Render("⚠ " + message)
	default:
		return s.StatusInfo.Render("ℹ " + message)
	}
}

// RenderLogEntry renders a log entry with timestamp and level
func (s Styles) RenderLogEntry(level, message string) string {
	var levelStyle lipgloss.Style
	var prefix string

	switch level {
	case "SUCCESS":
		levelStyle = s.LogSuccess
		prefix = "[SUCCESS]"
	case "ERROR":
		levelStyle = s.LogError
		prefix = "[ERROR]  "
	case "WARNING":
		levelStyle = s.StatusWarning
		prefix = "[WARNING]"
	default:
		levelStyle = s.LogInfo
		prefix = "[INFO]   "
	}

	return levelStyle.Render(prefix) + " " + s.LogEntry.Render(message)
}

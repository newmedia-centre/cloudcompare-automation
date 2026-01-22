package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/cloudcompare-automation/internal/processor"
)

// viewWelcome renders the welcome/home screen
func (m Model) viewWelcome() string {
	s := m.styles

	// Compact logo for small terminals
	var logo string
	if m.width >= 60 && m.height >= 16 {
		logo = `
  ╔═╗┬  ┌─┐┬ ┬┌┬┐╔═╗┌─┐┌┬┐┌─┐┌─┐┬─┐┌─┐
  ║  │  │ ││ │ ││║  │ ││││├─┘├─┤├┬┘├┤
  ╚═╝┴─┘└─┘└─┘─┴┘╚═╝└─┘┴ ┴┴  ┴ ┴┴└─└─┘
       ╔═╗┬ ┬┌┬┐┌─┐┌┬┐┌─┐┌┬┐┬┌─┐┌┐┌
       ╠═╣│ │ │ │ ││││├─┤ │ ││ ││││
       ╩ ╩└─┘ ┴ └─┘┴ ┴┴ ┴ ┴ ┴└─┘┘└┘`
	} else {
		logo = "☁ CloudCompare Automation"
	}

	logoStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	title := s.Title.Copy().
		Foreground(secondaryColor).
		Render("LAS Point Cloud Processing")

	var description string
	if m.height >= 12 {
		description = s.TextMuted.Render(`
• Compute normals with MST orientation
• Poisson Surface Reconstruction
• Save CloudCompare projects (.bin)`)
	}

	// Start prompt
	startPrompt := s.ButtonActive.Copy().
		MarginTop(1).
		Render(" Press ENTER to Start ")

	// Footer
	footer := s.Footer.Render(
		s.RenderKeyHelp("enter", "start") + "  " +
			s.RenderKeyHelp("q", "quit"),
	)

	// Build content
	content := lipgloss.JoinVertical(lipgloss.Center,
		logoStyle.Render(logo),
		"",
		title,
		description,
		"",
		startPrompt,
	)

	// Center the content
	contentBox := lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left,
		contentBox,
		"",
		footer,
	)
}

// viewFileBrowser renders the directory browser screen
func (m Model) viewFileBrowser() string {
	s := m.styles

	// Header
	header := s.HeaderTitle.Render("📂 Select Directory")

	// Current path (truncate if needed)
	pathDisplay := m.currentDir
	maxPathLen := m.width - 6
	if len(pathDisplay) > maxPathLen && maxPathLen > 10 {
		pathDisplay = "..." + pathDisplay[len(pathDisplay)-maxPathLen+3:]
	}
	currentPath := s.CurrentPath.Render(pathDisplay)

	// Directory listing
	maxVisible := m.height - 8
	if maxVisible < 3 {
		maxVisible = 3
	}

	var items []string

	// Add parent directory option at top
	if m.cursor == -1 {
		items = append(items, s.SelectedItem.Render("▶ 📁 .."))
	} else {
		items = append(items, s.Directory.Render("  📁 .."))
	}

	// Calculate visible range
	startIdx := m.browseScroll
	endIdx := startIdx + maxVisible - 1 // -1 for parent dir
	if endIdx > len(m.entries) {
		endIdx = len(m.entries)
	}

	for i := startIdx; i < endIdx; i++ {
		entry := m.entries[i]
		name := entry.Name()

		// Truncate long names
		maxNameLen := m.width - 10
		if len(name) > maxNameLen && maxNameLen > 10 {
			name = name[:maxNameLen-3] + "..."
		}

		if i == m.cursor {
			items = append(items, s.SelectedItem.Render("▶ 📁 "+name))
		} else {
			items = append(items, s.Directory.Render("  📁 "+name))
		}
	}

	if len(m.entries) == 0 {
		items = append(items, s.TextMuted.Render("  (no subdirectories)"))
	}

	listing := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Scroll indicator
	scrollInfo := ""
	if len(m.entries) > maxVisible-1 {
		scrollInfo = s.TextMuted.Render(fmt.Sprintf(" [%d-%d of %d]", startIdx+1, endIdx, len(m.entries)))
	}

	// Selected info
	selectedInfo := s.StatusInfo.Render("Will use: " + m.currentDir)

	// Footer
	footer := s.Footer.Render(
		s.RenderKeyHelp("↑↓", "nav") + " " +
			s.RenderKeyHelp("enter", "open") + " " +
			s.RenderKeyHelp("←", "parent") + " " +
			s.RenderKeyHelp("space", "select") + " " +
			s.RenderKeyHelp("esc", "cancel"),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		currentPath,
		"",
		listing,
		scrollInfo,
		"",
		selectedInfo,
		"",
		footer,
	)
}

// viewParams renders the parameter configuration screen
func (m Model) viewParams() string {
	s := m.styles

	// Calculate available dimensions
	isCompact := m.height < 20
	isNarrow := m.width < 80

	// Header
	header := s.HeaderTitle.Render("⚙️  Configuration")

	// Error message if any
	errorMsg := ""
	if m.err != nil {
		errText := m.err.Error()
		maxErrLen := m.width - 10
		if len(errText) > maxErrLen && maxErrLen > 10 {
			errText = errText[:maxErrLen-3] + "..."
		}
		errorMsg = s.StatusError.Render("⚠ " + errText)
	}

	// Determine label width based on terminal width
	labelWidth := 14
	formWidth := 40
	if isNarrow {
		labelWidth = 10
		formWidth = 30
	}

	// Form fields - must match FocusedField order in model.go
	type formField struct {
		label string
		short string
		field FocusedField
	}

	fields := []formField{
		{"Input Dir", "Input", FocusInputDir},
		{"Output Dir", "Output", FocusOutputSubdir},
		{"KNN", "KNN", FocusKNN},
		{"Octree Depth", "Depth", FocusOctreeDepth},
		{"Samples/Node", "Samples", FocusSamplesPerNode},
		{"Point Weight", "Weight", FocusPointWeight},
		{"Boundary", "Bound", FocusBoundaryType},
	}

	// Determine which fields to show based on height
	maxFields := len(fields)
	if isCompact {
		maxFields = 4 // Show first 4 fields in compact mode
	}

	var formRows []string
	for i, f := range fields {
		if i >= maxFields {
			break
		}

		labelText := f.label
		if isNarrow {
			labelText = f.short
		}

		label := s.FormLabel.Copy().Width(labelWidth).Render(labelText)

		var input string
		if int(f.field) < len(m.inputs) {
			// Set width based on field type
			if f.field == FocusInputDir || f.field == FocusOutputSubdir {
				m.inputs[f.field].Width = formWidth - labelWidth
			} else {
				m.inputs[f.field].Width = 10
			}

			if m.focusedField == f.field {
				input = s.FormInputActive.Render(m.inputs[f.field].View())
			} else {
				input = s.FormInput.Render(m.inputs[f.field].View())
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left, label, input)
		formRows = append(formRows, row)
	}

	if maxFields < len(fields) {
		formRows = append(formRows, s.TextMuted.Render(fmt.Sprintf("  (+%d more in larger window)", len(fields)-maxFields)))
	}

	form := lipgloss.JoinVertical(lipgloss.Left, formRows...)

	// Start button
	var startButton string
	if m.focusedField == FocusStartButton {
		startButton = s.ButtonActive.Render(" ▶ Start Processing ")
	} else {
		startButton = s.Button.Render(" ▶ Start Processing ")
	}

	// Build left panel (form)
	leftPanel := lipgloss.JoinVertical(lipgloss.Left,
		form,
		"",
		startButton,
	)

	// Build right panel (summary) - only if wide enough
	var content string
	if !isNarrow && m.width >= 70 {
		// Summary panel
		inputDir := m.inputs[FocusInputDir].Value()
		if inputDir == "" {
			inputDir = m.selectedDir
		}

		outputSubdir := m.inputs[FocusOutputSubdir].Value()
		if outputSubdir == "" {
			outputSubdir = "Processed"
		}

		// Build full output path
		outputPath := inputDir + "/" + outputSubdir

		octreeDepth := m.inputs[FocusOctreeDepth].Value()
		if octreeDepth == "" {
			octreeDepth = "11"
		}

		// Calculate summary panel width to fill remaining space
		summaryWidth := m.width - formWidth - 8 // 8 for margins/padding
		if summaryWidth < 30 {
			summaryWidth = 30
		}

		// Wrap long paths
		wrapPath := func(path string, width int) []string {
			if len(path) <= width-2 {
				return []string{path}
			}
			var lines []string
			for len(path) > width-2 {
				lines = append(lines, path[:width-2])
				path = path[width-2:]
			}
			if len(path) > 0 {
				lines = append(lines, path)
			}
			return lines
		}

		summaryLines := []string{
			s.BoxTitle.Render("📋 Summary"),
			"",
			s.Text.Render("Input:"),
		}
		for _, line := range wrapPath(inputDir, summaryWidth) {
			summaryLines = append(summaryLines, s.StatusInfo.Render(" "+line))
		}

		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, s.Text.Render("Output:"))
		for _, line := range wrapPath(outputPath, summaryWidth) {
			summaryLines = append(summaryLines, s.StatusInfo.Render(" "+line))
		}

		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, s.Text.Render("Quality: Depth "+octreeDepth))

		// Count LAS files if possible
		if inputDir != "" {
			if count, err := countLASFiles(inputDir); err == nil && count > 0 {
				summaryLines = append(summaryLines, "")
				summaryLines = append(summaryLines, s.TextSuccess.Render(fmt.Sprintf("📁 %d LAS file(s) found", count)))
			}
		}

		summaryContent := lipgloss.JoinVertical(lipgloss.Left, summaryLines...)

		rightPanel := s.Box.Copy().
			Width(summaryWidth).
			Render(summaryContent)

		// Join panels horizontally
		content = lipgloss.JoinHorizontal(lipgloss.Top,
			leftPanel,
			"  ",
			rightPanel,
		)
	} else {
		content = leftPanel
	}

	// Browse hint
	browseHint := s.TextMuted.Render("Press 'b' to browse directories")

	// Footer
	footer := s.Footer.Render(
		s.RenderKeyHelp("tab", "next") + " " +
			s.RenderKeyHelp("b", "browse") + " " +
			s.RenderKeyHelp("enter", "start") + " " +
			s.RenderKeyHelp("esc", "back"),
	)

	// Build final view
	var parts []string
	parts = append(parts, header)
	if errorMsg != "" {
		parts = append(parts, errorMsg)
	}
	parts = append(parts, "")
	parts = append(parts, content)
	parts = append(parts, "")
	parts = append(parts, browseHint)
	parts = append(parts, "")
	parts = append(parts, footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// viewProcessing renders the processing progress screen
func (m Model) viewProcessing() string {
	s := m.styles

	// Check for celebration mode
	var celebrationLine string
	if m.IsCelebrating() {
		celebStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)
		celebrationLine = celebStyle.Render(m.GetCelebration())
	}

	// Animated header with particles and wave effect
	particle := m.GetParticles()
	waveChars := []string{"~", "≈", "~", "≈", "~"}
	waveIdx := m.animFrame % len(waveChars)
	wave := ""
	for i := 0; i < 3; i++ {
		wave += waveChars[(waveIdx+i)%len(waveChars)]
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true)

	header := headerStyle.Render(fmt.Sprintf("%s %s Processing %s %s", particle, wave, wave, particle))

	// Progress info with animated separator
	elapsed := m.elapsedTime.Round(time.Second)
	separators := []string{"│", "┃", "│", "┃"}
	sep := separators[m.animFrame%len(separators)]

	progressInfo := s.Text.Render(fmt.Sprintf(
		"Files: %d/%d %s Time: %s",
		m.filesDone, m.filesTotal, sep, elapsed,
	))

	// Progress bar
	var progressPercent float64
	if m.filesTotal > 0 {
		progressPercent = float64(m.filesDone) / float64(m.filesTotal)
	}

	barWidth := m.width - 6
	if barWidth > 60 {
		barWidth = 60
	}
	if barWidth < 20 {
		barWidth = 20
	}
	m.progress.Width = barWidth
	progressBar := m.progress.ViewAs(progressPercent)

	// Current file info box
	var fileInfoLines []string

	if m.currentFile != "" {
		// File name with animation
		display := m.currentFile
		maxLen := m.width - 15
		if len(display) > maxLen && maxLen > 10 {
			display = "..." + display[len(display)-maxLen+3:]
		}
		fileInfoLines = append(fileInfoLines, s.StatusInfo.Render("📄 "+display))

		// Point count (if loaded)
		if m.pointCount != "" {
			fileInfoLines = append(fileInfoLines, s.TextSuccess.Render("   ✓ "+m.pointCount))
		}

		// Mesh faces (if created)
		if m.meshFaces != "" {
			fileInfoLines = append(fileInfoLines, s.TextSuccess.Render("   ✓ "+m.meshFaces))
		}

		fileInfoLines = append(fileInfoLines, "")

		// Step progress visualization
		stepNames := []string{
			"Loading point cloud",
			"Computing normals",
			"Converting to DIP",
			"Poisson reconstruction",
			"Saving project",
		}

		fileInfoLines = append(fileInfoLines, s.BoxTitle.Render("📊 Pipeline Progress"))
		fileInfoLines = append(fileInfoLines, "")

		for i, name := range stepNames {
			stepNum := i + 1
			var stepLine string

			if stepNum < m.currentStepNum {
				// Completed step - green checkmark
				stepLine = s.TextSuccess.Render(fmt.Sprintf("   ✓ [%d/5] %s", stepNum, name))
			} else if stepNum == m.currentStepNum {
				// Current step - animated spinner and progress bar
				spinner := m.GetStepSpinner()
				miniProgress := m.GetStepProgress()

				stepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
				stepLine = stepStyle.Render(fmt.Sprintf("   %s [%d/5] %s", spinner, stepNum, name))
				fileInfoLines = append(fileInfoLines, stepLine)

				// Add mini progress bar for current step
				progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
				stepLine = progressStyle.Render(fmt.Sprintf("         %s", miniProgress))
			} else {
				// Future step - dimmed
				stepLine = s.TextMuted.Render(fmt.Sprintf("   ○ [%d/5] %s", stepNum, name))
			}

			fileInfoLines = append(fileInfoLines, stepLine)
		}

	} else {
		// Initializing animation with bouncing dots
		frames := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
		frame := frames[m.animFrame%len(frames)]

		// Bouncing dots animation
		dots := []string{"   ", ".  ", ".. ", "...", " ..", "  .", "   "}
		dot := dots[m.animFrame%len(dots)]

		initStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
		fileInfoLines = append(fileInfoLines, initStyle.Render(fmt.Sprintf("   %s Initializing CloudComPy%s", frame, dot)))
		fileInfoLines = append(fileInfoLines, "")

		// Loading bar animation
		loadBarWidth := 20
		loadPos := m.animFrame % (loadBarWidth * 2)
		if loadPos >= loadBarWidth {
			loadPos = loadBarWidth*2 - loadPos - 1
		}
		loadBar := ""
		for i := 0; i < loadBarWidth; i++ {
			if i == loadPos || i == loadPos+1 || i == loadPos-1 {
				loadBar += "█"
			} else {
				loadBar += "░"
			}
		}
		loadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#06B6D4"))
		fileInfoLines = append(fileInfoLines, loadStyle.Render("   "+loadBar))
		fileInfoLines = append(fileInfoLines, "")
		fileInfoLines = append(fileInfoLines, s.TextMuted.Render("   Setting up environment..."))
	}

	fileInfo := lipgloss.JoinVertical(lipgloss.Left, fileInfoLines...)

	// Log viewer
	logTitle := s.BoxTitle.Render("📜 Log")

	maxLogLines := m.height - 25 - len(fileInfoLines)
	if maxLogLines < 2 {
		maxLogLines = 2
	}

	var logLines []string
	startLog := len(m.logs) - maxLogLines
	if startLog < 0 {
		startLog = 0
	}

	for i := startLog; i < len(m.logs); i++ {
		log := m.logs[i]
		logLines = append(logLines, s.RenderLogEntry(string(log.Level), log.Message))
	}

	if len(logLines) == 0 {
		// Animated waiting message
		waitFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		waitFrame := waitFrames[m.animFrame%len(waitFrames)]
		waitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
		logLines = append(logLines, waitStyle.Render(fmt.Sprintf(" %s Waiting for output...", waitFrame)))
	}

	logContent := strings.Join(logLines, "\n")

	// Footer with subtle animation
	cancelHint := s.RenderKeyHelp("ctrl+c", "cancel")

	// Add a subtle breathing effect to the footer
	footerAccent := []string{"─", "━", "─", "━"}
	accent := footerAccent[m.animFrame%len(footerAccent)]

	footer := s.Footer.Render(
		fmt.Sprintf("%s%s%s %s", accent, accent, accent, cancelHint),
	)

	// Build final view with optional celebration
	var parts []string
	parts = append(parts, header)
	if celebrationLine != "" {
		parts = append(parts, celebrationLine)
	}
	parts = append(parts, "")
	parts = append(parts, progressInfo)
	parts = append(parts, progressBar)
	parts = append(parts, "")
	parts = append(parts, fileInfo)
	parts = append(parts, "")
	parts = append(parts, logTitle)
	parts = append(parts, logContent)
	parts = append(parts, "")
	parts = append(parts, footer)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// viewResults renders the results screen
func (m Model) viewResults() string {
	s := m.styles

	// Determine success/failure from result data
	var statusIcon, statusText string
	var statusStyle lipgloss.Style

	// Use result data if available, otherwise fall back to tracked counts
	successCount := m.filesDone
	failedCount := m.result.FailedCount
	totalFiles := m.filesTotal

	if m.result.SuccessCount > 0 {
		successCount = m.result.SuccessCount
	}
	if m.result.TotalFiles > 0 {
		totalFiles = m.result.TotalFiles
	}

	// If we have successful files and no failures, it's a success
	if successCount > 0 && failedCount == 0 {
		statusIcon = "✅"
		statusText = "Complete!"
		statusStyle = s.StatusSuccess
	} else if successCount == 0 && failedCount > 0 {
		statusIcon = "❌"
		statusText = "Failed"
		statusStyle = s.StatusError
	} else if successCount > 0 && failedCount > 0 {
		statusIcon = "⚠️"
		statusText = "Partial"
		statusStyle = s.StatusWarning
	} else {
		// No data - check if any success messages in logs
		hasSuccess := false
		for _, log := range m.logs {
			if log.Level == processor.LogSuccess && strings.Contains(log.Message, "Successfully processed") {
				hasSuccess = true
				successCount = 1
				break
			}
		}
		if hasSuccess {
			statusIcon = "✅"
			statusText = "Complete!"
			statusStyle = s.StatusSuccess
		} else {
			statusIcon = "❌"
			statusText = "Failed"
			statusStyle = s.StatusError
		}
	}

	// Header
	header := statusStyle.Copy().Bold(true).Render(statusIcon + " " + statusText)

	// Stats
	elapsed := m.elapsedTime.Round(time.Millisecond * 100)

	// Ensure we show at least 1 for total if we have any data
	if totalFiles == 0 && (successCount > 0 || failedCount > 0) {
		totalFiles = successCount + failedCount
	}
	if totalFiles == 0 {
		totalFiles = 1 // At least 1 file was attempted
	}

	stats := lipgloss.JoinVertical(lipgloss.Left,
		s.Text.Render(fmt.Sprintf("Total:      %d", totalFiles)),
		s.TextSuccess.Render(fmt.Sprintf("Success:    %d", successCount)),
		s.TextError.Render(fmt.Sprintf("Failed:     %d", failedCount)),
		s.TextMuted.Render(fmt.Sprintf("Time:       %s", elapsed)),
	)

	// Output info
	outputDir := m.params.InputDir
	if outputDir == "" || outputDir == "." {
		outputDir, _ = os.Getwd()
	}
	outputPath := fmt.Sprintf("%s/%s", outputDir, m.params.OutputSubdir)

	// Truncate path if needed
	maxPathLen := m.width - 10
	if len(outputPath) > maxPathLen && maxPathLen > 10 {
		outputPath = "..." + outputPath[len(outputPath)-maxPathLen+3:]
	}

	outputInfo := lipgloss.JoinVertical(lipgloss.Left,
		s.TextMuted.Render("Output:"),
		s.StatusInfo.Render("📂 "+outputPath),
	)

	// Recent logs (compact)
	maxLogLines := m.height - 16
	if maxLogLines < 2 {
		maxLogLines = 2
	}
	if maxLogLines > 8 {
		maxLogLines = 8
	}

	startLog := len(m.logs) - maxLogLines
	if startLog < 0 {
		startLog = 0
	}

	var logLines []string
	for i := startLog; i < len(m.logs); i++ {
		log := m.logs[i]
		logLines = append(logLines, s.RenderLogEntry(string(log.Level), log.Message))
	}

	logContent := ""
	if len(logLines) > 0 {
		logContent = strings.Join(logLines, "\n")
	}

	// Footer
	footer := s.Footer.Render(
		s.RenderKeyHelp("enter", "restart") + "  " +
			s.RenderKeyHelp("q", "quit"),
	)

	// Build view based on available space
	if m.height >= 18 && len(logLines) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			s.BoxTitle.Render("📊 Statistics"),
			stats,
			"",
			outputInfo,
			"",
			s.BoxTitle.Render("📜 Log"),
			logContent,
			"",
			footer,
		)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			stats,
			"",
			outputInfo,
			"",
			footer,
		)
	}
}

// Helper function to count LAS files in a directory
func countLASFiles(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".las") {
			count++
		}
	}
	return count, nil
}

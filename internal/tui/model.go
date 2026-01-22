package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cloudcompare-automation/internal/processor"
)

// Step-specific spinner frames for visual variety
var (
	loadingFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	normalFrames  = []string{"◐", "◓", "◑", "◒"}
	meshFrames    = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"}
	saveFrames    = []string{"💾", "📀", "💿", "📀"}
	dipFrames     = []string{"↑", "↗", "→", "↘", "↓", "↙", "←", "↖"}
	pulseFrames   = []string{"●", "◉", "○", "◉"}

	// Progress bar characters
	progressFull  = "█"
	progressEmpty = "░"

	// Particle effects for completed steps
	sparkles = []string{"✨", "⭐", "💫", "✨"}
)

// Screen represents the current view in the TUI
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenFileBrowser
	ScreenParams
	ScreenProcessing
	ScreenResults
)

// FocusedField represents which form field is currently focused
type FocusedField int

const (
	FocusInputDir FocusedField = iota
	FocusOutputSubdir
	FocusKNN
	FocusOctreeDepth
	FocusSamplesPerNode
	FocusPointWeight
	FocusBoundaryType
	FocusStartButton
	FocusFieldCount
)

// Model represents the main application state
type Model struct {
	// Current screen
	screen Screen

	// Styling
	styles Styles

	// Window dimensions
	width  int
	height int

	// File browser state
	currentDir   string
	entries      []os.DirEntry
	cursor       int
	selectedDir  string
	browseScroll int

	// Form inputs
	inputs       []textinput.Model
	focusedField FocusedField

	// Parameters
	params processor.Params

	// Processing state
	processor   *processor.Processor
	processing  bool
	logs        []processor.LogEntry
	logScroll   int
	maxLogs     int
	progress    progress.Model
	spinner     spinner.Model
	currentFile string
	currentStep string
	currentStepNum int
	pointCount  string
	meshFaces   string
	filesTotal  int
	filesDone   int
	startTime   time.Time
	elapsedTime time.Duration

	// Animation state
	animFrame    int
	animTick     int
	particlePos  int
	completedSteps []bool
	stepStartTime time.Time
	celebrating  bool
	celebrateFrame int

	// Results
	result processor.ProcessingResult

	// Error message
	err error
}

// LogMsg is sent when a new log entry is received
type LogMsg processor.LogEntry

// ProcessingDoneMsg is sent when processing completes
type ProcessingDoneMsg processor.ProcessingResult

// TickMsg is for periodic updates during processing
type TickMsg time.Time

// PollLogsMsg triggers polling for new logs
type PollLogsMsg struct{}

// AnimTickMsg triggers animation updates
type AnimTickMsg time.Time

// New creates a new Model with default settings
func New() Model {
	styles := DefaultStyles()

	// Initialize text inputs
	inputs := make([]textinput.Model, FocusFieldCount-1) // -1 because button isn't an input

	// Input directory
	inputs[FocusInputDir] = textinput.New()
	inputs[FocusInputDir].Placeholder = "Current directory"
	inputs[FocusInputDir].CharLimit = 512
	inputs[FocusInputDir].Width = 40

	// Output subdirectory
	inputs[FocusOutputSubdir] = textinput.New()
	inputs[FocusOutputSubdir].Placeholder = "Processed"
	inputs[FocusOutputSubdir].CharLimit = 256
	inputs[FocusOutputSubdir].Width = 20

	// KNN
	inputs[FocusKNN] = textinput.New()
	inputs[FocusKNN].Placeholder = "6"
	inputs[FocusKNN].CharLimit = 4
	inputs[FocusKNN].Width = 10

	// Octree depth
	inputs[FocusOctreeDepth] = textinput.New()
	inputs[FocusOctreeDepth].Placeholder = "11"
	inputs[FocusOctreeDepth].CharLimit = 4
	inputs[FocusOctreeDepth].Width = 10

	// Samples per node
	inputs[FocusSamplesPerNode] = textinput.New()
	inputs[FocusSamplesPerNode].Placeholder = "1.5"
	inputs[FocusSamplesPerNode].CharLimit = 6
	inputs[FocusSamplesPerNode].Width = 10

	// Point weight
	inputs[FocusPointWeight] = textinput.New()
	inputs[FocusPointWeight].Placeholder = "2.0"
	inputs[FocusPointWeight].CharLimit = 6
	inputs[FocusPointWeight].Width = 10

	// Boundary type
	inputs[FocusBoundaryType] = textinput.New()
	inputs[FocusBoundaryType].Placeholder = "2"
	inputs[FocusBoundaryType].CharLimit = 1
	inputs[FocusBoundaryType].Width = 10

	// Get current directory
	cwd, _ := os.Getwd()

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Initialize spinner
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = styles.Spinner

	return Model{
		screen:       ScreenWelcome,
		styles:       styles,
		currentDir:   cwd,
		selectedDir:  cwd,
		inputs:       inputs,
		focusedField: FocusInputDir,
		params:       processor.DefaultParams(),
		maxLogs:      500,
		logs:         make([]processor.LogEntry, 0),
		progress:     prog,
		spinner:      spin,
		width:        80,
		height:       24,
		completedSteps: make([]bool, 5),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadDirectory(m.currentDir),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handlers
		switch msg.String() {
		case "ctrl+c":
			if m.processing {
				// Cancel processing
				if m.processor != nil {
					m.processor.Stop()
				}
				m.processing = false
				m.elapsedTime = time.Since(m.startTime)
				return m, nil
			}
			return m, tea.Quit

		case "q":
			if !m.processing && m.screen != ScreenParams {
				return m, tea.Quit
			}

		case "esc":
			switch m.screen {
			case ScreenFileBrowser:
				m.screen = ScreenParams
				return m, nil
			case ScreenParams:
				m.screen = ScreenWelcome
				return m, nil
			case ScreenResults:
				m.screen = ScreenWelcome
				return m, nil
			}
		}

		// Screen-specific key handlers
		switch m.screen {
		case ScreenWelcome:
			return m.updateWelcome(msg)
		case ScreenFileBrowser:
			return m.updateFileBrowser(msg)
		case ScreenParams:
			return m.updateParams(msg)
		case ScreenProcessing:
			return m.updateProcessing(msg)
		case ScreenResults:
			return m.updateResults(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = min(m.width-20, 60)
		return m, nil

	case spinner.TickMsg:
		if m.processing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case AnimTickMsg:
		if m.processing {
			m.animTick++
			m.animFrame = (m.animFrame + 1) % 60
			m.particlePos = (m.particlePos + 1) % 20

			// Celebration animation counter
			if m.celebrating {
				m.celebrateFrame++
				if m.celebrateFrame > 30 {
					m.celebrating = false
					m.celebrateFrame = 0
				}
			}

			return m, tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
				return AnimTickMsg(t)
			})
		}
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case LogMsg:
		m.logs = append(m.logs, processor.LogEntry(msg))
		if len(m.logs) > m.maxLogs {
			m.logs = m.logs[1:]
		}
		// Auto-scroll to bottom
		m.logScroll = len(m.logs) - 1

		// Check for file processing indicators
		if strings.Contains(msg.Message, "Processing:") {
			m.currentFile = strings.TrimPrefix(msg.Message, "Processing: ")
			m.currentFile = strings.TrimSpace(m.currentFile)
		}
		if msg.Level == processor.LogSuccess && strings.Contains(msg.Message, "Successfully processed:") {
			m.filesDone++
		}

		return m, nil

	case PollLogsMsg:
		// Poll for logs from processor
		if m.processor == nil {
			return m, nil
		}

		var newLogs []processor.LogEntry
		// Drain all available logs
		for {
			select {
			case log, ok := <-m.processor.LogChan():
				if !ok {
					// Channel closed
					return m, nil
				}
				newLogs = append(newLogs, log)
			default:
				// No more logs available
				goto done
			}
		}
	done:
		// Process all new logs
		for _, log := range newLogs {
			m.logs = append(m.logs, log)
			if len(m.logs) > m.maxLogs {
				m.logs = m.logs[1:]
			}
			m.logScroll = len(m.logs) - 1

			// Track current file
			if strings.Contains(log.Message, "Processing:") {
				m.currentFile = strings.TrimPrefix(log.Message, "Processing: ")
				m.currentFile = strings.TrimSpace(m.currentFile)
				// Reset stats for new file
				m.currentStep = ""
				m.pointCount = ""
				m.meshFaces = ""
			}

			// Track current step [1/5], [2/5], etc.
			if strings.Contains(log.Message, "[") && strings.Contains(log.Message, "/5]") {
				// Extract step info like "[1/5] Loading point cloud..."
				m.currentStep = log.Message
				m.stepStartTime = time.Now()

				// Parse step number
				if strings.Contains(log.Message, "[1/5]") {
					m.currentStepNum = 1
				} else if strings.Contains(log.Message, "[2/5]") {
					m.currentStepNum = 2
					m.completedSteps[0] = true
				} else if strings.Contains(log.Message, "[3/5]") {
					m.currentStepNum = 3
					m.completedSteps[1] = true
				} else if strings.Contains(log.Message, "[4/5]") {
					m.currentStepNum = 4
					m.completedSteps[2] = true
				} else if strings.Contains(log.Message, "[5/5]") {
					m.currentStepNum = 5
					m.completedSteps[3] = true
				}
			}

			// Track point count
			if strings.Contains(log.Message, "Loaded") && strings.Contains(log.Message, "points") {
				m.pointCount = log.Message
			}

			// Track mesh faces
			if strings.Contains(log.Message, "Mesh created with") {
				m.meshFaces = log.Message
			}

			if log.Level == processor.LogSuccess && strings.Contains(log.Message, "Successfully processed:") {
				m.filesDone++
				m.completedSteps[4] = true
				m.celebrating = true
				m.celebrateFrame = 0
			}
		}

		// Keep polling if still processing
		if m.processing {
			return m, tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
				return PollLogsMsg{}
			})
		}
		return m, nil

	case ProcessingDoneMsg:
		m.processing = false
		m.elapsedTime = time.Since(m.startTime)
		m.result = processor.ProcessingResult(msg)

		// Final drain of logs
		if m.processor != nil {
			for {
				select {
				case log, ok := <-m.processor.LogChan():
					if !ok {
						goto finaldone
					}
					m.logs = append(m.logs, log)
					if log.Level == processor.LogSuccess && strings.Contains(log.Message, "Successfully processed:") {
						m.filesDone++
					}
				default:
					goto finaldone
				}
			}
		}
	finaldone:

		// Use the result's success count if available
		if m.result.SuccessCount > 0 {
			m.filesDone = m.result.SuccessCount
		}
		if m.result.TotalFiles > 0 {
			m.filesTotal = m.result.TotalFiles
		}

		m.screen = ScreenResults
		return m, nil

	case TickMsg:
		if m.processing {
			m.elapsedTime = time.Since(m.startTime)
			return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return TickMsg(t)
			})
		}
		return m, nil

	case directoryLoadedMsg:
		m.entries = msg.entries
		m.cursor = 0
		m.browseScroll = 0
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	switch m.screen {
	case ScreenWelcome:
		return m.viewWelcome()
	case ScreenFileBrowser:
		return m.viewFileBrowser()
	case ScreenParams:
		return m.viewParams()
	case ScreenProcessing:
		return m.viewProcessing()
	case ScreenResults:
		return m.viewResults()
	default:
		return "Unknown screen"
	}
}

// Screen update functions

func (m Model) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		m.screen = ScreenParams
		m.inputs[FocusInputDir].SetValue(m.selectedDir)
		m.inputs[FocusInputDir].Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m Model) updateFileBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxVisible := m.height - 10
	if maxVisible < 3 {
		maxVisible = 3
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.browseScroll {
				m.browseScroll = m.cursor
			}
		}

	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
			if m.cursor >= m.browseScroll+maxVisible {
				m.browseScroll = m.cursor - maxVisible + 1
			}
		}

	case "enter", "right", "l":
		if len(m.entries) > 0 && m.cursor < len(m.entries) {
			entry := m.entries[m.cursor]
			if entry.IsDir() {
				newPath := filepath.Join(m.currentDir, entry.Name())
				m.currentDir = newPath
				return m, m.loadDirectory(newPath)
			}
		}

	case "backspace", "h", "left":
		parent := filepath.Dir(m.currentDir)
		if parent != m.currentDir {
			m.currentDir = parent
			return m, m.loadDirectory(parent)
		}

	case "s", " ":
		// Select current directory
		m.selectedDir = m.currentDir
		m.inputs[FocusInputDir].SetValue(m.selectedDir)
		m.screen = ScreenParams
		return m, nil
	}

	return m, nil
}

func (m Model) updateParams(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.focusedField = (m.focusedField + 1) % FocusFieldCount
		return m, m.updateFocus()

	case "shift+tab", "up":
		m.focusedField = (m.focusedField - 1 + FocusFieldCount) % FocusFieldCount
		return m, m.updateFocus()

	case "enter":
		if m.focusedField == FocusStartButton {
			return m.startProcessing()
		}
		// Move to next field
		m.focusedField = (m.focusedField + 1) % FocusFieldCount
		return m, m.updateFocus()

	case "b", "ctrl+b":
		// Open file browser
		m.screen = ScreenFileBrowser
		return m, m.loadDirectory(m.currentDir)

	case "ctrl+v":
		// Paste from clipboard
		if int(m.focusedField) < len(m.inputs) {
			if text, err := clipboard.ReadAll(); err == nil {
				// Clean up pasted text (remove newlines, trim)
				text = strings.TrimSpace(text)
				text = strings.ReplaceAll(text, "\n", "")
				text = strings.ReplaceAll(text, "\r", "")
				m.inputs[m.focusedField].SetValue(text)
				return m, nil
			}
		}
		return m, nil
	}

	// Update the focused input
	if int(m.focusedField) < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusedField], cmd = m.inputs[m.focusedField].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) updateProcessing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C is handled globally
	return m, nil
}

func (m Model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "r":
		// Reset and go back to welcome
		m.screen = ScreenWelcome
		m.logs = make([]processor.LogEntry, 0)
		m.filesDone = 0
		m.currentFile = ""
		m.err = nil
		return m, nil
	}
	return m, nil
}

// Helper functions

func (m Model) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == int(m.focusedField) {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m Model) startProcessing() (tea.Model, tea.Cmd) {
	// Parse parameters from inputs
	m.params.InputDir = m.inputs[FocusInputDir].Value()
	if m.params.InputDir == "" {
		m.params.InputDir = m.selectedDir
	}

	m.params.OutputSubdir = m.inputs[FocusOutputSubdir].Value()
	if m.params.OutputSubdir == "" {
		m.params.OutputSubdir = "Processed"
	}

	// Parse numeric values with defaults
	m.params.KNN = 6
	fmt.Sscanf(m.inputs[FocusKNN].Value(), "%d", &m.params.KNN)
	if m.params.KNN <= 0 {
		m.params.KNN = 6
	}

	m.params.OctreeDepth = 11
	fmt.Sscanf(m.inputs[FocusOctreeDepth].Value(), "%d", &m.params.OctreeDepth)
	if m.params.OctreeDepth <= 0 {
		m.params.OctreeDepth = 11
	}

	m.params.SamplesPerNode = 1.5
	fmt.Sscanf(m.inputs[FocusSamplesPerNode].Value(), "%f", &m.params.SamplesPerNode)
	if m.params.SamplesPerNode <= 0 {
		m.params.SamplesPerNode = 1.5
	}

	m.params.PointWeight = 2.0
	fmt.Sscanf(m.inputs[FocusPointWeight].Value(), "%f", &m.params.PointWeight)
	if m.params.PointWeight <= 0 {
		m.params.PointWeight = 2.0
	}

	m.params.BoundaryType = 2
	fmt.Sscanf(m.inputs[FocusBoundaryType].Value(), "%d", &m.params.BoundaryType)
	if m.params.BoundaryType < 0 || m.params.BoundaryType > 2 {
		m.params.BoundaryType = 2
	}

	// Create processor
	m.processor = processor.New(m.params)

	// Validate
	if err := m.processor.ValidateInputDir(); err != nil {
		m.err = err
		return m, nil
	}

	// Count files
	count, _ := m.processor.CountLASFiles()
	m.filesTotal = count
	m.filesDone = 0

	// Find scripts
	if err := m.processor.FindScripts(); err != nil {
		m.err = err
		return m, nil
	}

	// Start processing
	m.processing = true
	m.screen = ScreenProcessing
	m.startTime = time.Now()
	m.elapsedTime = 0
	m.logs = make([]processor.LogEntry, 0)
	m.currentFile = ""
	m.currentStep = ""
	m.currentStepNum = 0
	m.pointCount = ""
	m.meshFaces = ""
	m.animFrame = 0
	m.animTick = 0
	m.particlePos = 0
	m.completedSteps = make([]bool, 5)
	m.celebrating = false
	m.celebrateFrame = 0
	m.err = nil

	if err := m.processor.Start(); err != nil {
		m.err = err
		m.processing = false
		return m, nil
	}

	return m, tea.Batch(
		m.spinner.Tick,
		m.listenForResult(),
		// Start polling for logs
		tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
			return PollLogsMsg{}
		}),
		// Start tick for elapsed time
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
		// Start animation tick
		tea.Tick(time.Millisecond*80, func(t time.Time) tea.Msg {
			return AnimTickMsg(t)
		}),
	)
}

func (m Model) listenForResult() tea.Cmd {
	return func() tea.Msg {
		if m.processor == nil {
			return nil
		}
		result := <-m.processor.ResultChan()
		return ProcessingDoneMsg(result)
	}
}

type directoryLoadedMsg struct {
	entries []os.DirEntry
	err     error
}

func (m Model) loadDirectory(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(path)
		// Filter to only show directories
		var dirs []os.DirEntry
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				dirs = append(dirs, e)
			}
		}
		return directoryLoadedMsg{entries: dirs, err: err}
	}
}

// GetElapsedTime returns the elapsed time (either running or final)
func (m Model) GetElapsedTime() time.Duration {
	return m.elapsedTime
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetStepSpinner returns an animated spinner for the current step
func (m Model) GetStepSpinner() string {
	switch m.currentStepNum {
	case 1:
		return loadingFrames[m.animFrame%len(loadingFrames)]
	case 2:
		return normalFrames[m.animFrame%len(normalFrames)]
	case 3:
		return dipFrames[m.animFrame%len(dipFrames)]
	case 4:
		return meshFrames[m.animFrame%len(meshFrames)]
	case 5:
		return saveFrames[m.animFrame%len(saveFrames)]
	default:
		return pulseFrames[m.animFrame%len(pulseFrames)]
	}
}

// GetStepProgress returns a mini progress bar for the current step
func (m Model) GetStepProgress() string {
	if m.currentStepNum == 0 {
		return ""
	}

	// Animated progress based on time in current step
	elapsed := time.Since(m.stepStartTime).Seconds()

	// Different expected durations per step
	var expectedDuration float64
	switch m.currentStepNum {
	case 1:
		expectedDuration = 5.0
	case 2:
		expectedDuration = 60.0
	case 3:
		expectedDuration = 2.0
	case 4:
		expectedDuration = 300.0 // Poisson takes long
	case 5:
		expectedDuration = 10.0
	default:
		expectedDuration = 30.0
	}

	// Calculate progress (cap at 95% to show it's still running)
	progress := elapsed / expectedDuration
	if progress > 0.95 {
		progress = 0.95
	}

	// Add a pulsing effect
	pulse := float64(m.animFrame%10) / 10.0 * 0.05
	progress += pulse
	if progress > 0.99 {
		progress = 0.99
	}

	barWidth := 15
	filled := int(progress * float64(barWidth))

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += progressFull
		} else {
			bar += progressEmpty
		}
	}

	return bar
}

// GetParticles returns sparkle particles for visual flair
func (m Model) GetParticles() string {
	if !m.processing {
		return ""
	}
	idx := m.particlePos % len(sparkles)
	return sparkles[idx]
}

// GetStepStatusLine returns a formatted status line for a step
func (m Model) GetStepStatusLine(stepNum int, stepName string, style lipgloss.Style) string {
	if stepNum < m.currentStepNum {
		// Completed step
		return style.Render(fmt.Sprintf("  ✓ Step %d: %s", stepNum, stepName))
	} else if stepNum == m.currentStepNum {
		// Current step with animation
		spinner := m.GetStepSpinner()
		return style.Render(fmt.Sprintf("  %s Step %d: %s", spinner, stepNum, stepName))
	} else {
		// Future step
		return style.Render(fmt.Sprintf("  ○ Step %d: %s", stepNum, stepName))
	}
}

// IsCelebrating returns true if we're showing a completion celebration
func (m Model) IsCelebrating() bool {
	return m.celebrating
}

// GetCelebration returns celebration animation frames
func (m Model) GetCelebration() string {
	if !m.celebrating {
		return ""
	}

	celebrations := []string{
		"🎉 ✨ File Complete! ✨ 🎉",
		"✨ 🎉 File Complete! 🎉 ✨",
		"🎊 ✨ File Complete! ✨ 🎊",
		"✨ 🎊 File Complete! 🎊 ✨",
		"⭐ 🎉 File Complete! 🎉 ⭐",
		"🎉 ⭐ File Complete! ⭐ 🎉",
	}

	return celebrations[m.celebrateFrame%len(celebrations)]
}

// GetCelebrationBorder returns an animated border for celebration
func (m Model) GetCelebrationBorder() string {
	if !m.celebrating {
		return ""
	}

	borders := []string{
		"═══════════════════════════════",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		"═══════════════════════════════",
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
	}

	return borders[m.celebrateFrame%len(borders)]
}

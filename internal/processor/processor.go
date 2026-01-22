package processor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LogInfo    LogLevel = "INFO"
	LogSuccess LogLevel = "SUCCESS"
	LogWarning LogLevel = "WARNING"
	LogError   LogLevel = "ERROR"
)

// LogEntry represents a single log message from the processor
type LogEntry struct {
	Level   LogLevel
	Message string
}

// FileResult represents the processing result for a single file
type FileResult struct {
	InputFile  string
	OutputFile string
	Success    bool
	Error      string
}

// ProcessingResult contains the final results of batch processing
type ProcessingResult struct {
	TotalFiles   int
	SuccessCount int
	FailedCount  int
	OutputDir    string
	Completed    bool
}

// Params holds all configuration parameters for processing
type Params struct {
	InputDir       string
	OutputSubdir   string
	KNN            int
	OctreeDepth    int
	SamplesPerNode float64
	PointWeight    float64
	BoundaryType   int
}

// DefaultParams returns the default processing parameters
func DefaultParams() Params {
	return Params{
		InputDir:       ".",
		OutputSubdir:   "Processed",
		KNN:            6,
		OctreeDepth:    11,
		SamplesPerNode: 1.5,
		PointWeight:    2.0,
		BoundaryType:   2,
	}
}

// Processor handles the execution of the CloudComPy processing script
type Processor struct {
	params     Params
	scriptPath string
	batPath    string
	scriptDir  string

	// Channels for communication
	logChan    chan LogEntry
	resultChan chan ProcessingResult

	// State
	running      bool
	mu           sync.Mutex
	cmd          *exec.Cmd
	successCount int
	failedCount  int
}

// New creates a new Processor instance
func New(params Params) *Processor {
	return &Processor{
		params:     params,
		logChan:    make(chan LogEntry, 500),
		resultChan: make(chan ProcessingResult, 1),
	}
}

// SetParams updates the processing parameters
func (p *Processor) SetParams(params Params) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.params = params
}

// GetParams returns the current processing parameters
func (p *Processor) GetParams() Params {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.params
}

// LogChan returns the channel for receiving log entries
func (p *Processor) LogChan() <-chan LogEntry {
	return p.logChan
}

// ResultChan returns the channel for receiving the final result
func (p *Processor) ResultChan() <-chan ProcessingResult {
	return p.resultChan
}

// IsRunning returns whether the processor is currently running
func (p *Processor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// FindScripts locates the Python script and batch file
func (p *Processor) FindScripts() error {
	// Get the executable's directory
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	execDir := filepath.Dir(execPath)

	// Get current working directory
	cwd, _ := os.Getwd()

	// Possible locations to search
	searchPaths := []string{
		cwd,
		execDir,
		filepath.Join(execDir, ".."),
		filepath.Join(execDir, "..", ".."),
		filepath.Join(cwd, ".."),
	}

	// Look for the Python script
	for _, basePath := range searchPaths {
		scriptPath := filepath.Join(basePath, "process_las_files.py")
		if _, err := os.Stat(scriptPath); err == nil {
			p.scriptPath, _ = filepath.Abs(scriptPath)
			p.scriptDir = filepath.Dir(p.scriptPath)
			break
		}
	}

	// Look for the batch file (Windows only)
	if runtime.GOOS == "windows" {
		for _, basePath := range searchPaths {
			batPath := filepath.Join(basePath, "run_cloudcompy.bat")
			if _, err := os.Stat(batPath); err == nil {
				p.batPath, _ = filepath.Abs(batPath)
				break
			}
		}
	}

	if p.scriptPath == "" {
		return fmt.Errorf("could not find process_las_files.py")
	}

	return nil
}

// CountLASFiles counts the number of LAS files in the input directory
func (p *Processor) CountLASFiles() (int, error) {
	inputDir := p.params.InputDir
	if inputDir == "" {
		inputDir = "."
	}

	// Make sure it's absolute
	absDir, err := filepath.Abs(inputDir)
	if err != nil {
		return 0, err
	}

	count := 0
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return 0, err
	}

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

// Start begins the processing in a goroutine
func (p *Processor) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("processor is already running")
	}
	p.running = true
	p.successCount = 0
	p.failedCount = 0
	p.mu.Unlock()

	// Find scripts if not already found
	if p.scriptPath == "" {
		if err := p.FindScripts(); err != nil {
			p.sendLog(LogError, err.Error())
			p.mu.Lock()
			p.running = false
			p.mu.Unlock()
			return err
		}
	}

	go p.run()
	return nil
}

// Stop attempts to stop the running process
func (p *Processor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
	p.running = false
}

func (p *Processor) run() {
	defer func() {
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}()

	// Get absolute input directory
	inputDir := p.params.InputDir
	if inputDir == "" || inputDir == "." {
		inputDir, _ = os.Getwd()
	}
	absInputDir, _ := filepath.Abs(inputDir)

	p.sendLog(LogInfo, fmt.Sprintf("Input: %s", absInputDir))

	// Build command arguments for the Python script
	args := p.buildArgs(absInputDir)

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" && p.batPath != "" {
		// On Windows, use the batch file wrapper
		// The batch file handles conda activation and environment setup
		allArgs := append([]string{"/c", p.batPath}, args...)
		cmd = exec.Command("cmd", allArgs...)
		p.sendLog(LogInfo, "Starting CloudComPy processing...")
	} else {
		// Direct Python execution (requires CloudComPy in PATH)
		allArgs := append([]string{p.scriptPath}, args...)
		cmd = exec.Command("python", allArgs...)
		p.sendLog(LogInfo, fmt.Sprintf("Running: python %s", p.scriptPath))
	}

	// Set environment
	cmd.Env = os.Environ()

	p.mu.Lock()
	p.cmd = cmd
	p.mu.Unlock()

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.sendLog(LogError, fmt.Sprintf("Failed to create stdout pipe: %v", err))
		p.sendResult(ProcessingResult{Completed: true, FailedCount: 1})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		p.sendLog(LogError, fmt.Sprintf("Failed to create stderr pipe: %v", err))
		p.sendResult(ProcessingResult{Completed: true, FailedCount: 1})
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		p.sendLog(LogError, fmt.Sprintf("Failed to start process: %v", err))
		p.sendResult(ProcessingResult{Completed: true, FailedCount: 1})
		return
	}

	// Read output in separate goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		p.readOutput(stdout)
	}()

	go func() {
		defer wg.Done()
		p.readOutput(stderr)
	}()

	// Wait for output reading to complete (this ensures all logs are captured)
	wg.Wait()

	// Wait for command to finish
	exitErr := cmd.Wait()

	// Small delay to ensure all logs are processed
	// (the channel should have all messages by now)

	// Determine result based on tracked success/fail counts
	p.mu.Lock()
	successCount := p.successCount
	failedCount := p.failedCount
	p.mu.Unlock()

	result := ProcessingResult{
		Completed:    true,
		SuccessCount: successCount,
		FailedCount:  failedCount,
		TotalFiles:   successCount + failedCount,
	}

	// If we have no counts but exit was clean, assume success
	if exitErr == nil && successCount == 0 && failedCount == 0 {
		// Check if we just didn't track properly, look at exit code
		result.SuccessCount = 1
		result.TotalFiles = 1
	}

	// If exit error and no tracked failures, mark as failed
	if exitErr != nil && failedCount == 0 {
		p.sendLog(LogError, fmt.Sprintf("Process exited with error: %v", exitErr))
		result.FailedCount = 1
		if result.TotalFiles == 0 {
			result.TotalFiles = 1
		}
	}

	p.sendResult(result)
}

func (p *Processor) buildArgs(absInputDir string) []string {
	args := []string{}

	// Input directory (always first positional argument)
	args = append(args, absInputDir)

	// Output subdirectory
	if p.params.OutputSubdir != "" && p.params.OutputSubdir != "Processed" {
		args = append(args, "--output-dir", p.params.OutputSubdir)
	}

	// KNN parameter
	if p.params.KNN != 6 && p.params.KNN > 0 {
		args = append(args, "--knn", fmt.Sprintf("%d", p.params.KNN))
	}

	// Octree depth
	if p.params.OctreeDepth != 11 && p.params.OctreeDepth > 0 {
		args = append(args, "--octree-depth", fmt.Sprintf("%d", p.params.OctreeDepth))
	}

	// Samples per node
	if p.params.SamplesPerNode != 1.5 && p.params.SamplesPerNode > 0 {
		args = append(args, "--samples-per-node", fmt.Sprintf("%.1f", p.params.SamplesPerNode))
	}

	// Point weight
	if p.params.PointWeight != 2.0 && p.params.PointWeight > 0 {
		args = append(args, "--point-weight", fmt.Sprintf("%.1f", p.params.PointWeight))
	}

	// Boundary type
	if p.params.BoundaryType != 2 && p.params.BoundaryType >= 0 && p.params.BoundaryType <= 2 {
		args = append(args, "--boundary-type", fmt.Sprintf("%d", p.params.BoundaryType))
	}

	return args
}

func (p *Processor) readOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	// Regex patterns for parsing output
	levelRegex := regexp.MustCompile(`^\[(\w+)\]\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Skip separator lines
		if strings.HasPrefix(line, "===") || strings.HasPrefix(line, "---") {
			continue
		}

		// Parse the log level from the line
		matches := levelRegex.FindStringSubmatch(line)
		if matches != nil {
			level := LogLevel(strings.ToUpper(matches[1]))
			message := matches[2]

			// Track success/failure for files
			if level == LogSuccess && strings.Contains(message, "Successfully processed:") {
				p.mu.Lock()
				p.successCount++
				p.mu.Unlock()
			}
			if level == LogError && (strings.Contains(message, "Failed to") || strings.Contains(message, "failed")) {
				p.mu.Lock()
				p.failedCount++
				p.mu.Unlock()
			}

			switch level {
			case LogSuccess, LogError, LogWarning, LogInfo:
				p.sendLog(level, message)
			default:
				p.sendLog(LogInfo, message)
			}
		} else {
			// No level prefix, treat as info
			p.sendLog(LogInfo, line)
		}
	}
}

func (p *Processor) sendLog(level LogLevel, message string) {
	select {
	case p.logChan <- LogEntry{Level: level, Message: message}:
	default:
		// Channel full, drop oldest and add new
		select {
		case <-p.logChan:
		default:
		}
		select {
		case p.logChan <- LogEntry{Level: level, Message: message}:
		default:
		}
	}
}

func (p *Processor) sendResult(result ProcessingResult) {
	select {
	case p.resultChan <- result:
	default:
	}
}

// ValidateInputDir checks if the input directory exists and contains LAS files
func (p *Processor) ValidateInputDir() error {
	inputDir := p.params.InputDir
	if inputDir == "" {
		inputDir = "."
	}

	absDir, err := filepath.Abs(inputDir)
	if err != nil {
		return fmt.Errorf("invalid input directory: %s", inputDir)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("input directory does not exist: %s", absDir)
	}

	if !info.IsDir() {
		return fmt.Errorf("input path is not a directory: %s", absDir)
	}

	count, err := p.CountLASFiles()
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("no LAS files found in: %s", absDir)
	}

	return nil
}

// GetBoundaryTypeName returns the human-readable name for a boundary type
func GetBoundaryTypeName(boundaryType int) string {
	switch boundaryType {
	case 0:
		return "Free"
	case 1:
		return "Dirichlet"
	case 2:
		return "Neumann"
	default:
		return "Unknown"
	}
}

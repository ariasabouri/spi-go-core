package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI state to manage the visibility of debug messages
var showDebug = false
var mutex = &sync.Mutex{}

// StartUI starts the terminal-based UI
func StartUI() {
	// Create application
	app := tview.NewApplication()

	// Create the main frame to show context metadata and progress
	frame := tview.NewFrame(tview.NewBox().SetBorder(true)).
		AddText("Server Profile Installer", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("Task: Initializing Environment", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("Status: Waiting for input...", false, tview.AlignLeft, tcell.ColorWhite).
		AddText("[Press Y to Confirm, N to Cancel, D to Toggle Debug]", false, tview.AlignLeft, tcell.ColorWhite)

	// Create the text view to display subprocess output
	outputView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Store output in a buffer for toggling debug visibility
	var outputBuffer []string

	// Create a flex layout to arrange the frame and output view
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(frame, 10, 1, false).
		AddItem(outputView, 0, 1, true)

	// Define keybindings for confirmation and toggling debug visibility
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y', 'Y':
			fmt.Fprintln(outputView, "[green]Confirmation received: Proceeding...")
			go executeSubprocess(outputView, &outputBuffer)
		case 'n', 'N':
			fmt.Fprintln(outputView, "[red]Process canceled.")
		case 'd', 'D':
			// Toggle debug visibility
			mutex.Lock()
			showDebug = !showDebug
			mutex.Unlock()
			updateDisplay(outputView, outputBuffer)
		}
		return event
	})

	// Set the UI in motion
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatalf("Failed to run UI: %v", err)
	}
}

// executeSubprocess runs a subprocess and logs every line of output to both a file and optionally the UI
func executeSubprocess(outputView *tview.TextView, outputBuffer *[]string) {
	// Open or create a log file for subprocess output
	logFile, err := os.OpenFile("subprocess.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(outputView, "[red]Error: Failed to open log file")
		return
	}
	defer logFile.Close()

	// Example subprocess (install Node.js)
	cmd := exec.Command("bash", "-c", "curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs")

	// Capture both stdout and stderr for logging
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(outputView, "[red]Error: Failed to get stdout pipe")
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(outputView, "[red]Error: Failed to get stderr pipe")
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(outputView, "[red]Error: Failed to start command")
		return
	}

	// Use a separate goroutine to log stdout to the file
	go logOutput(stdout, logFile, outputView, outputBuffer)

	// Use a separate goroutine to log stderr to the file
	go logOutput(stderr, logFile, outputView, outputBuffer)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Fprintln(logFile, "[red]Error: Command execution failed")
	} else {
		fmt.Fprintln(logFile, "[green]Success: Node.js installation complete")
	}
}

// logOutput reads from the subprocess pipe and writes to both the log file and optionally the UI
func logOutput(pipe io.ReadCloser, logFile *os.File, outputView *tview.TextView, outputBuffer *[]string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Log everything to the file
		fmt.Fprintln(logFile, line)

		// Optionally store in buffer for display in UI
		mutex.Lock()
		*outputBuffer = append(*outputBuffer, line)
		mutex.Unlock()

		// Update the display with relevant lines based on debug mode
		updateDisplay(outputView, *outputBuffer)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(logFile, "[red]Error reading from pipe: ", err)
	}
}

// updateDisplay updates the text view based on the current state (showDebug toggle)
func updateDisplay(outputView *tview.TextView, outputBuffer []string) {
	// Clear the current output
	outputView.Clear()

	// Re-display only the relevant lines (hide debug lines if showDebug is false)
	for _, line := range outputBuffer {
		if strings.Contains(line, "Debug:") && !showDebug {
			// Skip debug lines if debug mode is not active
			continue
		}
		// Display the line
		fmt.Fprintln(outputView, line)
	}
}

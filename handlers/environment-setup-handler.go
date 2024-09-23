package handlers

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"spi-go-core/internal/config"
	"strings"
)

func handleSetup(w http.ResponseWriter, r *http.Request) {
	logFile, err := os.OpenFile("install.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()
	EnvironmentSetupHandler(logFile)
}

// EnvironmentSetupHandler handles the environment setup tasks with filtered UI output and full terminal logging
func EnvironmentSetupHandler(logFile *os.File) error {
	// Example subprocess (install Node.js)
	cmd := exec.Command("bash", "-c", "curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs")

	// Capture both stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Log stdout and stderr to terminal and file (all messages)
	go logOutput(stdout, os.Stdout, logFile)
	go logOutput(stderr, os.Stdout, logFile)

	// Also capture relevant messages for UI display (errors, warnings, and key updates)
	go captureRelevantOutput(stdout, "stdout")
	go captureRelevantOutput(stderr, "stderr")

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		log.Printf("Error: Command execution failed: %v", err)
		fmt.Fprintf(logFile, "Error: Command execution failed: %v\n", err)
		return err
	}

	log.Println("Node.js installation completed successfully.")
	fmt.Fprintln(logFile, "Node.js installation completed successfully.")
	return nil
}

// logOutput logs the output from a pipe to both the terminal and the log file (full logging)
func logOutput(pipe io.ReadCloser, terminalOutput io.Writer, logFile *os.File) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		// Log everything to the terminal and log file
		fmt.Fprintln(terminalOutput, line)
		fmt.Fprintln(logFile, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(logFile, "[Error] Failed to read pipe: %v\n", err)
	}
}

// captureRelevantOutput filters relevant messages (errors, warnings, key progress) for UI display
func captureRelevantOutput(pipe io.ReadCloser, pipeType string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Show only relevant messages (errors, warnings, or important messages)
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "warning") ||
			strings.Contains(strings.ToLower(line), "done") ||
			strings.Contains(strings.ToLower(line), "progress") {

			// Display relevant lines in the UI (or store for UI display)
			log.Printf("[%s] %s", pipeType, line) // Placeholder for where UI would be used
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[Error] Failed to read pipe for UI: %v", err)
	}
}

// detectDistro detects the Linux distribution by reading /etc/os-release
func detectDistro() (string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("failed to open /etc/os-release: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			distro := strings.TrimPrefix(line, "ID=")
			return strings.Trim(distro, "\""), nil
		}
	}
	return "", fmt.Errorf("distribution not found")
}

// installNode installs Node.js based on the detected distribution and filters output based on verbosity level
func installNode(distro string) error {
	var cmd *exec.Cmd

	// Select the right command based on the distribution and add debug/verbose flags if necessary
	switch distro {
	case "debian", "ubuntu":
		// Add --verbose flag if debug is enabled
		if config.GlobalConfig.Logging.Verbosity == "debug" {
			cmd = exec.Command("bash", "-c", "curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs --verbose")
		} else {
			cmd = exec.Command("bash", "-c", "curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs")
		}
	case "centos", "rhel", "fedora":
		if config.GlobalConfig.Logging.Verbosity == "debug" {
			cmd = exec.Command("bash", "-c", "curl -fsSL https://rpm.nodesource.com/setup_16.x | bash - && yum install -y nodejs --verbose")
		} else {
			cmd = exec.Command("bash", "-c", "curl -fsSL https://rpm.nodesource.com/setup_16.x | bash - && yum install -y nodejs")
		}
	default:
		return fmt.Errorf("unsupported distribution: %s", distro)
	}

	// Get pipe to capture the output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Filter and log the output based on verbosity level, including debug if set
	go filterOutput(stdout, "stdout")
	go filterOutput(stderr, "stderr")

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	log.Println("Node.js installation completed successfully.")
	return nil
}

// filterOutput filters and logs output based on verbosity level, including debug
func filterOutput(pipe io.ReadCloser, pipeType string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Show more or less based on the verbosity level
		switch config.GlobalConfig.Logging.Verbosity {
		case "verbose":
			// Log everything
			log.Printf("[%s] %s", pipeType, line)

		case "normal":
			// Filter for errors, warnings, and important information
			if strings.Contains(strings.ToLower(line), "error") ||
				strings.Contains(strings.ToLower(line), "warning") ||
				strings.Contains(strings.ToLower(line), "done") {
				log.Printf("[%s] %s", pipeType, line)
			}

		case "quiet":
			// Only log errors
			if strings.Contains(strings.ToLower(line), "error") {
				log.Printf("[%s] %s", pipeType, line)
			}

		case "debug":
			// Log everything and add a "[DEBUG]" tag to all output
			log.Printf("[DEBUG][%s] %s", pipeType, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from %s: %v", pipeType, err)
	}
}

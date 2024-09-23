package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"spi-go-core/internal/config"
)

type ExecHandler struct {
	Config *config.AppConfig
}

// CommandResponse represents the structure of the response
type CommandResponse struct {
	Output string `json:"output"` // Output of the executed command
	Error  string `json:"error"`  // Error message, if any
}

// Global variable for allowed commands
var allowedCommands = map[string]bool{}

// SetAllowedCommands updates the allowed commands based on config
func SetAllowedCommands(commands []string) {
	for _, cmd := range commands {
		allowedCommands[cmd] = true
	}
}

// RequestPayload represents the structure of the incoming request for exec
type RequestPayload struct {
	Command string `json:"command"`
}

// HandleExecCommand handles the POST request to execute a command
func HandleExecCommand(w http.ResponseWriter, r *http.Request) {
	if config.GlobalConfig == nil {
		http.Error(w, "Unable to read configuration for handler", http.StatusInternalServerError)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON payload to extract the command
	var payload RequestPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Check if encryption is enabled in the config
	var commandStr string
	if config.GlobalConfig.EncryptionConfig.Enabled {
		// Handle encryption if enabled (not shown here for simplicity)
	} else {
		// If encryption is not enabled, use the command as is
		commandStr = payload.Command
	}

	// Extract and validate the command
	cmd := extractCommand(commandStr)
	if !isWhitelistedCommand(cmd) {
		http.Error(w, fmt.Sprintf("Command '%s' is not allowed", cmd), http.StatusForbidden)
		return
	}

	// Execute the system command
	out, err := execCommand(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Command execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the response as JSON and send it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)

	// Send the output back
	//w.Write([]byte(out))
}

// extractCommand extracts the base command from the input (e.g., "ls -la" -> "ls")
func extractCommand(fullCommand string) string {
	var cmd string
	_, err := fmt.Sscanf(fullCommand, "%s", &cmd)
	if err != nil {
		return ""
	}
	return cmd
}

// isWhitelistedCommand checks if the given command is allowed
func isWhitelistedCommand(command string) bool {
	_, exists := allowedCommands[command]
	return exists
}

// execCommand executes the system command and returns the output or error
func execCommand(command string) (string, error) {
	// Execute the command using os/exec
	output, err := exec.Command("sh", "-c", command).Output()

	// Prepare the response
	response := CommandResponse{
		Output: string(output),
	}
	if err != nil {
		response.Error = err.Error()
	}

	return response.Output, err
}

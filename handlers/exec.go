package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RequestPayload represents the structure of the incoming request for exec
type RequestPayload struct {
	Command string `json:"command"`
}

// HandleExecCommand handles the POST request to execute a command
func HandleExecCommand(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON command
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Execute the system command (simplified for this example)
	out, err := execCommand(payload.Command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Command execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the output back
	w.Write([]byte(out))
}

// Simplified execCommand (this can be expanded later)
func execCommand(command string) (string, error) {
	// This is a placeholder. In reality, you'll execute the command and capture its output
	return fmt.Sprintf("Executed: %s", command), nil
}

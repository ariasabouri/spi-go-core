package handlers

import (
	"net/http"
)

// FunctionMap is a global map for registering handler functions by action name
var FunctionMap = map[string]http.HandlerFunc{}

// RegisterHandlers registers the action names to the appropriate handler functions
func RegisterHandlers() {
	// Map action names to their corresponding handler functions
	FunctionMap["HandleExecCommand"] = HandleExecCommand
	FunctionMap["HandleRoot"] = HandleRoot
}

package handlers

import (
	"net/http"
)

// HandleRoot handles the basic GET request at the root
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ServerProfileInstaller (SPI) API is running securely via HTTPS.\n"))
}

package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"spi-go-core/helpers"
	"spi-go-core/internal/encryption"
	"time"
)

type KeyExchangeRequest struct {
	TSAppPublicKey string `json:"tsAppPublicKey"`
}

type KeyExchangeResponse struct {
	GoCorePublicKey string `json:"goCorePublicKey"`
}

type VerificationRequest struct {
	EncryptedResponse string `json:"encryptedResponse"`
}

type VerificationResponse struct {
	EncryptedResponse string `json:"encryptedResponse"`
	OwnChallenge      string `json:"ownChallenge"`
}

type FinalizationRequest struct {
	Secret string `json:"secret"`
}

// HandleKeyExchange handles the initial public key exchange and returns a session ID
func HandleKeyExchange(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for key exchange")
	var req KeyExchangeRequest

	// Parse the request body to get the TypeScript app's public key
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate the session ID now, marking the beginning of the session
	sessionID := encryption.GenerateReqId()

	// Use the new function to parse the public key
	tsAppPublicKey, err := encryption.ParsePublicKey(req.TSAppPublicKey)
	if err != nil {
		helpers.JSONError(w, "Failed to parse TypeScript app public key", http.StatusBadRequest)
		return
	}
	log.Printf("Received and parsed TypeScript app public key.")

	// Create the ConnectionData struct with the public key
	connectionData := encryption.ConnectionData{
		PublicKey: tsAppPublicKey,
		Timestamp: time.Now(),
	}

	// Store the connection data using the existing StoreConnectionData function
	encryption.StoreConnectionData(sessionID, connectionData)
	log.Printf("Stored connection data for session %s", sessionID)

	// Get Go core's public key in PEM format
	goCorePublicKeyPEM, err := encryption.GetGoCorePublicKeyPEM()
	if err != nil {
		helpers.JSONError(w, "Failed to load Go public key", http.StatusInternalServerError)
		return
	}

	// Prepare the response containing Go core's public key and session ID
	response := KeyExchangeResponse{
		GoCorePublicKey: goCorePublicKeyPEM,
	}
	w.Header().Set("X-Request-ID", sessionID)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error sending response:", err)
		return
	}

	log.Printf("Successfully completed key exchange with session ID: %s", sessionID)
}

// HandleMessageVerification handles the decrypted message from TS app and re-encrypts it
func HandleMessageVerification(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from header
	sessionID := r.Header.Get("X-Request-ID")
	if sessionID == "" {
		helpers.JSONError(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	// Retrieve per-session data
	connectionData, err := encryption.GetConnectionData(sessionID)
	if err != nil {
		helpers.JSONError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Load Go core's private key
	goPrivateKey, _, err := encryption.LoadGoKeys()
	if err != nil {
		helpers.JSONError(w, "Failed to load Go private key", http.StatusInternalServerError)
		return
	}

	// Parse and decode the encrypted message from the TS app
	var req VerificationResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	encryptedMessage, err := base64.StdEncoding.DecodeString(req.EncryptedResponse)
	if err != nil {
		helpers.JSONError(w, "Failed to decode message", http.StatusBadRequest)
		return
	}

	// Decrypt the message with Go core's private key
	decryptedMessage, err := encryption.DecryptWithPrivateKey(encryptedMessage, goPrivateKey)
	if err != nil {
		helpers.JSONError(w, "Failed to decrypt message", http.StatusInternalServerError)
		return
	}

	// Re-encrypt the message using the TypeScript app's public key from the session data
	reEncryptedMessage, err := encryption.EncryptWithPublicKey(decryptedMessage, connectionData.PublicKey)
	if err != nil {
		helpers.JSONError(w, "Failed to encrypt message with TypeScript app's public key", http.StatusInternalServerError)
		return
	}

	// Generate and store the challenge secret
	challengeSecret := encryption.GenerateRandomString(64)
	connectionData.ChallengeSecret = challengeSecret
	encryption.StoreConnectionData(sessionID, connectionData)

	// Encrypt the challenge secret with the TypeScript app's public key
	ownChallenge, err := encryption.EncryptWithPublicKey([]byte(challengeSecret), connectionData.PublicKey)
	if err != nil {
		helpers.JSONError(w, "Failed to encrypt own challenge message with TypeScript app's public key", http.StatusInternalServerError)
		return
	}
	log.Printf("Requesting final challenge from client. Secret: %s", challengeSecret)

	// Send the re-encrypted message and own challenge back to the TypeScript app
	w.Header().Set("Content-Type", "application/json")
	response := VerificationResponse{
		EncryptedResponse: base64.StdEncoding.EncodeToString(reEncryptedMessage),
		OwnChallenge:      base64.StdEncoding.EncodeToString(ownChallenge),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error sending response:", err)
		return
	}
}

// HandleSuccess handles the final confirmation from the TypeScript app that the handshake was successful
func HandleSuccess(w http.ResponseWriter, r *http.Request) {
	// Retrieve session ID from header
	sessionID := r.Header.Get("X-Request-ID")
	if sessionID == "" {
		helpers.JSONError(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	// Retrieve per-session data
	connectionData, err := encryption.GetConnectionData(sessionID)
	if err != nil {
		helpers.JSONError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Load Go core's private key
	goPrivateKey, _, err := encryption.LoadGoKeys()
	if err != nil {
		helpers.JSONError(w, "Failed to load Go private key", http.StatusInternalServerError)
		return
	}

	// Parse and decode the secret from the TypeScript app
	var req FinalizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.JSONError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	encryptedSecret, err := base64.StdEncoding.DecodeString(req.Secret)
	if err != nil {
		helpers.JSONError(w, "Failed to decode message", http.StatusBadRequest)
		return
	}

	// Decrypt the secret with Go core's private key
	decryptedSecret, err := encryption.DecryptWithPrivateKey(encryptedSecret, goPrivateKey)
	if err != nil {
		helpers.JSONError(w, "Failed to decrypt message", http.StatusInternalServerError)
		return
	}

	// Check if the decrypted secret matches the stored challenge secret
	if !bytes.Equal(decryptedSecret, []byte(connectionData.ChallengeSecret)) {
		helpers.JSONError(w, "Failed to verify final secret", http.StatusUnauthorized)
		return
	}

	// Mark the session as validated
	connectionData.Validated = true
	encryption.StoreConnectionData(sessionID, connectionData)

	log.Println("Connection successfully validated with the TypeScript app.")

	// Respond to the TypeScript app
	w.Header().Set("Content-Type", "application/json")

	response := struct {
		Msg string `json:"msg"`
	}{
		Msg: "handshake successful!",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error sending response:", err)
		return
	}
}

package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	mathRand "math/rand"
	"os"
	"spi-go-core/helpers"
	"spi-go-core/internal/config"
	"sync"
	"time"
)

type ConnectionData struct {
	PublicKey       *rsa.PublicKey
	Timestamp       time.Time
	ChallengeSecret string
	Validated       bool
}

var (
	TsAppPublicKey      *rsa.PublicKey
	ConnectionValidated = false
	connectionLock      sync.Mutex
	connectionDataMap   = make(map[string]ConnectionData)
	connectionDataMux   = sync.RWMutex{}
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

// GenerateRandomString returns a random string of length n
func GenerateRandomString(n int) string {
	seededRand := mathRand.New(mathRand.NewSource(time.Now().UnixNano())) // Use mathRand alias
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// LoadGoKeys loads Go core private and public keys
func LoadGoKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	log.Printf("Trying to read private key in %s...", config.GlobalConfig.Encryption.PrivateKey)
	pathPrivateKey, _ := helpers.ResolvePath(config.GlobalConfig.Encryption.PrivateKey)
	log.Printf("Resolving path %s...", pathPrivateKey)
	privateKeyFile, err := os.ReadFile(pathPrivateKey)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(privateKeyFile)
	if block == nil {
		log.Println("Failed to decode private key PEM block")
		return nil, nil, errors.New("failed to decode private key PEM block")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Successfully read private key.")

	log.Println("Trying to read public key...")
	publicKeyFile, err := os.ReadFile(config.GlobalConfig.Encryption.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	blockPub, _ := pem.Decode(publicKeyFile)
	publicKey, err := x509.ParsePKIXPublicKey(blockPub.Bytes)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Successfully read public key.")

	return privateKey.(*rsa.PrivateKey), publicKey.(*rsa.PublicKey), nil
}

// ParsePublicKey parses a PEM-encoded public key string and returns an *rsa.PublicKey
func ParsePublicKey(pemKey string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key type, expected RSA")
	}
	return publicKey, nil
}

// EncryptWithPublicKey encrypts data with a given public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key is nil")
	}
	encryptedMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pub, msg)
	if err != nil {
		return nil, err
	}
	return encryptedMsg, nil
}

// DecryptWithPrivateKey decrypts data with a given private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("private key is nil")
	}
	decryptedMsg, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		return nil, err
	}
	return decryptedMsg, nil
}

// GenerateReqId generates a request ID (UUID-like or random string)
func GenerateReqId() string {
	requestIDBytes := make([]byte, 16)
	_, err := rand.Read(requestIDBytes)
	if err != nil {
		panic("Failed to generate request ID")
	}
	requestID := base64.URLEncoding.EncodeToString(requestIDBytes)
	return requestID
}

// IsConnectionValidated checks if a connection is validated for a given request ID
func IsConnectionValidated(requestID string) bool {
	connectionLock.Lock()
	defer connectionLock.Unlock()

	connData, exists := connectionDataMap[requestID]
	if !exists {
		return false
	}

	// Add expiration logic if needed (e.g., 5 minutes)
	if time.Since(connData.Timestamp) > 5*time.Minute {
		delete(connectionDataMap, requestID)
		return false
	}

	return true
}

// StoreConnectionData stores connection data for a session
func StoreConnectionData(sessionID string, data ConnectionData) {
	log.Printf("Storing connection data for Session ID %s...", sessionID)
	connectionDataMux.Lock()
	defer connectionDataMux.Unlock()
	connectionDataMap[sessionID] = data
}

// StoreTsAppPublicKey stores the TypeScript app's public key
func StoreTsAppPublicKey(pemKey string) error {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return errors.New("failed to decode PEM block containing public key")
	}
	tsAppPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	TsAppPublicKey = tsAppPublicKey.(*rsa.PublicKey)
	return nil
}

// GetGoCorePublicKeyPEM returns Go core's public key in PEM format
func GetGoCorePublicKeyPEM() (string, error) {
	_, goPublicKey, err := LoadGoKeys()
	if err != nil {
		return "", err
	}
	goPublicKeyBytes, err := x509.MarshalPKIXPublicKey(goPublicKey)
	if err != nil {
		return "", err
	}
	goCorePublicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: goPublicKeyBytes,
	})
	return string(goCorePublicKeyPEM), nil
}

func GetConnectionData(sessionID string) (ConnectionData, error) {
	connectionDataMux.RLock()
	defer connectionDataMux.RUnlock()
	data, exists := connectionDataMap[sessionID]
	if !exists {
		return ConnectionData{}, errors.New("session not found")
	}
	return data, nil
}

package encryption

import (
	"bytes"
	"path/filepath"
	"spi-go-core/internal/config"
	"testing"
)

func TestEncryptionDecryption(t *testing.T) {
	// Load the configuration as before
	configPath, err := filepath.Abs("../../config.json")
	if err != nil {
		t.Fatalf("Failed to get absolute path of config.json: %v", err)
	}
	_, err = config.LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Adjust the paths in the configuration to be correct relative to the test directory
	config.GlobalConfig.Encryption.PrivateKey = filepath.Join("../../", config.GlobalConfig.Encryption.PrivateKey)
	config.GlobalConfig.Encryption.PublicKey = filepath.Join("../../", config.GlobalConfig.Encryption.PublicKey)

	// Load the Go core private and public keys as before
	privateKey, publicKey, err := LoadGoKeys()
	if err != nil {
		t.Fatalf("Failed to load Go keys: %v", err)
	}

	// Test message
	originalMessage := []byte("This is a test message.")
	t.Logf("Test message: %s", originalMessage)

	// Encrypt the message using the public key
	encryptedMessage, err := EncryptWithPublicKey(originalMessage, publicKey)
	if err != nil {
		t.Fatalf("Failed to encrypt message: %v", err)
	}
	t.Logf("Encrypted message: %s", encryptedMessage)

	// Decrypt the message using the private key
	decryptedMessage, err := DecryptWithPrivateKey(encryptedMessage, privateKey)
	if err != nil {
		t.Fatalf("Failed to decrypt message: %v", err)
	}
	t.Logf("Decrypted message: %s", decryptedMessage)

	// Compare the decrypted message with the original
	if !bytes.Equal(originalMessage, decryptedMessage) {
		t.Errorf("Decrypted message does not match original.\nOriginal: %s\nDecrypted: %s", originalMessage, decryptedMessage)
	} else {
		t.Logf("Encryption and decryption successful. Message: %s", decryptedMessage)
	}
}

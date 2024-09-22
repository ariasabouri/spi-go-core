# Go Core API Security Specifications

To address concerns around **privilege escalation** and **secure communication**, we need to implement several security measures. Here's how we can ensure the **Go Core API** is secure.

## 1. Preventing Privilege Escalation

### Access Control
- The Go Core will only accept API requests from **authenticated and authorized** processes.
- Restrict access by requiring the TypeScript application to authenticate using tokens, certificates, or pre-shared keys.
- Limit which user/group can start the API and restrict API access to specific users/services (e.g., run Go Core as a specific user or restrict network access via firewalls).

### Input Validation and Command Whitelisting
- Validate all input to avoid malicious commands.
- Implement a **whitelist** of allowed commands (e.g., mounting, package installations) and reject unsafe commands (`rm -rf`, `sudo`).
- Sanitize input using regular expressions.

### Limit Privileges
- Run Go Core with **minimal privileges** (non-root user) and restrict access to only necessary resources (file system, network).
- Only elevate privileges when absolutely required (e.g., for mounting disks).

### User-Based Restrictions
- Run Go Core under a **specific service account**.
- Use UNIX domain sockets to restrict API access to certain users.

## 2. Secure Communication (HTTPS)

### HTTPS
- All API communication should happen over **HTTPS** to protect against **MITM attacks**.
- The Go Core can generate or use a self-signed SSL certificate, or the user can provide one.
- The API will enforce HTTPS and reject HTTP requests.

### Mutual TLS (mTLS)
- Implement **mutual TLS (mTLS)** to ensure both the client (TypeScript app) and server (Go Core) authenticate each other.
- The TypeScript app will authenticate using its certificate, and the Go Core will verify it.

## 3. Public-Key Encryption for Command Payloads

### Public-Private Key Pair
- The Go Core will generate a **public-private key pair**. The public key will be used by the TypeScript app to encrypt command payloads.
- Only the Go Core can decrypt the commands with its private key, protecting against eavesdropping.

### Command Encryption Flow
1. Go Core generates a key pair.
2. The TypeScript app retrieves the public key during handshake.
3. The TypeScript app encrypts commands using the public key.
4. The Go Core decrypts commands and validates them before execution.

### Signature Verification
- The TypeScript app will **sign command payloads** using its private key, and the Go Core will verify the signature to ensure integrity.

## 4. Implementation Details

### a. HTTPS Setup with Go

Use the following Go code to set up a secure HTTPS server:

```go
package main

import (
    "crypto/tls"
    "log"
    "net/http"
)

func setupSecureServer() {
    cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        log.Fatal(err)
    }

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
    }

    server := &http.Server{
        Addr:      ":8443",
        TLSConfig: tlsConfig,
        Handler:   nil,
    }

    log.Println("Starting secure server on port 8443")
    log.Fatal(server.ListenAndServeTLS("", ""))
}

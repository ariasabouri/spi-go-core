import * as https from 'https';
import * as crypto from 'crypto';
import * as fs from 'fs';

class SPIClient {
    private publicKey: string;
    private privateKey: string;
    private hostname: string;
    private port: number;
    private rejectUnauthorized: boolean;

    constructor(
        publicKeyPath: string,
        privateKeyPath: string,
        hostname: string = 'localhost',
        port: number = 8443,
        rejectUnauthorized: boolean = false
    ) {
        this.publicKey = fs.readFileSync(publicKeyPath, 'utf8');
        this.privateKey = fs.readFileSync(privateKeyPath, 'utf8');
        this.hostname = hostname;
        this.port = port;
        this.rejectUnauthorized = rejectUnauthorized;
    }

    // Encrypt a command using the public key
    public encryptCommand(command: string): Buffer {
        const buffer = Buffer.from(command, 'utf8');
        const encrypted = crypto.publicEncrypt(this.publicKey, buffer);
        return encrypted;
    }

    // Decrypt a response from the server using the private key
    public decryptResponse(response: Buffer): string {
        const decrypted = crypto.privateDecrypt(this.privateKey, response);
        return decrypted.toString('utf8');
    }

    // Send an encrypted command to the Go server
    public sendCommand(command: string): Promise<string> {
        return new Promise((resolve, reject) => {
            const encryptedCommand = this.encryptCommand(command);

            const options = {
                hostname: this.hostname,
                port: this.port,
                path: '/api/exec',
                method: 'POST',
                headers: {
                    'Content-Type': 'application/octet-stream',
                    'Content-Length': encryptedCommand.length,
                },
                rejectUnauthorized: this.rejectUnauthorized, // For self-signed certificates
            };

            const req = https.request(options, (res) => {
                let data = '';

                res.on('data', (chunk) => {
                    data += chunk;
                });

                res.on('end', () => {
                    resolve(data);
                });
            });

            req.on('error', (error) => {
                reject(`Request error: ${error.message}`);
            });

            // Write the encrypted command to the request body
            req.write(encryptedCommand);
            req.end();
        });
    }
}

// Usage example
const client = new SPIClient(
    'path/to/public_key.pem',
    'path/to/private_key.pem',
    'localhost',
    8443
);

client.sendCommand('ls -la')
    .then(response => {
        console.log('Server Response:', response);
    })
    .catch(error => {
        console.error('Error:', error);
    });

import * as crypto from 'crypto';
import * as fs from 'fs';

export class RSAUtility {
    private publicKey: string;
    private privateKey: string;

    constructor(publicKeyPath: string, privateKeyPath: string) {
        this.publicKey = fs.readFileSync(publicKeyPath, 'utf8');
        this.privateKey = fs.readFileSync(privateKeyPath, 'utf8');
    }

    // Encrypt a message using the public key
    public encryptMessage(message: string): Buffer {
        const buffer = Buffer.from(message, 'utf8');
        const encrypted = crypto.publicEncrypt(this.publicKey, buffer);
        return encrypted;
    }

    // Decrypt a message using the private key
    public decryptMessage(encryptedMessage: Buffer): string {
        const decrypted = crypto.privateDecrypt(this.privateKey, encryptedMessage);
        return decrypted.toString('utf8');
    }
}

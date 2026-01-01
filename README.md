# go-send

[![CI](https://github.com/VinMeld/go-send/actions/workflows/ci.yml/badge.svg)](https://github.com/VinMeld/go-send/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/VinMeld/go-send)](https://goreportcard.com/report/github.com/VinMeld/go-send)

A simple, secure file-sharing application written in Go. It uses end-to-end encryption (E2EE) to ensure that only the intended recipient can decrypt and read the files.

## Features
- **End-to-End Encryption**: Files are encrypted on the client side using X25519 and XSalsa20-Poly1305.
- **Ephemeral Keys**: A new symmetric key is generated for every file transfer.
- **Auto-Delete**: Optional flag to delete files from the server immediately after download.
- **S3 Support**: Can use AWS S3 for file storage.
- **Structured Logging**: Server uses `log/slog` for machine-readable logs.
- **CI/CD**: Automated testing and linting via GitHub Actions.
- **Store-and-Forward**: Send files to users even when they are offline. The server stores the encrypted blob.
- **Key Management**: Simple CLI for generating identity keys and managing a local address book of public keys.
- **Client-Server Architecture**:
  - **Server**: HTTP backend for storing encrypted blobs and user metadata.
  - **Client**: CLI tool for encryption, decryption, and management.

## Architecture

### Crypto
- **Identity Keys**: Each user has a long-term Ed25519/X25519 keypair.
- **File Encryption**:
  1. A random ephemeral keypair is generated for each file transfer.
  2. The file content is encrypted using the Ephemeral Private Key and the Recipient's Public Key.
  3. The Ephemeral Public Key is attached to the file metadata.
  4. The recipient decrypts using their Private Key and the attached Ephemeral Public Key.

### Directory Structure
- `cmd/client`: Main entry point for the CLI application.
- `cmd/server`: Main entry point for the HTTP server.
- `internal/client`: Client-specific logic (Config, Commands).
- `internal/server`: Server-specific logic (Storage, Handlers).
- `internal/crypto`: Shared cryptographic utilities.
- `internal/models`: Shared data structures.

## Installation

```bash
# Clone the repository
git clone https://github.com/VinMeld/go-send.git
cd go-send

# Initialize dependencies
go mod tidy
```

## Usage

### 1. Start the Server
The server stores the encrypted files. Run this in a separate terminal or on a remote machine.

```bash
go run cmd/server/main.go -port :9090
# Server listening on :9090
```

### Configuration (.env)
The server supports configuration via a `.env` file or environment variables.

**Example `.env`:**
```env
PORT=:9090
STORAGE_TYPE=s3 # Options: local (default), s3
AWS_BUCKET=my-bucket
AWS_REGION=us-east-1
# Optional: Restrict registration
REGISTRATION_TOKEN=secret123
```

To use S3 storage:
1. Set `STORAGE_TYPE=s3`
2. Set `AWS_BUCKET` and `AWS_REGION`
3. Ensure AWS credentials are set (e.g., `~/.aws/credentials` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` env vars).

### 2. Client Setup (Alice & Bob)

**Initialize Alice:**
```bash
go run cmd/client/main.go config init --user alice --config alice.json
# Output: Public Key: <ALICE_PUB_KEY>
```

**Configure Server (Optional):**
If the server is not on localhost:8080, set the URL:
```bash
go run cmd/client/main.go set-server http://localhost:9090 --config alice.json
```

**Check Connection:**
```bash
go run cmd/client/main.go ping --config alice.json
# Output: Pong! Server is reachable
```

**Initialize Bob:**
```bash
go run cmd/client/main.go config init --user bob --config bob.json
# Output: Public Key: <BOB_PUB_KEY>
```

**Register with Server (If Token Required):**
```bash
go run cmd/client/main.go register --token secret123 --config alice.json
```

### 3. Exchange Keys
Alice needs Bob's public key to send him a file.

```bash
# Alice adds Bob
go run cmd/client/main.go add-user bob <BOB_PUB_KEY> --config alice.json
```

### 4. Send a File
Alice sends a file to Bob.

```bash
```bash
echo "Top Secret" > secret.txt
go run cmd/client/main.go send-file bob secret.txt --config alice.json

# Send with Auto-Delete (File removed from server after download)
go run cmd/client/main.go send-file bob secret.txt --auto-delete --config alice.json
```

### 5. Receive a File
Bob checks for files and downloads them.

```bash
# List files
go run cmd/client/main.go list-files --config bob.json

# Download and Decrypt
go run cmd/client/main.go download-file <FILE_ID> --config bob.json
```

### 6. Delete a File
Both the sender and recipient can delete a file from the server.

```bash
go run cmd/client/main.go delete-file <FILE_ID> --config alice.json
```

The decrypted file will be saved with its original filename.

## Code Overview

- **`internal/crypto/crypto.go`**: Wrappers around `golang.org/x/crypto/nacl/box` for easy encryption/decryption.
- **`internal/server/storage.go`**: Simple JSON-based file persistence for the server (MVP).
- **`internal/client/send_cmd.go`**: Logic for generating ephemeral keys, encrypting files, and uploading.
- **`internal/client/download_cmd.go`**: Logic for downloading and decrypting using the recipient's private key.
- **`internal/server/handler.go`**: HTTP handlers for file and user management.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

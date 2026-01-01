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

## Installation

### Client

**Binary:**
Download the latest binary for your platform from the [Releases](https://github.com/VinMeld/go-send/releases) page.

**Homebrew (macOS/Linux):**
```bash
brew install vinmeld/tap/go-send
```

**AUR (Arch Linux):**
```bash
yay -S go-send-bin
```

### Server

**Binary:**
Download the `go-send-server` binary from the [Releases](https://github.com/VinMeld/go-send/releases) page.

**Docker:**
```bash
docker pull meldrum123454/go-send
```

### Server Configuration

The server supports configuration via environment variables or a `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Port to listen on | `:8080` |
| `STORAGE_TYPE` | Storage backend (`local` or `s3`) | `local` |
| `AWS_BUCKET` | AWS S3 Bucket name (if `STORAGE_TYPE=s3`) | - |
| `AWS_REGION` | AWS Region (if `STORAGE_TYPE=s3`) | - |
| `REGISTRATION_TOKEN` | Secret token required for user registration | - |

## Commands

```text
go-send
Secure file sending CLI

Usage:
  go-send [command]

Available Commands:
  add-user      Add a known user
  completion    Generate the autocompletion script for the specified shell
  config        Manage configuration
  delete-file   Delete a file from the server
  download-file Download and decrypt a file
  help          Help about any command
  list-files    List files waiting for the current user
  list-users    List known users (local and server)
  login         Authenticate with the server
  ping          Check connection to the server
  register      Register the current user with the server
  remove-user   Remove a known user
  send-file     Send an encrypted file
  set-server    Set the remote server URL
  set-user      Set current active user

Flags:
      --config string   config file (default is $HOME/.config/go-send/config.json)
  -h, --help            help for go-send
```

## Usage Scenario: Alice & Bob

### 1. Start the Server

**Using Binary:**
```bash
./go-send-server -port :9090
```

**Using Docker:**
```bash
docker run -p 9090:8080 meldrum123454/go-send
```

### 2. Client Setup

**Initialize Alice:**
```bash
go-send config init --user alice --server http://localhost:9090 --config alice.json
# Output: Public Key: <ALICE_PUB_KEY>
```

**Initialize Bob:**
```bash
go-send config init --user bob --server http://localhost:9090 --config bob.json
# Output: Public Key: <BOB_PUB_KEY>
```

**Check Connection:**
```bash
go-send ping --config alice.json
# Output: Pong! Server is reachable
```

**Register with Server (If Token Required):**
```bash
go-send register --token secret123 --config alice.json
go-send register --token secret123 --config bob.json
```

**Login:**
```bash
go-send login --config alice.json
go-send login --config bob.json
```

### 3. User Discovery & Listing
You can list users known to the server. This is helpful to find usernames.

```bash
go-send list-users --config alice.json
```

### 4. Send a File
Alice sends a file to Bob. If Bob is not in Alice's local address book, the client will automatically fetch Bob's keys from the server (User Discovery).

```bash
echo "Top Secret" > secret.txt
go-send send-file bob secret.txt --config alice.json

# Send with Auto-Delete (File removed from server after download)
go-send send-file bob secret.txt --auto-delete --config alice.json
```

### 5. Receive a File
Bob lists his files and downloads them.

```bash
# List files (shows Index and ID)
go-send list-files --config bob.json
# Output:
# 1 - [FILE_ID] secret.txt (from alice) - <TIMESTAMP>

# Download and Decrypt using Index
go-send download-file 1 --config bob.json
# Or using ID
go-send download-file <FILE_ID> --config bob.json
```

### 6. Delete a File
Both the sender and recipient can delete a file from the server.

```bash
go-send delete-file <FILE_ID> --config alice.json
```

## Testing

The project includes comprehensive testing:

- **Unit Tests**: Cover individual components and logic.
- **Integration Tests**: Verify the interaction between the client and server.

To run tests:
```bash
go test ./...
```

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

## Code Overview

- **`internal/crypto/crypto.go`**: Wrappers around `golang.org/x/crypto/nacl/box` for easy encryption/decryption.
- **`internal/server/storage.go`**: Simple JSON-based file persistence for the server (MVP).
- **`internal/client/send_cmd.go`**: Logic for generating ephemeral keys, encrypting files, and uploading.
- **`internal/client/download_cmd.go`**: Logic for downloading and decrypting using the recipient's private key.
- **`internal/server/handler.go`**: HTTP handlers for file and user management.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

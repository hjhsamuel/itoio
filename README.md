# itoio

`itoio` is a high-performance, room-based real-time communication platform built with Go and WebRTC. It features screen sharing, instant messaging, and a secure invite-only registration system.

## 🚀 Features

- **P2P Communication**: Real-time media streaming using WebRTC and signaling via WebSockets.
- **Screen Sharing**: Easily share your screen with other participants in a room.
- **Room Management**: Flexible room-based organization for group communication.
- **Invite-only Registration**: Secure onboarding process via invite codes (max 5 per user).
- **Embedded STUN/TURN**: Built-in STUN/TURN server support for reliable NAT traversal.
- **JWT Authentication**: Secure user sessions with JWT tokens and Argon2id password hashing.
- **Cross-platform**: Written in Go, it can be easily deployed across different operating systems.

## 🛠️ Prerequisites

- **Go**: 1.25 or higher
- **Node.js**: 18.x or higher (for building the UI)
- **npm**: for dependency management

## 📦 Installation

### 1. Build the Frontend

Navigate to the `ui` directory and build the web interface:

```bash
cd ui
npm install
npm run build
```

### 2. Build the Backend

From the project root, build the Go executable:

```bash
go build -o ito cmd/*.go
```

## ⚙️ Configuration

`itoio` can be configured using a YAML file or environment variables.

### YAML Configuration

Create a `config.yaml` file (default name) in the same directory as the executable:

```yaml
server:
  port: 5001
  storage_path: "ito.db" # Use ":memory:" for in-memory storage
  # tls_cert_file: "path/to/cert.pem"
  # tls_key_file: "path/to/key.pem"

turn:
  port: 15432
  public_ip: "your.public.ip"
  realm: "ito-webrtc"
  mode: "stun" # "stun" or "turn"

log:
  level: "info"
  path: "log/ito.log"
```

### Environment Variables

All configuration options can be overridden using environment variables with the `ITOIO` prefix:

- `ITOIO_SERVER|PORT`
- `ITOIO_TURN|PUBLIC_IP`
- `ITOIO_LOG|LEVEL`
- (Use `|` as a separator for nested keys)

## 🏃 Running the Server

Start the server using the `server` command:

```bash
./ito server -c config.yaml
```

On the first run, `itoio` will initialize an administrator account and print the credentials to the console:
```text
Initialized first user: admin/<random-password>
```

## 🏗️ Project Structure

- `cmd/`: CLI entry points and command definitions.
- `config/`: Configuration parsing and global settings.
- `internal/`: Core business logic, data access, and server handlers.
  - `dao/`: Data Access Object layer using BuntDB.
  - `server/`: WebSocket and TURN server implementations.
- `pkg/`: Utility packages (snowflake, token, password).
- `ui/`: Frontend source code and build scripts.

## 📄 License

This project is licensed under the [LICENSE](LICENSE) file.
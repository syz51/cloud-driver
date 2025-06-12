# Cloud Driver Server

🚀 A simple RESTful API server for 115cloud storage operations. Credentials are provided directly in API requests - no user authentication or database required!

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/cloud-driver)](https://goreportcard.com/report/github.com/yourusername/cloud-driver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24.4+-blue.svg)](https://golang.org/)

## Features

### Core Features

- **🚀 Stateless Operation**: No database or user management required
- **🔑 Credential-based Requests**: 115cloud credentials passed directly in API requests
- **⚡ Lightweight**: Minimal dependencies, fast startup
- **🏥 Health Monitoring**: Built-in health check endpoints
- **🐳 Docker Ready**: Easy deployment with simple configuration

### 115Cloud Integration

- ✅ User authentication and information retrieval
- ✅ File and directory listing with navigation
- ✅ Offline download task management (add, list, delete, clear)
- ✅ File operations (info, download links)
- 🔄 Upload operations (planned)
- 🔄 Advanced file management (move, copy, delete) (planned)

## Architecture

The project follows Go best practices with a clean, layered architecture:

```
cloud-driver/
├── cmd/
│   └── cloud-driver/          # Application entry point
│       └── main.go
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── server/               # HTTP server and routing
│   ├── handlers/             # HTTP request handlers
│   │   ├── drive115.go      # 115cloud API endpoints
│   │   └── health.go        # Health check endpoints
│   ├── services/            # Business logic layer
│   │   └── drive115.go      # 115cloud integration service
│   └── models/              # Data models and request/response structures
├── config.yml                # Configuration file
├── config.yaml.example       # Example configuration
├── .air.conf                 # Air hot reload configuration
├── go.mod                    # Go module definition
└── README.md                 # This file
```

### Technology Stack

- **Framework**: [Echo v4](https://echo.labstack.com/) - High performance HTTP router
- **Configuration**: [Viper](https://github.com/spf13/viper) for flexible config management
- **115Cloud Client**: [115driver](https://github.com/SheltonZhu/115driver) for 115cloud API integration
- **Development**: [Air](https://github.com/air-verse/air) for hot reloading

## Quick Start

### Prerequisites

- **Go 1.24.4+**
- **Docker** (optional, for easy deployment)

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd cloud-driver

# Copy and configure the settings
cp config.yaml.example config.yml
# Edit config.yml with your settings (mainly just server port)

# Build and run with Docker
docker build -t cloud-driver .
docker run -p 8080:8080 -v $(pwd)/config.yml:/app/config.yml cloud-driver
```

### Option 2: Direct Run

```bash
# Clone the repository
git clone <repository-url>
cd cloud-driver

# Install dependencies
go mod download

# Copy and configure
cp config.yaml.example config.yml
# Edit config.yml with your server settings

# Run the application
go run cmd/cloud-driver/main.go

# Or with hot reload for development
go install github.com/air-verse/air@latest
air
```

## Configuration

### Configuration File (`config.yml`)

```yaml
server:
  host: "0.0.0.0" # Server bind address
  port: 8080 # Server port
```

### Environment Variables

Override configuration using environment variables with the `CLOUD_DRIVER_` prefix:

```bash
export CLOUD_DRIVER_SERVER_PORT=8080
export CLOUD_DRIVER_SERVER_HOST=0.0.0.0
```

### Getting 115Cloud Credentials

1. Log in to your 115cloud account in a web browser
2. Open browser developer tools (F12)
3. Go to the Network tab
4. Perform any action on the 115cloud website
5. Find a request to 115cloud API and extract the following from cookies:
   - `UID`: User ID
   - `CID`: Session ID
   - `SEID`: Secure session ID
   - `KID`: Key ID

## API Documentation

All API endpoints require 115cloud credentials to be included in the request body:

```json
{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  }
  // ... other request parameters
}
```

### Health Check

```bash
GET /health
```

### Get User Information

```bash
POST /api/v1/115/user
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  }
}
```

### List Files

```bash
POST /api/v1/115/files
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  },
  "dir_id": 0  // 0 for root directory, optional
}
```

### List Offline Tasks

```bash
POST /api/v1/115/tasks
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  },
  "page": 1  // Optional, defaults to 1
}
```

### Add Offline Download Task

```bash
POST /api/v1/115/tasks/add
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  },
  "urls": [
    "http://example.com/file1.zip",
    "magnet:?xt=urn:btih:..."
  ],
  "save_dir_id": "0"  // Optional, defaults to root directory
}
```

### Delete Offline Tasks

```bash
POST /api/v1/115/tasks/delete
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  },
  "hashes": ["hash1", "hash2"],
  "delete_files": false  // Whether to delete files as well
}
```

### Clear Offline Tasks

```bash
POST /api/v1/115/tasks/clear
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  },
  "clear_flag": 1  // Clear flag
}
```

### Get File Information

```bash
POST /api/v1/115/files/:id
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  }
}
```

### Get Download Information

```bash
POST /api/v1/115/files/:id/download
Content-Type: application/json

{
  "credentials": {
    "uid": "your_uid",
    "cid": "your_cid",
    "seid": "your_seid",
    "kid": "your_kid"
  }
}
```

## Development

### Hot Reload with Air

For development, use Air for automatic reloading:

```bash
# Install air
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

### Project Structure

- `cmd/cloud-driver/main.go` - Application entry point
- `internal/config/` - Configuration management
- `internal/server/` - HTTP server setup and routing
- `internal/handlers/` - HTTP request handlers
- `internal/services/` - Business logic and 115cloud integration
- `internal/models/` - Data structures for requests and responses

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/yourusername/cloud-driver/issues) page
2. Create a new issue if your problem isn't already reported
3. Provide detailed information about your environment and the issue

## Acknowledgments

- [115driver](https://github.com/SheltonZhu/115driver) - Excellent 115cloud Go client library
- [Echo](https://echo.labstack.com/) - High performance HTTP framework
- All the contributors who helped make this project better

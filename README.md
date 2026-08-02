# Cloud Driver Server

🚀 A simple RESTful API server for 115cloud storage operations. Credentials are provided directly in API requests - no user authentication or database required!

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/cloud-driver)](https://goreportcard.com/report/github.com/yourusername/cloud-driver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.26.5+-blue.svg)](https://go.dev/)

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
- ✅ Local file uploads with rapid/OSS transfer
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

- **Go 1.26.5+**
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
upload_session_secret: "replace-with-a-random-secret-at-least-32-characters"
upload_part_body_limit: "17M" # 16 MiB parts plus request guard
allowed_origins:
  - "https://drive.example.com"
  - "http://localhost:3012"
```

### Environment Variables

Override configuration using environment variables with the `CLOUD_DRIVER_` prefix:

```bash
export CLOUD_DRIVER_SERVER_PORT=8080
export CLOUD_DRIVER_SERVER_HOST=0.0.0.0
export CLOUD_DRIVER_UPLOAD_SESSION_SECRET='replace-with-a-random-secret-at-least-32-characters'
export CLOUD_DRIVER_UPLOAD_PART_BODY_LIMIT=17M
export CLOUD_DRIVER_ALLOWED_ORIGINS='https://drive.example.com,http://localhost:3012'
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

### Check Folder Videos

Checks direct files only; directories are excluded. `has_videos` is true only
when a direct video filename contains `indexed_name`, case-insensitively. The
response also includes the first page of direct files so folder review can
render file contents without a second 115 request.

```bash
POST /api/v1/115/files/video-check?dir_id=0&limit=25&indexed_name=mukd-569
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

### Upload a Local File

Large files use resumable 16 MiB requests. Browser computes SHA1 first so 115
can attempt rapid upload before any file bytes move. Normal uploads stream each
part directly from request body into 115 OSS; Cloud Run never stages full file
in heap or writable filesystem.

1. `POST /api/v1/115/uploads/init` with credentials, destination, name, size,
   full-file SHA1, and first-128-KiB SHA1.
2. If response is `sign_check`, hash requested inclusive byte range and repeat
   init with `sign_key` and `sign_value`.
3. If response is `upload`, keep returned encrypted bearer token. Use
   `POST /api/v1/115/uploads/status` to resume, then send raw parts to
   `PUT /api/v1/115/uploads/part`
   with `Authorization: Bearer ...` and `X-Part-Number`.
4. Call `POST /api/v1/115/uploads/complete`. Call
   `POST /api/v1/115/uploads/abort` when discarding.

Init request:

```bash
curl -X POST http://localhost:8080/api/v1/115/uploads/init \
  -H 'Content-Type: application/json' \
  -d '{
    "credentials":{"uid":"...","cid":"...","seid":"...","kid":"..."},
    "dir_id":"0","file_name":"video.mp4","file_size":729897389,
    "sha1":"FULL_FILE_SHA1","pre_sha1":"FIRST_128_KIB_SHA1"
  }'
```

Sessions expire after 24 hours and remain decryptable for seven more days only
so abandoned multipart uploads can be aborted. OSS credentials stay server-side
and refresh independently for every part.

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

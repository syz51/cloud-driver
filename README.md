# Cloud Driver Server

🚀 A multi-user RESTful API server for cloud storage operations with built-in user authentication and credential management, starting with 115cloud drive support.

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/cloud-driver)](https://goreportcard.com/report/github.com/yourusername/cloud-driver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24.4+-blue.svg)](https://golang.org/)

## Features

### Core Features

- **🔐 User Authentication**: Secure user registration, login, and session management
- **👥 Multi-User Support**: Each user manages their own cloud storage credentials
- **📦 Database Integration**: PostgreSQL database with automated migrations
- **🔧 Credential Management**: Secure storage and management of 115cloud credentials
- **🐳 Docker Support**: Easy deployment with Docker Compose
- **⚡ Hot Reload**: Development mode with automatic rebuilds using Air
- **🏥 Health Monitoring**: Built-in health check endpoints

### 115Cloud Integration

- ✅ User authentication and information retrieval
- ✅ File and directory listing with navigation
- ✅ Offline download task management (add, list, delete, clear)
- ✅ File operations (info, download links)
- ✅ Multi-credential support per user
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
│   ├── middleware/           # HTTP middleware (auth, etc.)
│   ├── handlers/             # HTTP request handlers
│   │   ├── auth.go          # User authentication endpoints
│   │   ├── credentials.go   # Credential management endpoints
│   │   ├── drive115.go      # 115cloud API endpoints
│   │   └── health.go        # Health check endpoints
│   ├── services/            # Business logic layer
│   │   ├── auth.go          # Authentication service
│   │   ├── credentials.go   # Credential management service
│   │   └── drive115.go      # 115cloud integration service
│   ├── models/              # Data models and request/response structures
│   ├── db/                  # Generated database queries (sqlc)
│   └── database/            # Database migrations and queries
├── database/                  # Database schema and queries
│   ├── schema.sql           # Database schema
│   ├── queries.sql          # SQL queries for sqlc
│   └── migrate.sql          # Initial migration
├── config.yml                # Configuration file
├── config.yaml.example       # Example configuration
├── compose.yml               # Docker Compose setup
├── sqlc.yaml                 # SQLC configuration
├── .air.conf                 # Air hot reload configuration
├── go.mod                    # Go module definition
└── README.md                 # This file
```

### Technology Stack

- **Framework**: [Echo v4](https://echo.labstack.com/) - High performance HTTP router
- **Database**: PostgreSQL 16 with [pgx/v5](https://github.com/jackc/pgx) driver
- **Query Builder**: [SQLC](https://sqlc.dev/) for type-safe database queries
- **Configuration**: [Viper](https://github.com/spf13/viper) for flexible config management
- **115Cloud Client**: [115driver](https://github.com/SheltonZhu/115driver) for 115cloud API integration
- **Development**: [Air](https://github.com/air-verse/air) for hot reloading

## Quick Start

### Prerequisites

- **Go 1.24.4+**
- **PostgreSQL 16+** (or use Docker Compose)
- **Docker & Docker Compose** (optional, for easy setup)

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd cloud-driver

# Start the database
docker-compose up -d postgres

# Copy and configure the settings
cp config.yaml.example config.yml
# Edit config.yml with your settings

# Run the application in development mode
go install github.com/air-verse/air@latest
air

# Or build and run directly
go run cmd/cloud-driver/main.go
```

### Option 2: Manual Setup

```bash
# Clone the repository
git clone <repository-url>
cd cloud-driver

# Install dependencies
go mod download

# Set up PostgreSQL database
createdb cloud_driver

# Run database migrations
psql -d cloud_driver -f database/schema.sql

# Copy and configure
cp config.yaml.example config.yml
# Edit config.yml with your database settings

# Generate database queries (optional, already committed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate

# Run the application
go run cmd/cloud-driver/main.go
```

## Configuration

### Configuration File (`config.yml`)

```yaml
server:
  host: "0.0.0.0" # Server bind address
  port: 8081 # Server port

database:
  host: "127.0.0.1" # PostgreSQL host
  port: 5432 # PostgreSQL port
  user: "postgres" # Database user
  password: "your_password"
  name: "cloud_driver"
  ssl_mode: "disable"
```

### Environment Variables

Override configuration using environment variables with the `CLOUD_DRIVER_` prefix:

```bash
export CLOUD_DRIVER_SERVER_PORT=8081
export CLOUD_DRIVER_DATABASE_HOST=localhost
export CLOUD_DRIVER_DATABASE_PASSWORD=your_password
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

### Authentication Endpoints

#### Register User

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "your_username",
  "email": "your_email@example.com",
  "password": "your_password"
}
```

#### Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

#### Get Profile

```bash
GET /api/v1/auth/profile
Authorization: Bearer <session_token>
```

#### Logout

```bash
POST /api/v1/auth/logout
Authorization: Bearer <session_token>
```

### Credential Management

#### Add 115Cloud Credentials

```bash
POST /api/v1/credentials
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "name": "My 115 Account",
  "uid": "your_115_uid",
  "cid": "your_115_cid",
  "seid": "your_115_seid",
  "kid": "your_115_kid"
}
```

#### List Credentials

```bash
GET /api/v1/credentials
Authorization: Bearer <session_token>
```

#### Update Credentials

```bash
PUT /api/v1/credentials/{id}
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "name": "Updated Account Name",
  "uid": "updated_uid",
  "cid": "updated_cid",
  "seid": "updated_seid",
  "kid": "updated_kid"
}
```

#### Set Credentials Active/Inactive

```bash
PATCH /api/v1/credentials/{id}/status
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "is_active": true
}
```

#### Delete Credentials

```bash
DELETE /api/v1/credentials/{id}
Authorization: Bearer <session_token>
```

### 115Cloud Operations

> **Note**: All 115cloud operations require authentication and active credentials

#### Get User Information

```bash
GET /api/v1/115/user
Authorization: Bearer <session_token>
```

#### List Files and Directories

```bash
GET /api/v1/115/files?dir_id=0
Authorization: Bearer <session_token>
```

#### Get File Information

```bash
GET /api/v1/115/files/{file_id}
Authorization: Bearer <session_token>
```

#### Get Download Information

```bash
POST /api/v1/115/files/{file_id}/download
Authorization: Bearer <session_token>
```

#### List Offline Download Tasks

```bash
GET /api/v1/115/tasks?page=1
Authorization: Bearer <session_token>
```

#### Add Offline Download Task

```bash
POST /api/v1/115/tasks/add
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "urls": ["magnet:?xt=urn:btih:...", "http://example.com/file.zip"],
  "save_dir_id": "0"
}
```

#### Delete Offline Tasks

```bash
DELETE /api/v1/115/tasks
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "hashes": ["task_hash_1", "task_hash_2"],
  "delete_files": false
}
```

#### Clear Offline Tasks

```bash
DELETE /api/v1/115/tasks/clear
Authorization: Bearer <session_token>
Content-Type: application/json

{
  "clear_flag": 1
}
```

### Health Check

```bash
GET /health
```

## Development

### Development Mode with Hot Reload

```bash
# Install Air for hot reloading
go install github.com/air-verse/air@latest

# Start development server
air
```

### Database Operations

```bash
# Generate database queries after modifying database/queries.sql
sqlc generate

# Apply database migrations
psql -d cloud_driver -f database/migrate.sql
```

### Common Go Commands

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build the application
go build -o bin/cloud-driver cmd/cloud-driver/main.go

# Tidy dependencies
go mod tidy

# Run with race detection
go run -race cmd/cloud-driver/main.go
```

## Project Structure Details

### Database Schema

The application uses PostgreSQL with the following main tables:

- **`users`**: User accounts with authentication
- **`user_sessions`**: Session management for authentication
- **`drive115_credentials`**: User-specific 115cloud credentials

### Service Layer

- **`AuthService`**: Handles user registration, login, and session management
- **`CredentialsService`**: Manages 115cloud credentials per user
- **`Drive115Service`**: Interfaces with 115cloud APIs using user credentials

### Handler Layer

- **`AuthHandler`**: Authentication REST endpoints
- **`CredentialsHandler`**: Credential management REST endpoints
- **`Drive115Handler`**: 115cloud operations REST endpoints
- **`HealthHandler`**: System health monitoring

## Deployment

### Docker Compose Production

```bash
# Create production configuration
cp config.yaml.example config.yml
# Edit with production settings

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

### Manual Deployment

```bash
# Build for production
CGO_ENABLED=0 GOOS=linux go build -o cloud-driver cmd/cloud-driver/main.go

# Set up systemd service (example)
sudo cp cloud-driver /usr/local/bin/
sudo cp config.yml /etc/cloud-driver/
# Configure systemd service file
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Ensure code is formatted (`go fmt ./...`)
5. Run tests (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go community best practices
- Add tests for new features
- Update documentation for API changes
- Use meaningful commit messages
- Ensure database migrations are backward compatible

## Security Considerations

- **Password Hashing**: User passwords are securely hashed using bcrypt
- **Session Management**: Secure session tokens with expiration
- **Credential Storage**: 115cloud credentials are encrypted at rest
- **Input Validation**: All API inputs are validated and sanitized
- **HTTPS**: Use HTTPS in production environments

## Troubleshooting

### Common Issues

1. **Database Connection Issues**

   ```bash
   # Check PostgreSQL is running
   pg_isready -h localhost -p 5432

   # Verify database exists
   psql -l | grep cloud_driver
   ```

2. **Port Already in Use**

   ```bash
   # Find process using port 8081
   lsof -i :8081

   # Kill the process or change port in config.yml
   ```

3. **115Cloud API Errors**
   - Verify credentials are valid and active
   - Check if 115cloud session hasn't expired
   - Ensure credentials are set as active in the database

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [115driver](https://github.com/SheltonZhu/115driver) - Go client library for 115cloud
- [Echo](https://echo.labstack.com/) - High performance Go web framework
- [SQLC](https://sqlc.dev/) - Type-safe SQL query generator
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Air](https://github.com/air-verse/air) - Live reloading for Go apps

## Support

If you encounter any issues or have questions:

1. Check the [Issues](../../issues) page for existing solutions
2. Create a new issue with:
   - Detailed description of the problem
   - Steps to reproduce
   - Configuration (without sensitive data)
   - Error logs and stack traces
3. For security vulnerabilities, please email instead of creating public issues

---

**⚠️ Disclaimer**: This project is not affiliated with 115cloud. Use responsibly and in accordance with 115cloud's terms of service. Ensure you have proper authorization to access and manage the cloud storage accounts you configure.

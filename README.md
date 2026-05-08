# 1337b04rd

> An anonymous imageboard backend engineered with architectural excellence

[![Go](https://img.shields.io/badge/Go-1.23+-blue?style=flat-square)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-336791?style=flat-square)](https://www.postgresql.org/)

1337b04rd is an anonymous imageboard backend written in Go. The system provides functionality for creating threads, posting comments, and attaching images, while automatically assigning pseudo-anonymous identities based on the Rick and Morty API. 

The project strictly follows **Hexagonal Architecture** (Ports and Adapters) principles and relies exclusively on the Go standard library, with the sole exception of the PostgreSQL database driver.

## Architecture

The project is structured into distinct layers to enforce separation of concerns:

*   **Domain (`internal/domain`)**: Contains business entities (`Post`, `Comment`, `UserSession`) and defines interfaces (Ports) for external dependencies.
*   **Service (`internal/service`)**: Implements core business logic (`PostService`), orchestrates data flow between adapters, and manages the background archival process (`archive_worker.go`).
*   **Adapters (`internal/adapters`)**:
    *   **DB (`db`)**: Implements data persistence using PostgreSQL.
    *   **S3 (`s3`)**: Handles image uploads to AWS S3-compatible storage using a custom implementation of AWS Signature Version 4.
    *   **API (`api`)**: Integrates with the external Rick and Morty API to assign avatars, utilizing in-memory caching for performance.
    *   **HTTP (`http`)**: Manages HTTP routing, request parsing, session middleware, and HTML template rendering.
*   **Main (`cmd/1337b04rd`)**: The entry point that wires all dependencies together and starts the HTTP server with graceful shutdown capabilities.

## Features

*   **Thread and Comment Creation**: Users can create new discussion threads and reply to existing ones.
*   **Image Uploads**: Support for attaching images to both threads and comments, stored securely in an S3-compatible backend (e.g., MinIO).
*   **Anonymous Sessions**: Users are automatically assigned an identity (avatar and name) upon their first visit, persisting via a 7-day cookie. Users have the option to manually change their display name.
*   **Automated Archival System**: A background worker evaluates thread activity. Threads with no comments are archived after 10 minutes. Threads with comments are archived 15 minutes after the latest reply.
*   **Standard Library Driven**: Built entirely using the Go standard library (`net/http`, `html/template`, `crypto/hmac`, `database/sql`, etc.) to demonstrate proficiency with core language features without relying on external web frameworks.
*   **Code Quality**: Enforces strict code formatting using `gofumpt` and maintains unit test coverage for critical business logic.

## Prerequisites

*   Go 1.23 or higher
*   PostgreSQL 16+
*   S3-compatible object storage (e.g., MinIO, AWS S3)

## Configuration

The application is configured via environment variables. If not provided, it falls back to default values suitable for local development.

| Variable | Default Value | Description |
|---|---|---|
| `DB_DSN` | `postgres://postgres:postgres@localhost:5432/1337b04rd?sslmode=disable` | PostgreSQL connection string |
| `S3_ENDPOINT` | `http://localhost:9000` | S3 API endpoint |
| `S3_PUBLIC_ENDPOINT` | `http://localhost:9000` | Public URL for accessing S3 objects |
| `S3_ACCESS_KEY` | `minioadmin` | S3 Access Key ID |
| `S3_SECRET_KEY` | `minioadmin` | S3 Secret Access Key |
| `S3_REGION` | `us-east-1` | S3 Region |
| `TEMPLATE_DIR` | `./template` | Path to the HTML templates directory |

## Deployment and Usage

### 1. Database and Storage Setup

For local development, it is recommended to use Docker Compose to spin up the required infrastructure:

```sh
docker-compose up -d
```

This will start a PostgreSQL instance on port 5432 and a MinIO instance on ports 9000 (API) and 9001 (Web UI). Database migrations are applied automatically upon application startup.

### 2. Building the Application

Compile the binary using the Go toolchain:

```sh
go build -o 1337b04rd ./cmd/1337b04rd
```

### 3. Running the Server

Start the application specifying the desired port:

```sh
./1337b04rd --port 8080
```

The server will be accessible at `http://localhost:8080`.

## Contact & Links

| Platform | Link |
|----------|------|
| GitHub | [zzhassyn](https://github.com/zzhassyn) |
| LinkedIn | [Zhassyn Zhalynuly](https://www.linkedin.com/in/zhassyn-zhalynuly) |
| Email | [zhasynz05@mail.ru](mailto:zhasynz05@mail.ru) |

# Yota Backend

## Tech Stack

- **Language:** [Go](https://go.dev/) (1.24+)
- **Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL
- **ORM:** [GORM](https://gorm.io/)
- **Database Migrations:** [Atlas](https://atlasgo.io/)
- **Caching:** Redis
- **Object Storage:** S3-Compatible Storage (MinIO Client)
- **Authentication:** JWT (JSON Web Tokens) & Google OAuth
- **Payment Processing:** Midtrans
- **Task Scheduling:** Cron
- **Documentation:** Swagger (Swaggo)
- **Containerization:** Docker & Docker Compose

## Features

- **Authentication & Authorization**: Secure user login and signup via JWT, with Google OAuth integration and session management.
- **RESTful API**: Structured and versioned API endpoints serving the core platform.
- **Data Persistence**: Reliable data storage using PostgreSQL with structured migration support via Atlas.
- **Payment Gateway**: Integrated with Midtrans for secure payment processing.
- **Caching Layer**: Redis integration for high-performance caching.
- **File Storage**: S3-compatible object storage integration.
- **Rate Limiting**: Built-in protection against API abuse.
- **Task Scheduling**: Background job scheduling using Cron.
- **API Documentation**: Interactive API documentation generated automatically via Swagger UI.
- **Docker Support**: Configured with Docker and Docker Compose for streamlined deployment and local development.

## Prerequisites

Ensure you have the following installed on your machine:

- [Go](https://go.dev/dl/) (v1.24 or later)
- [Docker](https://www.docker.com/products/docker-desktop) & Docker Compose

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Vilamuzz/yota-backend.git
cd yota-backend
```

### 2. Configure Environment

Copy the example environment file and update the values to match your local setup:

```bash
cp .env.example .env
```

> [!IMPORTANT]
> Make sure to update critical security keys in `.env`, especially `JWT_SECRET_KEY`, `GOOGLE_CLIENT_*`, payment credentials, and database credentials for production environments.

### 3. Run with Docker (Recommended)

The easiest way to start the application and its external dependencies (Postgres, Redis, Object Storage) is using Docker Compose:

```bash
docker-compose up -d --build
```

The server will start and be accessible at `http://localhost:8080`.

### 4. Local Development (Manual Setup)

If you prefer running the Go application locally outside of Docker:

1. **Start Dependencies**: Ensure PostgreSQL and Redis are running. You can start just the dependencies using Docker:
   ```bash
   docker-compose up -d postgres redis
   ```
2. **Download Modules**:
   ```bash
   go mod download
   ```
3. **Run Application**:
   ```bash
   go run cmd/server/main.go
   ```

## API Documentation

Once the server is running, you can access the full API documentation using the Swagger UI interface:

[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Project Structure

```text
yota-backend/
├── app/             # Application business logic (Models, Handlers, Services)
├── cmd/             # Entry points for the application
├── config/          # Configuration loading and setup
├── docs/            # Swagger documentation files
├── internal/        # Private application and library code
├── migrations/      # Database migrations
├── pkg/             # Public libraries/utils
├── .env.example     # Environment variable template
├── docker-compose.yml
└── go.mod
```

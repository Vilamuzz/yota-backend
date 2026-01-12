# Yota Backend

Yota Backend is a Go-based REST API service with JWT authentication, PostgreSQL database, and auto-generated Swagger documentation.

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose installed
- (Optional) [Make](https://www.gnu.org/software/make/) for easier commands

### 1. Clone the Repository

```sh
git clone <repository-url>
cd yota-backend
```

### 2. Start the Services

```sh
docker-compose up -d
# or, if you have make installed:
make up
```

### 3. Check Services

```sh
docker-compose ps
# or
make logs
```

### 4. Access the Application

- **API Base URL:** [http://localhost:8080](http://localhost:8080)
- **Swagger Docs:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Project Structure

```
.
├── app/                # Application logic (delivery, repository, usecase)
├── config/             # Configuration (DB, JWT)
├── domain/             # Domain models, interfaces, requests
├── docs/               # Swagger docs
├── pkg/                # Utilities (validation, response, jwt, etc.)
├── Dockerfile          # Docker build file
├── docker-compose.yml  # Docker Compose setup
├── Makefile            # Common development commands
├── main.go             # Application entrypoint
└── .env.example        # Example environment variables
```

## Configuration

- Environment variables are set in `docker-compose.yml` by default.
- To customize, copy `.env.example` to `.env` and edit as needed.
- Main variables:
  - `APP_NAME`, `APP_ENV`, `PORT`, `DB`, `JWT_SECRET_KEY_*`, `JWT_TTL`, etc.

## Useful Commands

With **Make**:

```sh
make up            # Start all services
make down          # Stop all services
make logs          # Show logs
make db-shell      # Access PostgreSQL shell
make restart-app   # Restart only the app
make swagger       # Generate Swagger docs
make test          # Run Go tests
```

With **Docker Compose**:

```sh
docker-compose up -d
docker-compose down
docker-compose logs -f
docker-compose exec postgres psql -U postgres -d yota_db
docker-compose exec app sh
```

## Development

- Code changes require rebuilding the app container:
  ```sh
  make restart-app
  # or
  docker-compose restart app
  ```
- To run tests:
  ```sh
  make test
  # or
  docker-compose exec app go test ./...
  ```

# Yota Backend

A robust and scalable backend service for the Yota application, built with Go.

## 🚀 Overview

Yota Backend is a high-performance RESTful API designed to power the Yota platform. It provides secure authentication, data management, and real-time capabilities.

## 🛠 Tech Stack

-   **Language:** [Go](https://go.dev/) (1.24+)
-   **Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
-   **Database:** PostgreSQL
-   **ORM:** [GORM](https://gorm.io/)
-   **Caching:** Redis
-   **Authentication:** JWT (JSON Web Tokens) & Google OAuth
-   **Documentation:** Swagger (Swaggo)
-   **Containerization:** Docker & Docker Compose

## ✨ Features

-   **Authentication & Authorization**: Secure user login/signup via JWT and Google OAuth integration.
-   **RESTful API**: Structured and versioned API endpoints.
-   **Data Persistence**: Reliable data storage using PostgreSQL with migration support.
-   **Caching Layer**: Redis integration for high-performance caching and session management.
-   **Rate Limiting**: Built-in protection against API abuse.
-   **API Documentation**: Interactive API documentation via Swagger UI.
-   **Docker Ready**: tailored `Dockerfile` and `docker-compose.yml` for easy deployment.

## 📋 Prerequisites

Ensure you have the following installed on your machine:

-   [Go](https://go.dev/dl/) (v1.24 or later)
-   [Docker](https://www.docker.com/products/docker-desktop) & Docker Compose
-   [Make](https://www.gnu.org/software/make/) (Optional, for running makefile commands if available)

## ⚡ Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Vilamuzz/yota-backend.git
cd yota-backend
```

### 2. Configure Environment

Copy the example environment file and update the values as needed:

```bash
cp .env.example .env
```

> [!IMPORTANT]
> Make sure to update critical security keys in `.env`, especially `JWT_SECRET_KEY`, `GOOGLE_CLIENT_*`, and database credentials for production environments.

### 3. Run with Docker (Recommended)

The easiest way to start the application and its dependencies (Postgres, Redis) is using Docker Compose:

```bash
docker-compose up -d --build
```

The server will start at `http://localhost:8080`.

### 4. Local Development (Manual Setup)

If you prefer running Go locally:

1.  **Start Dependencies**: Ensure PostgreSQL and Redis are running (you can use `docker-compose up postgres redis -d`).
2.  **Download Modules**:
    ```bash
    go mod download
    ```
3.  **Run Application**:
    ```bash
    go run cmd/server/main.go
    ```

## 📚 API Documentation

Once the server is running, you can access the full API documentation using Swagger UI:

👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

## 📂 Project Structure

```
yota-backend/
├── app/             # Application business logic (Models, Handlers, Services)
├── cmd/             # Entry points for the application
├── config/          # Configuration loading and setup
├── docs/            # Swagger documentation files
├── internal/        # Private application and library code
├── pkg/             # Public libraries/utils
├── .env.example     # Environment variable template
├── docker-compose.yml
└── go.mod
```
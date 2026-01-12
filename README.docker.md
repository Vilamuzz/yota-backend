# Docker Setup Guide

## Prerequisites

- Docker installed ([Get Docker](https://docs.docker.com/get-docker/))
- Docker Compose installed (usually comes with Docker Desktop)

## Quick Start

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd yota-backend
   ```

2. **Start the services**

   ```bash
   make up
   # or
   docker-compose up -d
   ```

3. **Check if services are running**

   ```bash
   docker-compose ps
   ```

4. **View logs**

   ```bash
   make logs
   # or
   docker-compose logs -f
   ```

5. **Access the application**
   - API: http://localhost:8080
   - Swagger Documentation: http://localhost:8080/swagger/index.html

## Available Commands

```bash
make help          # Show all available commands
make up            # Start all services
make down          # Stop all services
make logs          # Show logs
make restart       # Restart services
make build         # Rebuild the application
make clean         # Remove all containers and volumes
make db-shell      # Access PostgreSQL shell
make swagger       # Generate swagger documentation
```

## Environment Configuration

The application uses environment variables defined in `docker-compose.yml`. To override them:

1. Copy `.env.docker` to `.env.local`:

   ```bash
   cp .env.docker .env.local
   ```

2. Edit `.env.local` with your values

3. Create `docker-compose.override.yml`:
   ```yaml
   version: "3.8"
   services:
     app:
       env_file:
         - .env.local
   ```

## Database Access

**Using Docker:**

```bash
make db-shell
# or
docker-compose exec postgres psql -U postgres -d yota_db
```

**Using external client:**

- Host: localhost
- Port: 5432
- Database: yota_db
- User: postgres
- Password: password

## Troubleshooting

### Services won't start

```bash
# Check logs
make logs

# Rebuild from scratch
make clean
make build
make up
```

### Port already in use

Edit `docker-compose.yml` and change the port mappings:

```yaml
ports:
  - "8081:8080" # Change 8081 to any available port
```

### Database connection issues

```bash
# Restart PostgreSQL
docker-compose restart postgres

# Check PostgreSQL logs
make logs-db
```

### Reset everything

```bash
# This will remove all data
make clean
make up
```

## Development Workflow

1. **Make code changes** in your local files
2. **Rebuild the app**:
   ```bash
   make restart-app
   ```
3. **View logs** to check for errors:
   ```bash
   make logs-app
   ```

## Production Deployment

The project includes GitHub Actions workflow for automated deployment. See [`.github/workflows/dev.yaml`](.github/workflows/dev.yaml) for details.

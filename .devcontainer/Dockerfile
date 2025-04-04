FROM mcr.microsoft.com/devcontainers/go:1.22-bullseye

# Install PostgreSQL client tools for database management
RUN apt-get update && apt-get install -y postgresql-client

# Install migrate CLI tool for database migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Set up environment
ENV CGO_ENABLED=0

WORKDIR /workspace
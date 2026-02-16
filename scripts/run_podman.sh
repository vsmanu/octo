#!/bin/bash
set -e

# Create a pod
podman pod create --name octo-pod -p 8080:8080 -p 5432:5432 || true

# Start TimescaleDB
echo "Starting TimescaleDB..."
podman run -d --pod octo-pod \
  --name timescaledb \
  --replace \
  -v timescaledb_data:/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=octo \
  docker.io/timescale/timescaledb:latest-pg16

# Wait for DB to be ready
echo "Waiting for TimescaleDB to be ready..."
sleep 10

# Build Master
echo "Building Master..."
podman build -t octo-master -f deployments/docker/Dockerfile.master .

# Start Master
echo "Starting Master..."
podman run -d --pod octo-pod \
  --name master \
  --replace \
  -v $(pwd)/config:/config:Z \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=password123 \
  -e DB_NAME=octo \
  -e CONFIG_PATH=/config/config.yml \
  octo-master

echo "Done! API is available at http://localhost:8080"

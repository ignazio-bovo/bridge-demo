#!/bin/bash

# Set working directory to project root
cd "$(dirname "$0")/.."

# Start squid database and processor
docker compose -f ./devops/docker/docker-compose-prod.yml up -d db &&
    while ! docker compose -f ./devops/docker/docker-compose-prod.yml exec db pg_isready -U postgres >/dev/null 2>&1; do
        echo "Waiting for database to be ready..."
        sleep 2
    done

# Start processor
docker compose -f ./devops/docker/docker-compose-prod.yml up -d sqd-processor --build

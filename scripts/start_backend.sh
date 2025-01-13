#!/bin/bash

# Change to backend directory
cd "$(dirname "$0")/../backend"

# Start squid database and processor
docker compose up -d db &&
    while ! docker compose exec db pg_isready -U postgres >/dev/null 2>&1; do
        echo "Waiting for database to be ready..."
        sleep 2
    done

# Apply database migrations
sqd migration:apply
sleep 5

# Run sqd up in the background
sqd run

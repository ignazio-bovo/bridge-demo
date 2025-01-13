#!/bin/bash

# Change to backend directory
cd "$(dirname "$0")/../backend"

# Start squid database and processor
echo "Starting Squid database and processor..."

# Run sqd up in the background
sqd up &

# Wait a few seconds to ensure database is ready
sleep 5

# Run sqd process ethereum
sqd process:eth

# Run sqd process subtensor
sqd process:subtensor

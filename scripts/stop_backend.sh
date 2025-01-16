#!/bin/bash

cd "$(dirname "$0")/.."
docker compose -f ./devops/docker/docker-compose.yml down

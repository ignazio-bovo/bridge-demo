#!/bin/bash

cd "$(dirname "$0")/.."
docker compose -f ./devops/docker/docker-compose-prod.yml down sqd-processor
docker compose -f ./devops/docker/docker-compose-prod.yml down db

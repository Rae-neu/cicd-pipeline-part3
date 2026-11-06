#!/bin/sh
set -eu

cd "$(dirname "$0")"

echo "Starting Go REST API and LocalStack..."

echo "API will be available at http://localhost:8081/tasks"

docker compose -p cicd-part3-app -f compose.app.yml up --build

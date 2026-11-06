#!/bin/sh
set -eu

cd "$(dirname "$0")"

echo "Starting Docker Compose test stack..."

trap 'docker compose -p cicd-part3-tests -f compose.test.yml down -v' EXIT

docker compose -p cicd-part3-tests -f compose.test.yml up \
  --build \
  --abort-on-container-exit \
  --exit-code-from tests

#!/bin/sh
set -eu
cd "$(dirname "$0")"
trap 'docker compose -f compose.app.yml down' EXIT
docker compose -f compose.app.yml up --build

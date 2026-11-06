
#!/bin/sh
set -eu

cd "$(dirname "$0")"

echo "Building REST API Docker image..."
docker build -f Dockerfile.api -t cicd-part2-api .

echo "Starting REST API on port 8081..."
docker run --rm -p 8081:8080 cicd-part2-api

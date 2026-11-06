
#!/bin/sh
set -eu

cd "$(dirname "$0")"

echo "Building test Docker image..."
docker build -f Dockerfile.test -t cicd-part2-test .

echo "Running automated tests..."
docker run --rm cicd-part2-test

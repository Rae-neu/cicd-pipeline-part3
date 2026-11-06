
# CI/CD Pipeline (Part 2) - Dockerized REST API

This project implements a Dockerized REST API using Go, with automated testing, GitHub Actions CI/CD, and Slack notifications.

It builds on Part 1 by adding REST endpoints, Docker containers, and automated workflow notifications.

## Features

- REST API supporting GET, POST, PUT, and DELETE
- Automated Go tests using httptest
- HTTP status code validation
- Separate Docker images for the API and tests
- Shell scripts for building and running containers
- GitHub Actions for automatic and manual CI runs
- Workflow summaries and Slack notifications

## Project Structure

- `api/main.go` - REST API implementation
- `api/main_test.go` - REST API tests
- `calculator.go` - Calculator from Part 1
- `calculator_test.go` - Original unit tests
- `Dockerfile.api` - Dockerfile for the API
- `Dockerfile.test` - Dockerfile for testing
- `run-api.sh` - Script to build and run the API
- `run-tests.sh` - Script to run Dockerized tests
- `.github/workflows/ci.yml` - CI/CD workflow
- `go.mod` - Go module configuration

## Requirements

- Go 1.27 or compatible version
- Docker Desktop with Docker Engine running
- Git

## REST API Endpoints

| Method | Endpoint | Description | Success Status |
|--------|----------|-------------|----------------|
| GET | /tasks | Get all tasks | 200 OK |
| GET | /tasks/{id} | Get a task by ID | 200 OK |
| POST | /tasks | Create a task | 201 Created |
| PUT | /tasks/{id} | Update a task | 200 OK |
| DELETE | /tasks/{id} | Delete a task | 204 No Content |

Invalid requests may return 400 Bad Request, 404 Not Found, or 405 Method Not Allowed.

Tasks are stored in memory and are reset when the API server restarts.

## Run Locally

Start the REST API:

```bash
go run ./api
```

The API runs at http://localhost:8080.

Run all tests with verbose output:

```bash
go test -v ./...
```

## Run with Docker

### REST API

Build the API image:

```bash
docker build -f Dockerfile.api -t cicd-part2-api .
```

Run the container:

```bash
docker run --rm -p 8081:8080 cicd-part2-api
```

Access the API at http://localhost:8081/tasks.

Alternatively, use the shell script in a Bash-compatible terminal:

```bash
sh run-api.sh
```

The API container runs until manually stopped.

### Automated Tests

Build the test image:

```bash
docker build -f Dockerfile.test -t cicd-part2-test .
```

Run the tests:

```bash
docker run --rm cicd-part2-test
```

Alternatively, use the shell script:

```bash
sh run-tests.sh
```

The test container exits with code 0 when all tests pass and a non-zero exit code when tests fail.

## CI/CD Pipeline

GitHub Actions runs the workflow on:

- Push
- Pull request
- Manual execution using workflow_dispatch

The workflow:

1. Checks out the repository.
2. Builds and runs Dockerized Go tests.
3. Builds the REST API Docker image.
4. Starts the API container and verifies its response.
5. Generates a GitHub Actions workflow summary.
6. Sends the workflow status and summary link to Slack.

## Slack Notifications

The workflow uses an Incoming Webhook to send notifications to the Slack channel #ci-notifications.

The webhook URL is stored securely as a GitHub Actions repository secret:

`SLACK_WEBHOOK_URL`

Notifications include the workflow status, repository, trigger, branch, and workflow summary link.

The webhook URL is not included in the source code.

## GitHub Repository

https://github.com/Rae-neu/cicd-pipeline-part2

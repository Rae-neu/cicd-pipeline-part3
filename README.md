# CI/CD Pipeline Part 3 — Orchestration and CRUD

This Go REST API uses **Docker Compose** to run an API and **LocalStack** (AWS DynamoDB and S3). Each task is stored as a DynamoDB item and as JSON at `s3://task-objects/tasks/<id>.json`.

## Requirements

Docker with Docker Compose v2+, Python 3. This project can be run entirely inside GitHub Codespaces. **No real AWS account is required.**

## Start the application

```bash
sh run-api.sh
# Alternative detached mode:
docker compose -f compose.app.yml up -d --build
```

The API is available at `http://localhost:8081/tasks`. LocalStack exposes port 4566. During startup, the init scripts in `localstack/init/ready.d` create the `tasks` DynamoDB table and `task-objects` S3 bucket. The API is launched only after LocalStack's health check passes.

## CRUD endpoints

| Method | Path | Success | Behavior |
|---|---|---|---|
| GET | `/tasks` | 200 | List tasks from DynamoDB |
| GET | `/tasks/{id}` | 200 | Read item from DynamoDB |
| POST | `/tasks` | 201 | Store in DynamoDB and S3 |
| PUT | `/tasks/{id}` | 200 | Update DynamoDB and S3 |
| DELETE | `/tasks/{id}` | 204 | Delete from DynamoDB and S3 |

Invalid requests return 400, missing items return 404, and unsupported methods return 405. Storage errors return 500.

Example POST:

```bash
curl -i -X POST http://localhost:8081/tasks -H 'Content-Type: application/json' -d '{"title":"Finish Part 3"}'
```

## Run all tests

```bash
# Dockerized Go unit tests (exit code reflects test success/failure)
sh run-tests.sh

# Full end-to-end API and AWS resource test (automatically starts/stops the stack)
python3 scripts/integration_test.py
```

The integration test verifies successful and invalid GET, POST, PUT and DELETE requests; it reads the corresponding DynamoDB records and S3 objects to confirm synchronization, then verifies both are deleted. It exits with a nonzero status on failure.

To stop the application:

```bash
docker compose -f compose.app.yml down
```

## CI/CD with GitHub Actions

`.github/workflows/ci.yml` runs on pushes, pull requests and manual dispatch. It starts the LocalStack test stack, runs the Go tests, stops it, then runs the end-to-end suite. GitHub Actions **fails if a test fails**, creates a workflow summary and sends Slack a result plus a brief excerpt of test output.

Configure Slack using the repository Actions secret `SLACK_WEBHOOK_URL` (a Slack Incoming Webhook). If the secret is absent, the notification step is skipped; setting it is required for Slack messages to actually arrive.

## Key files

- `api/main.go` — application entry point and standalone in-memory mode
- `api/part3_cloud.go` — cloud-backed REST handlers used by Docker Compose
- `api/aws.go`, `api/dynamodb.go`, `api/s3.go` — LocalStack clients and persistence
- `localstack/init/ready.d/` — local AWS resource initialization
- `compose.app.yml` / `run-api.sh` — application stack and launcher
- `compose.test.yml` / `run-tests.sh` — unit test stack and launcher
- `scripts/integration_test.py` — end-to-end checks
- `.github/workflows/ci.yml` — automated pipeline and Slack notification

Repository: https://github.com/Rae-neu/cicd-pipeline-part3

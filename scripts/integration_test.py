#!/usr/bin/env python3
"""End-to-end REST CRUD + DynamoDB + S3 verification using LocalStack."""
import json
import subprocess
import sys
import time
import urllib.error
import urllib.request

COMPOSE = ["docker", "compose", "-f", "compose.app.yml"]
BASE = "http://127.0.0.1:8081"

def run(args, **kwargs):
    return subprocess.run(args, check=True, text=True, **kwargs)

def aws(*args):
    result = run(COMPOSE + ["exec", "-T", "localstack", "awslocal", *args], capture_output=True)
    return json.loads(result.stdout) if result.stdout.strip() else {}

def s3_json(key):
    path = "/tmp/cicd-part3-inspect.json"
    aws("s3api", "get-object", "--bucket", "task-objects", "--key", key, path)
    result = run(COMPOSE + ["exec", "-T", "localstack", "cat", path], capture_output=True)
    return json.loads(result.stdout)

def request(method, path, payload=None):
    body = json.dumps(payload).encode() if payload is not None else None
    req = urllib.request.Request(BASE + path, data=body, method=method,
                                 headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=8) as resp:
            return resp.status, resp.read().decode()
    except urllib.error.HTTPError as ex:
        return ex.code, ex.read().decode()

def check(name, method, path, code, payload=None):
    status, body = request(method, path, payload)
    assert status == code, f"{name}: expected HTTP {code}, got {status}: {body}"
    print(f"PASS {name} (HTTP {status})", flush=True)
    return json.loads(body) if body.strip().startswith(("{", "[")) else body

def db_task(id):
    obj = aws("dynamodb", "get-item", "--table-name", "tasks",
              "--key", json.dumps({"id": {"N": str(id)}}))
    return obj.get("Item")

def main():
    try:
        run(COMPOSE + ["up", "-d", "--build"])
        for attempt in range(30):
            try:
                request("GET", "/tasks")
                break
            except Exception:
                time.sleep(1)
        else:
            raise RuntimeError("API did not start")
        check("GET list", "GET", "/tasks", 200)
        check("GET nonexistent task", "GET", "/tasks/999999", 404)
        check("GET invalid ID", "GET", "/tasks/abc", 400)
        check("POST missing title", "POST", "/tasks", 400, {})
        check("Unsupported PATCH", "PATCH", "/tasks", 405)
        created = check("POST valid task", "POST", "/tasks", 201, {"title": "Part 3 integration"})
        id = created["id"]
        key = f"tasks/{id}.json"
        assert db_task(id)["title"]["S"] == "Part 3 integration", "DynamoDB POST mismatch"
        assert s3_json(key) == created, "S3 POST mismatch"
        print("PASS POST persisted in both DynamoDB and S3", flush=True)
        assert check("GET task by ID", "GET", f"/tasks/{id}", 200) == created
        assert created in check("GET list includes task", "GET", "/tasks", 200)
        check("PUT invalid title", "PUT", f"/tasks/{id}", 400, {})
        updated = check("PUT update task", "PUT", f"/tasks/{id}", 200, {"title": "Updated integration"})
        assert updated["title"] == "Updated integration"
        assert db_task(id)["title"]["S"] == updated["title"], "DynamoDB PUT mismatch"
        assert s3_json(key) == updated, "S3 PUT mismatch"
        print("PASS PUT updated both DynamoDB and S3", flush=True)
        check("DELETE task", "DELETE", f"/tasks/{id}", 204)
        check("GET deleted task", "GET", f"/tasks/{id}", 404)
        assert not db_task(id), "DynamoDB record still exists"
        contents = aws("s3api", "list-objects-v2", "--bucket", "task-objects", "--prefix", key)
        assert contents.get("KeyCount", 0) == 0, "S3 object still exists"
        print("PASS DELETE removed both DynamoDB item and S3 object", flush=True)
        print("ALL PART 3 INTEGRATION TESTS PASSED", flush=True)
    finally:
        run(COMPOSE + ["down", "--remove-orphans"], stdout=subprocess.DEVNULL)

if __name__ == "__main__":
    try:
        main()
    except Exception as err:
        print(f"FAIL: {err}", file=sys.stderr, flush=True)
        sys.exit(1)

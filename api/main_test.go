package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestRouter() http.Handler {
	mu.Lock()
	tasks = []Task{}
	nextID = 1
	mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", tasksHandler)
	mux.HandleFunc("/tasks/", taskByIDHandler)
	return mux
}

func sendRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func checkStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("Expected HTTP %d, got HTTP %d", want, got)
	}
}

func TestTaskLifecycle(t *testing.T) {
	router := setupTestRouter()

	t.Run("GET empty tasks", func(t *testing.T) {
		resp := sendRequest(router, http.MethodGet, "/tasks", "")
		checkStatus(t, resp.Code, http.StatusOK)

		if strings.TrimSpace(resp.Body.String()) != "[]" {
			t.Errorf("Expected empty task list, got %s", resp.Body.String())
		}
	})

	t.Run("POST create task", func(t *testing.T) {
		resp := sendRequest(router, http.MethodPost, "/tasks",
			`{"title":"Finish homework"}`)
		checkStatus(t, resp.Code, http.StatusCreated)

		var task Task
		if err := json.Unmarshal(resp.Body.Bytes(), &task); err != nil {
			t.Fatal(err)
		}
		if task.ID != 1 || task.Title != "Finish homework" {
			t.Errorf("Unexpected task: %+v", task)
		}
	})

	t.Run("GET task by ID", func(t *testing.T) {
		resp := sendRequest(router, http.MethodGet, "/tasks/1", "")
		checkStatus(t, resp.Code, http.StatusOK)
	})

	t.Run("PUT update task", func(t *testing.T) {
		resp := sendRequest(router, http.MethodPut, "/tasks/1",
			`{"title":"Complete Part 2"}`)
		checkStatus(t, resp.Code, http.StatusOK)

		var task Task
		if err := json.Unmarshal(resp.Body.Bytes(), &task); err != nil {
			t.Fatal(err)
		}
		if task.Title != "Complete Part 2" {
			t.Errorf("Task was not updated: %+v", task)
		}
	})

	t.Run("GET updated task list", func(t *testing.T) {
		resp := sendRequest(router, http.MethodGet, "/tasks", "")
		checkStatus(t, resp.Code, http.StatusOK)

		var result []Task
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result) != 1 || result[0].Title != "Complete Part 2" {
			t.Errorf("Unexpected task list: %+v", result)
		}
	})

	t.Run("DELETE task", func(t *testing.T) {
		resp := sendRequest(router, http.MethodDelete, "/tasks/1", "")
		checkStatus(t, resp.Code, http.StatusNoContent)
	})

	t.Run("GET deleted task returns 404", func(t *testing.T) {
		resp := sendRequest(router, http.MethodGet, "/tasks/1", "")
		checkStatus(t, resp.Code, http.StatusNotFound)
	})
}

func TestInvalidRequests(t *testing.T) {
	router := setupTestRouter()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"Invalid POST", "POST", "/tasks", `{}`, 400},
		{"Malformed JSON", "POST", "/tasks", `{bad`, 400},
		{"Invalid ID", "GET", "/tasks/abc", "", 400},
		{"Missing task", "GET", "/tasks/999", "", 404},
		{"Missing task update", "PUT", "/tasks/999",
			`{"title":"Updated"}`, 404},
		{"Unsupported method", "PATCH", "/tasks", "", 405},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := sendRequest(router, tc.method, tc.path, tc.body)
			checkStatus(t, resp.Code, tc.status)
		})
	}
}

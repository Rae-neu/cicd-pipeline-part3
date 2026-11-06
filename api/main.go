package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var (
	tasks  = []Task{}
	nextID = 1
	mu     sync.Mutex
)

func main() {
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/", taskByIDHandler)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET all tasks and POST a new task
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		defer mu.Unlock()
		json.NewEncoder(w).Encode(tasks)

	case http.MethodPost:
		var input Task
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil ||
			strings.TrimSpace(input.Title) == "" {
			http.Error(w, "Invalid task", http.StatusBadRequest)
			return
		}

		mu.Lock()
		task := Task{ID: nextID, Title: input.Title}
		nextID++
		tasks = append(tasks, task)
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET, PUT and DELETE a task by ID
func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	index := -1
	for i, task := range tasks {
		if task.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(tasks[index])

	case http.MethodPut:
		var input Task
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil ||
			strings.TrimSpace(input.Title) == "" {
			http.Error(w, "Invalid task", http.StatusBadRequest)
			return
		}

		tasks[index].Title = input.Title
		json.NewEncoder(w).Encode(tasks[index])

	case http.MethodDelete:
		tasks = append(tasks[:index], tasks[index+1:]...)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

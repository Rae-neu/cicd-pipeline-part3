package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var part3NextID atomic.Int64

func registerPart3CloudHandlers(mux *http.ServeMux) error {
	ctx := context.Background()
	db, storage, err := newAWSClients(ctx)
	if err != nil {
		return err
	}
	previous, err := listTasks(ctx, db)
	if err != nil {
		return err
	}
	var maxID int64
	for _, task := range previous {
		if int64(task.ID) > maxID {
			maxID = int64(task.ID)
		}
	}
	part3NextID.Store(maxID)
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		part3TasksHandler(w, r, db, storage)
	})
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		part3TaskHandler(w, r, db, storage)
	})
	return nil
}

func part3StorageError(w http.ResponseWriter, err error) {
	log.Printf("Cloud storage error: %v", err)
	http.Error(w, "Cloud storage error", http.StatusInternalServerError)
}

func part3ReadInput(r *http.Request) (string, error) {
	var input Task
	err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20)).Decode(&input)
	if err != nil {
		return "", err
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return "", errors.New("title must not be empty")
	}
	return title, nil
}

func part3TasksHandler(w http.ResponseWriter, r *http.Request, db *dynamodb.Client, storage *s3.Client) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		results, err := listTasks(r.Context(), db)
		if err != nil {
			part3StorageError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(results)
	case http.MethodPost:
		title, err := part3ReadInput(r)
		if err != nil {
			http.Error(w, "Invalid task", http.StatusBadRequest)
			return
		}
		task := Task{ID: int(part3NextID.Add(1)), Title: title}
		if err := saveTaskObject(r.Context(), storage, task); err != nil {
			part3StorageError(w, err)
			return
		}
		if err := saveTask(r.Context(), db, task); err != nil {
			_ = deleteTaskObject(r.Context(), storage, task.ID)
			part3StorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(task)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func part3TaskHandler(w http.ResponseWriter, r *http.Request, db *dynamodb.Client, storage *s3.Client) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/tasks/"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	task, found, err := getTask(r.Context(), db, id)
	if err != nil {
		part3StorageError(w, err)
		return
	}
	if !found {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(task)
	case http.MethodPut:
		title, err := part3ReadInput(r)
		if err != nil {
			http.Error(w, "Invalid task", http.StatusBadRequest)
			return
		}
		updated := Task{ID: id, Title: title}
		if err := saveTaskObject(r.Context(), storage, updated); err != nil {
			part3StorageError(w, err)
			return
		}
		if err := saveTask(r.Context(), db, updated); err != nil {
			_ = saveTaskObject(r.Context(), storage, task)
			part3StorageError(w, err)
			return
		}
		_ = json.NewEncoder(w).Encode(updated)
	case http.MethodDelete:
		if err := deleteTaskObject(r.Context(), storage, id); err != nil {
			part3StorageError(w, err)
			return
		}
		if err := deleteTask(r.Context(), db, id); err != nil {
			_ = saveTaskObject(r.Context(), storage, task)
			part3StorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

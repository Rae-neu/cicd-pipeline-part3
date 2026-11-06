package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func bucketName() string {
	return os.Getenv("S3_BUCKET")
}

func taskObjectKey(id int) string {
	return fmt.Sprintf("tasks/%d.json", id)
}

func saveTaskObject(ctx context.Context, storage *s3.Client, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	_, err = storage.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName()),
		Key:         aws.String(taskObjectKey(task.ID)),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	return err
}

func getTaskObject(ctx context.Context, storage *s3.Client, id int) (Task, error) {
	result, err := storage.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName()),
		Key:    aws.String(taskObjectKey(id)),
	})
	if err != nil {
		return Task{}, err
	}
	defer result.Body.Close()

	var task Task
	if err := json.NewDecoder(result.Body).Decode(&task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func deleteTaskObject(ctx context.Context, storage *s3.Client, id int) error {
	_, err := storage.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName()),
		Key:    aws.String(taskObjectKey(id)),
	})
	return err
}

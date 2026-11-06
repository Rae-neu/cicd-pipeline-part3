package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func tableName() string {
	return os.Getenv("DYNAMODB_TABLE")
}

func saveTask(ctx context.Context, db *dynamodb.Client, task Task) error {
	_, err := db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName()),
		Item: map[string]types.AttributeValue{
			"id":    &types.AttributeValueMemberN{Value: strconv.Itoa(task.ID)},
			"title": &types.AttributeValueMemberS{Value: task.Title},
		},
	})
	return err
}

func decodeTask(item map[string]types.AttributeValue) (Task, error) {
	idAttr, ok := item["id"].(*types.AttributeValueMemberN)
	if !ok {
		return Task{}, fmt.Errorf("invalid task ID")
	}
	id, err := strconv.Atoi(idAttr.Value)
	if err != nil {
		return Task{}, err
	}

	titleAttr, ok := item["title"].(*types.AttributeValueMemberS)
	if !ok {
		return Task{}, fmt.Errorf("invalid task title")
	}

	return Task{ID: id, Title: titleAttr.Value}, nil
}

func getTask(ctx context.Context, db *dynamodb.Client, id int) (Task, bool, error) {
	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberN{Value: strconv.Itoa(id)},
		},
	})
	if err != nil {
		return Task{}, false, err
	}
	if len(result.Item) == 0 {
		return Task{}, false, nil
	}
	task, err := decodeTask(result.Item)
	return task, true, err
}

func listTasks(ctx context.Context, db *dynamodb.Client) ([]Task, error) {
	tasks := []Task{}

	paginator := dynamodb.NewScanPaginator(db, &dynamodb.ScanInput{
		TableName: aws.String(tableName()),
	})

	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range result.Items {
			task, err := decodeTask(item)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, task)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func deleteTask(ctx context.Context, db *dynamodb.Client, id int) error {
	_, err := db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName()),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberN{Value: strconv.Itoa(id)},
		},
	})
	return err
}

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"task-runner-service/internal/config"
	"task-runner-service/internal/service"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewStorage(cfg config.RedisConfig) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

func (s *RedisStorage) SaveTask(ctx context.Context, task service.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, "tasks", task.ID, data)
	pipe.SAdd(ctx, fmt.Sprintf("tasks:status:%s", task.Status), task.ID)
	pipe.Expire(ctx, task.ID, 72*time.Hour) // TTL 3 дня

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to save task in Redis: %w", err)
	}

	return nil
}

func (s *RedisStorage) GetTask(ctx context.Context, id string) (*service.Task, error) {
	data, err := s.client.HGet(ctx, "tasks", id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task from Redis: %w", err)
	}

	var task service.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

func (s *RedisStorage) GetTasks(ctx context.Context, status string, limit, offset int) ([]service.Task, error) {
	var taskIDs []string
	var err error

	if status != "" {
		taskIDs, err = s.client.SMembers(ctx, fmt.Sprintf("tasks:status:%s", status)).Result()
	} else {
		taskIDs, err = s.client.HKeys(ctx, "tasks").Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get task IDs: %w", err)
	}

	if offset >= len(taskIDs) {
		return []service.Task{}, nil
	}

	end := offset + limit
	if end > len(taskIDs) {
		end = len(taskIDs)
	}
	taskIDs = taskIDs[offset:end]

	var tasks []service.Task
	for _, id := range taskIDs {
		task, err := s.GetTask(ctx, id)
		if err != nil {
			continue
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

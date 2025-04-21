package service

import (
	"context"
	"fmt"
	"time"

	v1 "task-runner-service/internal/api/v1"

	"github.com/RichardKnop/machinery/v1/tasks"
)

type Storage interface {
	SaveTask(ctx context.Context, task Task) error
	GetTask(ctx context.Context, id string) (*Task, error)
	GetTasks(ctx context.Context, status string, limit, offset int) ([]Task, error)
}

type Task struct {
	ID        string
	Name      string
	Args      []tasks.Arg
	Status    string
	CreatedAt time.Time
	Result    interface{}
	Error     error
}

type RunnerService struct {
	server  MachineryServer
	storage Storage
}

func NewRunnerService(server MachineryServer, storage Storage) *RunnerService {
	return &RunnerService{
		server:  server,
		storage: storage,
	}
}

func (s *RunnerService) SendTask(ctx context.Context, name string, args []tasks.Arg, queue string) (string, error) {
	signature := &tasks.Signature{
		Name: name,
		Args: args,
	}

	if queue != "" {
		signature.RoutingKey = queue
	}

	asyncResult, err := s.server.SendTask(signature)
	if err != nil {
		return "", fmt.Errorf("failed to send task: %w", err)
	}

	task := Task{
		ID:        asyncResult.GetState().TaskUUID,
		Name:      name,
		Args:      args,
		Status:    tasks.StatePending,
		CreatedAt: time.Now(),
	}

	if err := s.storage.SaveTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to save task metadata: %w", err)
	}

	return task.ID, nil
}

func (s *RunnerService) GetTaskStatus(ctx context.Context, id string) (*v1.TaskResponse, error) {
	task, err := s.storage.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	state, err := s.server.GetBackend().GetState(task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task state: %w", err)
	}

	var result interface{}
	var errorMsg string
	if task.Error != nil {
		errorMsg = task.Error.Error()
	}
	result, err = s.retrieveResultFromBackend(task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve result from backend: %w", err)
	}
	return &v1.TaskResponse{
		ID:        task.ID,
		Status:    state.State,
		Result:    result,
		Error:     errorMsg,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *RunnerService) retrieveResultFromBackend(taskID string) (interface{}, error) {
	state, err := s.server.GetBackend().GetState(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task state: %w", err)
	}

	if state.IsCompleted() {

		if len(state.Results) > 0 && state.Results[0] != nil {
			return state.Results[0].Value, nil
		}
		return nil, nil
	}

	return nil, nil
}

func (s *RunnerService) GetTasks(ctx context.Context, status string, limit, offset int) ([]v1.TaskResponse, error) {
	tasks, err := s.storage.GetTasks(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var responses []v1.TaskResponse
	for _, task := range tasks {
		var errorMsg string
		if task.Error != nil {
			errorMsg = task.Error.Error()
		}

		responses = append(responses, v1.TaskResponse{
			ID:        task.ID,
			Name:      task.Name,
			Status:    task.Status,
			Result:    task.Result,
			Error:     errorMsg,
			CreatedAt: task.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

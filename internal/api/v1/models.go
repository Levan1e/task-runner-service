package v1

import (
	"context"

	"github.com/RichardKnop/machinery/v1/tasks"
)

type Handler struct {
	taskService TaskService
}

type TaskRequest struct {
	Name  string      `json:"name"`
	Args  []tasks.Arg `json:"args"`
	Queue string      `json:"queue,omitempty"`
}

type TaskService interface {
	SendTask(ctx context.Context, name string, args []tasks.Arg, queue string) (string, error)
	GetTaskStatus(ctx context.Context, id string) (*TaskResponse, error)
	GetTasks(ctx context.Context, status string, limit, offset int) ([]TaskResponse, error)
}

type TaskResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name,omitempty"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}

package mocks

import (
	"context"
	v1 "task-runner-service/internal/api/v1"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) SendTask(ctx context.Context, name string, args []tasks.Arg, queue string) (string, error) {
	argsList := m.Called(ctx, name, args, queue)
	return argsList.String(0), argsList.Error(1)
}

func (m *MockTaskService) GetTaskStatus(ctx context.Context, id string) (*v1.TaskResponse, error) {
	argsList := m.Called(ctx, id)
	return argsList.Get(0).(*v1.TaskResponse), argsList.Error(1)
}

func (m *MockTaskService) GetTasks(ctx context.Context, status string, limit, offset int) ([]v1.TaskResponse, error) {
	argsList := m.Called(ctx, status, limit, offset)
	return argsList.Get(0).([]v1.TaskResponse), argsList.Error(1)
}

func (m *MockTaskService) Initialize() {
	m.On("SendTask", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("mockedTaskID", nil)
	m.On("GetTaskStatus", mock.Anything, mock.Anything).Return(&v1.TaskResponse{
		ID:        "mockedTaskID",
		Status:    tasks.StatePending,
		Result:    nil,
		Error:     "",
		CreatedAt: "2025-04-21T15:10:00Z",
	}, nil)
	m.On("GetTasks", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]v1.TaskResponse{
		{
			ID:        "mockedTaskID",
			Name:      "mockedTask",
			Status:    tasks.StatePending,
			Result:    nil,
			Error:     "",
			CreatedAt: "2025-04-21T15:10:00Z",
		},
	}, nil)
}

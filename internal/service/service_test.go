package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"task-runner-service/internal/service"

	"github.com/RichardKnop/machinery/v1/backends/iface"
	"github.com/RichardKnop/machinery/v1/backends/result"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct{ mock.Mock }

func (m *MockStorage) SaveTask(ctx context.Context, task service.Task) error {
	return m.Called(ctx, task).Error(0)
}
func (m *MockStorage) GetTask(ctx context.Context, id string) (*service.Task, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*service.Task), args.Error(1)
}
func (m *MockStorage) GetTasks(ctx context.Context, status string, limit, offset int) ([]service.Task, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]service.Task), args.Error(1)
}

type MockBackend struct{ mock.Mock }

func (m *MockBackend) InitGroup(groupUUID string, taskUUIDs []string) error {
	return m.Called(groupUUID, taskUUIDs).Error(0)
}
func (m *MockBackend) GroupCompleted(groupUUID string, groupTaskCount int) (bool, error) {
	args := m.Called(groupUUID, groupTaskCount)
	return args.Bool(0), args.Error(1)
}
func (m *MockBackend) GroupTaskStates(groupUUID string, groupTaskCount int) ([]*tasks.TaskState, error) {
	args := m.Called(groupUUID, groupTaskCount)
	return args.Get(0).([]*tasks.TaskState), args.Error(1)
}
func (m *MockBackend) TriggerChord(groupUUID string) (bool, error) {
	args := m.Called(groupUUID)
	return args.Bool(0), args.Error(1)
}
func (m *MockBackend) SetStatePending(signature *tasks.Signature) error {
	return m.Called(signature).Error(0)
}
func (m *MockBackend) SetStateReceived(signature *tasks.Signature) error {
	return m.Called(signature).Error(0)
}
func (m *MockBackend) SetStateStarted(signature *tasks.Signature) error {
	return m.Called(signature).Error(0)
}
func (m *MockBackend) SetStateRetry(signature *tasks.Signature) error {
	return m.Called(signature).Error(0)
}
func (m *MockBackend) SetStateSuccess(signature *tasks.Signature, results []*tasks.TaskResult) error {
	return m.Called(signature, results).Error(0)
}
func (m *MockBackend) SetStateFailure(signature *tasks.Signature, errMsg string) error {
	return m.Called(signature, errMsg).Error(0)
}
func (m *MockBackend) PurgeState(taskUUID string) error {
	return m.Called(taskUUID).Error(0)
}
func (m *MockBackend) PurgeGroupMeta(groupUUID string) error {
	return m.Called(groupUUID).Error(0)
}
func (m *MockBackend) IsAMQP() bool {
	return m.Called().Bool(0)
}
func (m *MockBackend) GetState(taskUUID string) (*tasks.TaskState, error) {
	args := m.Called(taskUUID)
	return args.Get(0).(*tasks.TaskState), args.Error(1)
}

var _ iface.Backend = (*MockBackend)(nil)

type MockServer struct{ mock.Mock }

func (m *MockServer) SendTask(sig *tasks.Signature) (*result.AsyncResult, error) {
	args := m.Called(sig)
	return args.Get(0).(*result.AsyncResult), args.Error(1)
}
func (m *MockServer) GetBackend() iface.Backend {
	return m.Called().Get(0).(iface.Backend)
}

var _ service.MachineryServer = (*MockServer)(nil)

func TestSendTask(t *testing.T) {
	cases := []struct {
		name       string
		serverErr  error
		storageErr error
		wantErr    bool
	}{
		{"Success", nil, nil, false},
		{"ServerFail", errors.New("srv fail"), nil, true},
		{"StorageFail", nil, errors.New("save fail"), true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := new(MockStorage)
			srv := new(MockServer)
			be := new(MockBackend)

			if c.serverErr != nil {
				srv.
					On("SendTask", mock.AnythingOfType("*tasks.Signature")).
					Return((*result.AsyncResult)(nil), c.serverErr)
			} else {
				sig := &tasks.Signature{Name: "n", UUID: "id-1"}
				async := result.NewAsyncResult(sig, be)
				srv.
					On("SendTask", mock.AnythingOfType("*tasks.Signature")).
					Return(async, nil)
				be.
					On("GetState", "id-1").
					Return(&tasks.TaskState{TaskUUID: "id-1"}, nil)
				st.
					On("SaveTask", mock.Anything, mock.MatchedBy(func(task service.Task) bool {
						return task.ID == "id-1"
					})).
					Return(c.storageErr)
			}

			svc := service.NewRunnerService(srv, st)
			_, err := svc.SendTask(context.Background(), "n", nil, "")

			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			srv.AssertExpectations(t)
			be.AssertExpectations(t)
			st.AssertExpectations(t)
		})
	}
}

func TestGetTaskStatus(t *testing.T) {
	cases := []struct {
		name        string
		storageTask *service.Task
		storageErr  error
		state       *tasks.TaskState
		stateErr    error
		wantErr     bool
		wantStatus  string
		wantResult  interface{}
	}{
		{
			name:        "Success",
			storageTask: &service.Task{ID: "tid", Status: "pending"},
			storageErr:  nil,
			state: &tasks.TaskState{
				TaskUUID: "tid",
				State:    tasks.StateSuccess,
				Results:  []*tasks.TaskResult{{Value: 123}},
			},
			stateErr:   nil,
			wantErr:    false,
			wantStatus: tasks.StateSuccess,
			wantResult: 123,
		},
		{
			name:        "StorageFail",
			storageTask: nil,
			storageErr:  errors.New("no task"),
			wantErr:     true,
		},
		{
			name:        "StateFail",
			storageTask: &service.Task{ID: "tid"},
			storageErr:  nil,
			state:       nil,
			stateErr:    errors.New("backend error"),
			wantErr:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := new(MockStorage)
			srv := new(MockServer)
			be := new(MockBackend)
			st.On("GetTask", mock.Anything, "tid").Return(c.storageTask, c.storageErr)
			if c.storageErr == nil {
				srv.On("GetBackend").Return(be)
				be.On("GetState", "tid").Return(c.state, c.stateErr)
			}

			svc := service.NewRunnerService(srv, st)
			resp, err := svc.GetTaskStatus(context.Background(), "tid")

			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.wantStatus, resp.Status)
				assert.Equal(t, c.wantResult, resp.Result)
			}

			st.AssertExpectations(t)
			srv.AssertExpectations(t)
			be.AssertExpectations(t)
		})
	}
}

type stubStorage struct{}

func (s *stubStorage) SaveTask(ctx context.Context, task service.Task) error { return nil }
func (s *stubStorage) GetTask(ctx context.Context, id string) (*service.Task, error) {
	return &service.Task{ID: id}, nil
}
func (s *stubStorage) GetTasks(ctx context.Context, status string, limit, offset int) ([]service.Task, error) {
	return nil, nil
}

type stubBackend struct{}

func (b *stubBackend) InitGroup(string, []string) error                            { return nil }
func (b *stubBackend) GroupCompleted(string, int) (bool, error)                    { return true, nil }
func (b *stubBackend) GroupTaskStates(string, int) ([]*tasks.TaskState, error)     { return nil, nil }
func (b *stubBackend) TriggerChord(string) (bool, error)                           { return true, nil }
func (b *stubBackend) SetStatePending(*tasks.Signature) error                      { return nil }
func (b *stubBackend) SetStateReceived(*tasks.Signature) error                     { return nil }
func (b *stubBackend) SetStateStarted(*tasks.Signature) error                      { return nil }
func (b *stubBackend) SetStateRetry(*tasks.Signature) error                        { return nil }
func (b *stubBackend) SetStateSuccess(*tasks.Signature, []*tasks.TaskResult) error { return nil }
func (b *stubBackend) SetStateFailure(*tasks.Signature, string) error              { return nil }
func (b *stubBackend) PurgeState(string) error                                     { return nil }
func (b *stubBackend) PurgeGroupMeta(string) error                                 { return nil }
func (b *stubBackend) IsAMQP() bool                                                { return false }
func (b *stubBackend) GetState(taskUUID string) (*tasks.TaskState, error) {
	return &tasks.TaskState{TaskUUID: taskUUID, State: tasks.StatePending}, nil
}

var _ iface.Backend = (*stubBackend)(nil)

type stubServer struct{ backend iface.Backend }

func (s *stubServer) SendTask(sig *tasks.Signature) (*result.AsyncResult, error) {
	return result.NewAsyncResult(sig, s.backend), nil
}
func (s *stubServer) GetBackend() iface.Backend { return s.backend }

var _ service.MachineryServer = (*stubServer)(nil)

func BenchmarkSendTask(b *testing.B) {
	st := &stubStorage{}
	be := &stubBackend{}
	srv := &stubServer{backend: be}
	svc := service.NewRunnerService(srv, st)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.SendTask(context.Background(), fmt.Sprintf("bench-%d", i), nil, "")
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
	}
}

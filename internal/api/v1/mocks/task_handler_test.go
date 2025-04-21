package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "task-runner-service/internal/api/v1"
	"task-runner-service/internal/service/mocks"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostInQueue_Success(t *testing.T) {
	mockTaskService := new(mocks.MockTaskService)

	expectedTaskName := "test_task"
	expectedArgs := []tasks.Arg{{Type: "string", Value: "test_value"}}
	expectedTaskID := "generated_task_id_123"

	mockTaskService.On("SendTask",
		mock.Anything,
		expectedTaskName,
		expectedArgs,
		"",
	).Return(expectedTaskID, nil)

	handler := v1.NewHandler(mockTaskService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	requestBody := v1.TaskRequest{
		Name: expectedTaskName,
		Args: expectedArgs,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response v1.TaskResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, expectedTaskID, response.ID)
	assert.Equal(t, tasks.StatePending, response.Status)
	mockTaskService.AssertExpectations(t)
}

func TestPostInQueue_InvalidRequest(t *testing.T) {
	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "Empty body",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid JSON",
			requestBody:  `{"name": 123}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing task name",
			requestBody:  `{"args": [{"type": "string", "value": "test"}]}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTaskService := new(mocks.MockTaskService)

			handler := v1.NewHandler(mockTaskService)
			router := chi.NewRouter()
			handler.RegisterRoutes(router)

			req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader([]byte(tc.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			assert.Equal(t, tc.expectedCode, recorder.Code)
			mockTaskService.AssertNotCalled(t, "SendTask")
		})
	}
}

func TestGetTaskStatus_Success(t *testing.T) {
	mockTaskService := new(mocks.MockTaskService)
	taskID := "task_id_123"

	expectedResponse := &v1.TaskResponse{
		ID:     taskID,
		Status: tasks.StateSuccess,
	}

	mockTaskService.On("GetTaskStatus", mock.Anything, taskID).
		Return(expectedResponse, nil)

	handler := v1.NewHandler(mockTaskService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/api/v1/tasks/"+taskID, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response v1.TaskResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, *expectedResponse, response)
	mockTaskService.AssertExpectations(t)
}

func TestGetTasks_Success(t *testing.T) {
	mockTaskService := new(mocks.MockTaskService)

	expectedTasks := []v1.TaskResponse{
		{
			ID:     "task_1",
			Status: tasks.StateSuccess,
		},
	}

	mockTaskService.On("GetTasks",
		mock.Anything,
		"",
		10,
		0,
	).Return(expectedTasks, nil)

	handler := v1.NewHandler(mockTaskService)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest("GET", "/api/v1/tasks?limit=10", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Tasks []v1.TaskResponse `json:"tasks"`
	}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, expectedTasks, response.Tasks)
	mockTaskService.AssertExpectations(t)
}

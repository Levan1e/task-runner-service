package v1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "task-runner-service/internal/api/v1"
	"task-runner-service/mocks"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostInQueue(t *testing.T) {
	mockTaskService := new(mocks.MockTaskService)
	mockTaskService.On("SendTask", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("mockedTaskID", nil)

	handler := v1.NewHandler(mockTaskService)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	reqBody, _ := json.Marshal(v1.TaskRequest{
		Name: "taskName",
		Args: []tasks.Arg{{Type: "string", Value: "arg1"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockTaskService.AssertExpectations(t)
}

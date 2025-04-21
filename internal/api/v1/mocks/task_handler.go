package v1_test

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) RegisterRoutes(r chi.Router) {
	m.Called(r)
}

func (m *MockHandler) PostInQueue(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *MockHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *MockHandler) GetFilter(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *MockHandler) Initialize() {
	m.On("PostInQueue", mock.Anything, mock.Anything).Return(nil)
	m.On("GetStatus", mock.Anything, mock.Anything).Return(nil)
	m.On("GetFilter", mock.Anything, mock.Anything).Return(nil)
}

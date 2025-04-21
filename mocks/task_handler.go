package mocks

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

// MockHandler - мок для Handler
type MockHandler struct {
	mock.Mock
}

// RegisterRoutes - мок метода RegisterRoutes
func (m *MockHandler) RegisterRoutes(r chi.Router) {
	m.Called(r)
}

// PostInQueue - мок метода PostInQueue
func (m *MockHandler) PostInQueue(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// GetStatus - мок метода GetStatus
func (m *MockHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// GetFilter - мок метода GetFilter
func (m *MockHandler) GetFilter(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

// Initialize - функция для настройки моков
func (m *MockHandler) Initialize() {
	m.On("PostInQueue", mock.Anything, mock.Anything).Return(nil)
	m.On("GetStatus", mock.Anything, mock.Anything).Return(nil)
	m.On("GetFilter", mock.Anything, mock.Anything).Return(nil)
}

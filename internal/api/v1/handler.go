package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"task-runner-service/pkg/logger"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func NewHandler(taskService TaskService) *Handler {
	return &Handler{
		taskService: taskService,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", h.HealthCheck)
		r.Post("/tasks", h.PostInQueue)
		r.Get("/tasks/{id}", h.GetStatus)
		r.Get("/tasks", h.GetFilter)
	})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Обработка запроса HealthCheck")
	render.JSON(w, r, map[string]string{"status": "ok"})
}

func (h *Handler) PostInQueue(w http.ResponseWriter, r *http.Request) {
	logger.Info("Обработка запроса PostInQueue")
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Ошибка разбора запроса PostInQueue: %v", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "invalid request"})
		return
	}

	taskID, err := h.taskService.SendTask(r.Context(), req.Name, req.Args, req.Queue)
	if err != nil {
		logger.Errorf("Ошибка отправки задачи PostInQueue: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	logger.Infof("Задача создана: ID=%s, статус=%s", taskID, tasks.StatePending)
	render.JSON(w, r, TaskResponse{
		ID:     taskID,
		Status: tasks.StatePending,
	})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	logger.Info("Обработка запроса GetStatus")
	taskID := chi.URLParam(r, "id")
	logger.Infof("Получение статуса задачи: ID=%s", taskID)

	task, err := h.taskService.GetTaskStatus(r.Context(), taskID)
	if err != nil {
		logger.Errorf("Ошибка получения статуса задачи: %v", err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "task not found"})
		return
	}

	logger.Infof("Статус задачи получен: %+v", task)
	render.JSON(w, r, task)
}

func (h *Handler) GetFilter(w http.ResponseWriter, r *http.Request) {
	logger.Info("Обработка запроса GetFilter")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	logger.Infof("Получение списка задач: статус=%s, лимит=%d, смещение=%d", status, limit, offset)

	tasks, err := h.taskService.GetTasks(r.Context(), status, limit, offset)
	if err != nil {
		logger.Errorf("Ошибка получения списка задач: %v", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	logger.Infof("Список задач получен: количество=%d", len(tasks))
	render.JSON(w, r, map[string]interface{}{
		"tasks": tasks,
		"meta": map[string]int{
			"limit":  limit,
			"offset": offset,
			"total":  len(tasks),
		},
	})
}

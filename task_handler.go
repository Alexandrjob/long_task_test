package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"long_task_manager_test/models"
	"net/http"
)

type TaskHandler struct {
	manager *TaskManager
}

func NewTaskHandler(manager *TaskManager) *TaskHandler {
	return &TaskHandler{manager}
}

func (h *TaskHandler) Create(c *gin.Context) {
	var task models.Task
	if err := json.NewDecoder(c.Request.Body).Decode(&task); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	result := h.manager.AddTask(&task)
	h.manager.RunTask(context.Background(), result)

	h.writeJSONResponse(c.Writer, http.StatusOK, result)
}

func (h *TaskHandler) GetInfo(c *gin.Context) {
	id := c.Param("id")
	result, err := h.manager.GetTaskInfo(id)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(c.Writer, http.StatusOK, result)
}

func (h *TaskHandler) GetResult(c *gin.Context) {
	id := c.Param("id")
	var result, err = h.manager.GetTaskResult(c, id)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(c.Writer, http.StatusOK, result)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.manager.RemoveTask(id)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

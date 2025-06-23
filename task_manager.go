package main

import (
	"context"
	"fmt"
	"long_task_manager_test/models"
	"long_task_manager_test/models/tasks"
	"strconv"
	"sync"
	"time"
)

type taskV struct {
	Task   *models.Task
	Result tasks.TaskResult
	Error  error
}

type TaskManager struct {
	counterId int64
	tasks     map[string]*taskV
	mu        sync.Mutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{tasks: make(map[string]*taskV)}
}

func (t *TaskManager) AddTask(data *models.Task) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	uid := strconv.FormatInt(t.counterId, 10)
	nTask := *models.NewTask()
	nTask.Id = uid
	nTask.Type = data.Type
	nTask.Params = data.Params
	*t.tasks[uid].Task = nTask

	t.counterId++
	return uid
}

func (t *TaskManager) RunTask(ctx context.Context, id string) error {
	t.mu.Lock()
	v, exists := t.tasks[id]
	t.mu.Unlock()

	if !exists {
		return fmt.Errorf("task not found: %v", id)
	}

	v.Task.Run(ctx)
	return nil
}

func (t *TaskManager) RemoveTask(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	v, exists := t.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %v", id)
	}

	v.Task.CancelFunc()
	delete(t.tasks, id)
	t.counterId--
	return nil
}

func (t *TaskManager) GetTaskInfo(id string) (models.TaskInfo, error) {
	t.mu.Lock()
	v, exists := t.tasks[id]
	t.mu.Unlock()

	if !exists {
		return models.TaskInfo{}, fmt.Errorf("task not found: %v", id)
	}

	var duration time.Duration
	if v.Task.FinishedAt.IsZero() {
		duration = time.Now().Sub(v.Task.CreatedAt)
	} else {
		duration = v.Task.FinishedAt.Sub(v.Task.CreatedAt)
	}

	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	info := models.TaskInfo{
		Status:                v.Task.Status,
		CreatedAt:             v.Task.CreatedAt,
		TimeProgressing:       duration,
		TimeProgressingString: fmt.Sprintf("%02d:%02d", minutes, seconds),
	}
	return info, nil
}

func (t *TaskManager) GetTaskResult(ctx context.Context, id string) (tasks.TaskResult, error) {
	t.mu.Lock()
	v, exists := t.tasks[id]
	t.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("task not found: %v", id)
	}

	if v.Result != nil || v.Error != nil {
		return v.Result, v.Error
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-v.Task.Ctx.Done():
		select {
		case err := <-v.Task.ErrChan:
			err = fmt.Errorf("task error: %v", err)
			v.Error = err
			return nil, v.Error
		default:

			v.Error = v.Task.Ctx.Err()
			return nil, v.Error
		}
	case err := <-v.Task.ErrChan:
		v.Error = fmt.Errorf("task error: %v", err)
		return nil, v.Error
	case result := <-v.Task.ResultChan:
		v.Result = result
		return result, nil
	}
}

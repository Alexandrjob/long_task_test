package models

import (
	"context"
	"fmt"
	"long_task_manager_test/models/tasks"
	"sync"
	"time"
)

type TaskType string

const (
	DownloadTask TaskType = "download"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

type TaskInfo struct {
	Status                TaskStatus
	CreatedAt             time.Time
	TimeProgressing       time.Duration
	TimeProgressingString string
}

type Task struct {
	Id         string
	Type       TaskType
	Status     TaskStatus
	Params     map[string]interface{}
	CreatedAt  time.Time
	FinishedAt time.Time
	ResultChan chan tasks.TaskResult
	ErrChan    chan error

	mu sync.Mutex

	Ctx        context.Context
	CancelFunc context.CancelFunc
}

func NewTask() *Task {
	return &Task{
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		ResultChan: make(chan tasks.TaskResult, 1),
		ErrChan:    make(chan error, 1),
	}
}

func (t *Task) Run(ctx context.Context) {
	t.mu.Lock()
	t.Status = StatusRunning
	t.mu.Unlock()

	taskCtx, cancel := context.WithCancel(ctx)
	t.mu.Lock()
	t.Ctx = taskCtx
	t.CancelFunc = cancel
	t.mu.Unlock()

	go func() {
		defer func() {
			t.mu.Lock()
			t.FinishedAt = time.Now()
			close(t.ResultChan)
			close(t.ErrChan)
			t.mu.Unlock()
		}()

		select {
		case <-taskCtx.Done():
			t.mu.Lock()
			t.Status = StatusCancelled
			t.mu.Unlock()
			return
		default:
		}

		f := getFunc(t.Type)
		if f == nil {
			t.mu.Lock()
			t.Status = StatusFailed
			t.mu.Unlock()
			select {
			case t.ErrChan <- fmt.Errorf("unsupported task type: %v", t.Type):
			case <-taskCtx.Done():
			}
			return
		}

		result, err := f(taskCtx, t)
		if err != nil {
			t.mu.Lock()
			t.Status = StatusFailed
			t.mu.Unlock()
			select {
			case t.ErrChan <- err:
			case <-taskCtx.Done():
			}
			return
		}

		select {
		case t.ResultChan <- result:
			t.mu.Lock()
			t.Status = StatusCompleted
			t.mu.Unlock()
		case <-taskCtx.Done():
			t.mu.Lock()
			t.Status = StatusCancelled
			t.mu.Unlock()
		}
	}()
}

func getFunc(t TaskType) func(ctx context.Context, task *Task) (tasks.TaskResult, error) {
	switch t {
	case DownloadTask:
		return func(ctx context.Context, task *Task) (tasks.TaskResult, error) {
			return tasks.DownloadFile(task.Params["url"].(string))
		}
	default:
		return nil
	}
}

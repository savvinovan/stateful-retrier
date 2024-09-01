package main

import "time"

// Task is a struct that represents a task in the database.
// And stores the state of the function.
type Task struct {
	// ID is a unique identifier for the task. It could be int64 or uuid.
	ID int64 `json:"id"`
	// FunctionName is a key for the function.
	FunctionName string `json:"function_name"`
	// Payload is a JSON-encoded string.
	Payload string `json:"payload"`
	// Status can be "pending", "running", "completed", "failed", "terminated".
	Status TaskStatus `json:"status"`
	// RetryCount is the number of times the task has been retried.
	RetryCount int `json:"retry_count"`
	// CreatedAt is the time the task was created.
	CreatedAt time.Time
	// UpdatedAt is the time the task was last updated.
	UpdatedAt time.Time
	// CompletedAt is the time the task was completed.
	// It is useful for tracking the time it took to complete the task.
	// And we can delete the task after a certain period of time. And use partitioning to archive old tasks.
	CompletedAt time.Time
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusTerminated TaskStatus = "terminated"
)

func (t *TaskStatus) String() string {
	return string(*t)
}

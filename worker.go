package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

// Worker is a process that executes tasks. It is responsible for fetching task from the database,
// fetch only not locked task, execute the task, and then updates task status in the db.
type Worker struct {
	db        *pgxpool.Pool
	functions map[FunctionKey]RetryableFn
}

// NewWorker creates a new Worker instance.
func NewWorker(db *pgxpool.Pool) *Worker {
	return &Worker{
		db:        db,
		functions: make(map[FunctionKey]RetryableFn),
	}
}

// RegisterFunction registers a function with a key.
// You need to register a function that would handle the task.
// Functions could have dependencies, you need just provide the function that has signature func(context.Context, string) error.
// Methods of the struct with dependencies could be provided as a function.
func (w *Worker) RegisterFunction(key FunctionKey, function RetryableFn) {
	w.functions[key] = function
}

// ProcessTasks fetches tasks from the database and processes them.
// It loops forever and fetches tasks in a loop. Process one task at a time.
// I don't want to complicate the code with worker pools, i just want to keep it simple as possible.
func (w *Worker) ProcessTasks(ctx context.Context) {
	for {
		task, err := w.fetchTask(ctx)
		if err != nil {
			log.Println("Error fetching task:", err)
			continue
		}
		if task == nil {
			// Wait before trying to fetch another task
			// I don't wanna make the worker complicated.
			// You should use fibonacci backoff for this.
			time.Sleep(10 * time.Second)
			continue
		}

		err = w.ProcessTask(ctx, task)
		if err != nil {
			if err.Error() == "termination conditions met for task "+task.FunctionName {
				log.Printf("Terminating task %s after %d retries", task.FunctionName, task.RetryCount)
				if err := w.updateTaskStatus(ctx, task.ID, TaskStatusTerminated); err != nil {
					log.Printf("Error updating task status: %v", err)
				}
			} else {
				log.Printf("Function %s failed: %v", task.FunctionName, err)
			}
		}
	}
}

// fetchTask retrieves the next pending task from the database and locks it.
// `FOR UPDATE` would grant exclusive lock on the row.
func (w *Worker) fetchTask(ctx context.Context) (*Task, error) {
	query := `
		SELECT id, function_name, payload, status, retry_count, created_at, updated_at
		FROM tasks
		WHERE status = 'pending'
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	var task Task
	if err := w.db.QueryRow(ctx, query).Scan(
		&task.ID,
		&task.FunctionName,
		&task.Payload,
		&task.Status,
		&task.RetryCount,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// ProcessTask executes the task and updates the status in the database.
// You should use transaction for this. I'm too lazy to implement it.
func (w *Worker) ProcessTask(ctx context.Context, task *Task) error {
	terminator := NewTerminator(5, 10*time.Hour, task.CreatedAt)
	if terminator.ShouldTerminate(task.RetryCount) {
		return fmt.Errorf("termination conditions met for task %s", task.FunctionName)
	}

	function, exists := w.functions[FunctionKey(task.FunctionName)]
	if !exists {
		return fmt.Errorf("function %s not registered", task.FunctionName)
	}

	if err := function(ctx, task.Payload); err != nil {
		task.RetryCount++
		task.UpdatedAt = time.Now()

		query := `
			UPDATE tasks
			SET retry_count = $1, updated_at = $2
			WHERE id = $3
		`

		// reduced scope of err
		if _, err := w.db.Exec(ctx, query, task.RetryCount, task.UpdatedAt, task.ID); err != nil {
			return err
		}

		return err
	}

	return w.updateTaskStatus(ctx, task.ID, TaskStatusCompleted)
}

// updateTaskStatus updates the status of a task in the database.
func (w *Worker) updateTaskStatus(ctx context.Context, taskID int64, status TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := w.db.Exec(ctx, query, status, time.Now(), taskID)
	return err
}

package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type StatefulRetrier struct {
	db         *pgxpool.Pool
	terminator *Terminator
	functions  map[FunctionKey]func(context.Context, string) error
}

func NewStatefulRetrier(db *pgxpool.Pool, terminator *Terminator) *StatefulRetrier {
	return &StatefulRetrier{
		db:         db,
		terminator: terminator,
		functions:  make(map[FunctionKey]func(context.Context, string) error),
	}
}

func (r *StatefulRetrier) ScheduleTask(ctx context.Context, key FunctionKey, payload interface{}) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := Task{
		FunctionName: string(key),
		Payload:      string(payloadData),
		Status:       TaskStatusPending,
		RetryCount:   0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO tasks (function_name, payload, status, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := r.db.Exec(
		ctx,
		query,
		task.FunctionName,
		task.Payload,
		task.Status,
		task.RetryCount,
		task.CreatedAt,
		task.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

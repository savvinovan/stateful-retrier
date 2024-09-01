package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create a new Worker instance.
	worker := NewWorker(db)

	// Register a function with a key.
	worker.RegisterFunction(MyFunctionKey, MyFunction)

	logger, _ := zap.NewDevelopment()
	funcWithDeps := NewMyFunctionWithDependencies(logger)
	worker.RegisterFunction(MyFunctionWithDependenciesKey, funcWithDeps.Execute)
}

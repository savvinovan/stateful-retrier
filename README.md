# Stateful retrier

## Purpose

This is a simple implementation of a stateful retrier. The stateful retrier is like durable functions. It retries the operation until it succeeds. It uses Postgres to store the state of the operation.

## Prerequisites

- Docker
- Golang 1.21

## How to run

```bash
docker-compose up

go run .
```

## Usage

```go

func main() {
    db, err := pgxpool.New("postgres", "dsn")
    if err != nil {
        log.Fatal(err)
    }

    // Setup terminator, it would terminate the task after 5 retries or 24 hours
    terminator := NewTerminator(5, 24 * time.Hour)

    retrier := NewStatefulRetrier(db, terminator)

    // Init worker
    worker := NewWorker(db)
	
    // Register function
    worker.RegisterFunction(MyFunctionKey, myFunction.Execute)

    // Start worker
    retrier.ScheduleTask(MyFunctionKey, myPayload)

    for {
        worker.ProcessTasks()
        time.Sleep(1 * time.Minute)
    }
}


```
package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

type FunctionKey string

type RetryableFn func(ctx context.Context, payload string) error

const (
	// MyFunctionKey is a key for MyFunction.
	MyFunctionKey                 FunctionKey = "MyFunction"
	MyFunctionWithDependenciesKey FunctionKey = "MyFunctionWithDependencies"
)

// MyFunction is a function that does something.
func MyFunction(_ context.Context, payload string) error {
	fmt.Println("MyFunction is running", payload)
	return nil
}

// check that MyFunction implements RetryableFn
var _ RetryableFn = MyFunction

func NewMyFunctionWithDependencies(l *zap.Logger) *MyFunctionWithDependencies {
	return &MyFunctionWithDependencies{
		log: l,
	}
}

// MyFunctionWithDependencies is a function that does something with dependencies.
type MyFunctionWithDependencies struct {
	log *zap.Logger
}

func (f *MyFunctionWithDependencies) Execute(ctx context.Context, payload string) error {
	f.log.Info("MyFunctionWithDependencies is executed", zap.String("payload", payload))
	return nil
}

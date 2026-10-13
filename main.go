package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	store := NewJobStore()

	_, err := store.Create(Job{
		Name:      "send-report",
		Interval:  2 * time.Second,
		NextRunAt: time.Now().Add(2 * time.Second),
		Status:    "ACTIVE",
	})

	if err != nil {
		panic(err)
	}
	workerPool := NewWorkerPool(3, store)
	workerPool.Start()
	scheduler := NewScheduler(store, workerPool)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		defer close(done)
		scheduler.Start(ctx)
	}()

	time.Sleep(10 * time.Second)

	cancel()

	<-done
	workerPool.Shutdown()

	fmt.Println("Main function finished")

}

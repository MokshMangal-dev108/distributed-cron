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
	scheduler := NewScheduler(store)
	ctx, cancel := context.WithCancel(context.Background())

	go scheduler.Start(ctx)

	time.Sleep(10 * time.Second)

	cancel()

	fmt.Println("Main function finished")

}

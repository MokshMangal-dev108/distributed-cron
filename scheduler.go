package main

import (
	"context"
	"fmt"
	"time"
)

type Scheduler struct {
	store *JobStore
}

func NewScheduler(store *JobStore) *Scheduler {
	return &Scheduler{
		store: store,
	}
}

func (sch *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler shutting down...")
			return
		default:
			sch.Execute()
		}
	}
}

func (sch *Scheduler) Execute() {
	jobs := sch.store.List()
	for _, job := range jobs {
		t := time.Now()
		if job.Status == "ACTIVE" && !job.NextRunAt.After(t) {
			fmt.Println("Executed ->", job.ID, job.Name, t)
			job.NextRunAt = t.Add(job.Interval)
		}
	}
}

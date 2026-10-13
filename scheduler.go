package main

import (
	"context"
	"fmt"
	"time"
)

type Scheduler struct {
	store      *JobStore
	workerPool *WorkerPool
}

func NewScheduler(store *JobStore, workerPool *WorkerPool) *Scheduler {
	return &Scheduler{
		store:      store,
		workerPool: workerPool,
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
		case <-ticker.C:
			sch.Execute()
		}
	}
}

func (sch *Scheduler) Execute() {
	jobs := sch.store.List()
	t := time.Now()
	for _, job := range jobs {
		if job.Status == "ACTIVE" && !job.NextRunAt.After(t) && !job.IsRunning {
			if sch.store.TryClaim(job.ID) {
				fmt.Println("Job claimed:", job.Name)
				job.NextRunAt = t.Add(job.Interval)
				sch.workerPool.Submit(job)
			}

		}
	}
}

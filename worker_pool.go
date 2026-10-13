package main

import (
	"fmt"
	"sync"
	"time"
)

type WorkerPool struct {
	workerCount int
	jobs        chan *Job
	wg          sync.WaitGroup
	store       *JobStore
}

func NewWorkerPool(workerCount int, store *JobStore) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan *Job, workerCount),
		store:       store,
		//wg:          sync.WaitGroup{}, we don't need this go automatically initialize this with empty mode
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for job := range p.jobs {
				p.process(job, workerID)
			}
		}(i + 1)
	}
}

func (p *WorkerPool) process(job *Job, workerID int) {
	// Always release the job when processing finishes.
	defer p.store.Release(job.ID)

	fmt.Printf("Worker %d started: %s (ID: %d)\n", workerID, job.Name, job.ID)

	// Simulate actual job execution.
	time.Sleep(3 * time.Second)

	fmt.Printf("Worker %d completed: %s (ID: %d)\n", workerID, job.Name, job.ID)
}

func (p *WorkerPool) Submit(job *Job) {
	p.jobs <- job
}

func (p *WorkerPool) Shutdown() {
	close(p.jobs) // here we can't use defer, because otherwise it will schedule for later to close this channel
	// and wg.wait will never stop
	p.wg.Wait()

}

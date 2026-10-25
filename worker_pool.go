package main

import (
	"fmt"
	"sync"
	"time"
)

type JobTask struct {
	jobs      *Job
	execution *Execution
}

type WorkerPool struct {
	workerCount int
	jobs        chan JobTask
	wg          sync.WaitGroup
	store       *JobStore
	executions  *ExecutionStore
}

func NewWorkerPool(workerCount int, store *JobStore, executions *ExecutionStore) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan JobTask, workerCount),
		store:       store,
		executions:  executions,
		//wg:          sync.WaitGroup{}, we don't need this go automatically initialize this with empty mode
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for task := range p.jobs {
				p.process(task, workerID)
			}
		}(i + 1)
	}
}

func (p *WorkerPool) process(task JobTask, workerID int) {
	job := task.jobs
	execution := task.execution
	// Always release the job when processing finishes.
	defer p.store.Release(job.ID)

	p.executions.MarkRunning(job.ID)

	fmt.Printf("Worker %d started: %s | Execution ID: %d\n", workerID, job.Name, execution.ID)

	// Simulate actual job execution.
	time.Sleep(3 * time.Second)

	p.executions.MarkCompleted(job.ID)

	fmt.Printf("Worker %d completed: %s | Execution ID: %d\n", workerID, job.Name, execution.ID)
}

func (p *WorkerPool) Submit(job *Job, execution *Execution) {
	p.jobs <- JobTask{
		jobs:      job,
		execution: execution,
	}
}

func (p *WorkerPool) Shutdown() {
	close(p.jobs) // here we can't use defer, because otherwise it will schedule for later to close this channel
	// and wg.wait will never stop
	p.wg.Wait()

}

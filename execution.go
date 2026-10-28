package main

import (
	"sync"
	"time"
)

type ExecutionStatus string

const (
	ExecutionPending   ExecutionStatus = "PENDING"
	ExecutionRunning   ExecutionStatus = "RUNNING"
	ExecutionCompleted ExecutionStatus = "COMPLETED"
	ExecutionFailed    ExecutionStatus = "FAILED"
)

type Execution struct {
	ID          int
	JobID       int
	Status      ExecutionStatus
	ScheduledAt time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	Error       string
}

type ExecutionStore struct {
	mu      sync.RWMutex
	nextID  int
	records map[int]*Execution
}

func NewExecutionStore() *ExecutionStore {
	return &ExecutionStore{
		nextID:  1,
		records: make(map[int]*Execution),
	}
}

func (es *ExecutionStore) Create(JobId int, scheduledAt time.Time) *Execution {
	es.mu.Lock()
	defer es.mu.Unlock()

	execution := &Execution{
		ID:          es.nextID,
		JobID:       JobId,
		Status:      ExecutionPending,
		ScheduledAt: scheduledAt,
	}
	es.records[execution.ID] = execution
	es.nextID++
	return execution
}

func (es *ExecutionStore) MarkRunning(Id int) bool {
	es.mu.Lock()
	defer es.mu.Unlock()
	val, exists := es.records[Id]
	if !exists {
		return false
	}
	val.Status = ExecutionRunning
	now := time.Now()
	val.StartedAt = &now
	return true
}

func (es *ExecutionStore) MarkCompleted(Id int) bool {
	es.mu.Lock()
	defer es.mu.Unlock()
	val, exists := es.records[Id]
	if !exists {
		return false
	}
	val.Status = ExecutionCompleted
	now := time.Now()
	val.FinishedAt = &now
	return true
}

func (es *ExecutionStore) MarkFailed(id int, err error) bool {
	es.mu.Lock()
	defer es.mu.Unlock()

	execution, exists := es.records[id]
	if !exists {
		return false
	}
	execution.Status = ExecutionFailed
	now := time.Now()
	execution.FinishedAt = &now
	if err != nil {
		execution.Error = err.Error()
	}

	return true
}

func (es *ExecutionStore) Get(id int) (*Execution, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	execution, exists := es.records[id]
	if !exists {
		return nil, false
	}

	return execution, true
}

func (es *ExecutionStore) List() []*Execution {
	es.mu.RLock()
	defer es.mu.RUnlock()

	executions := make([]*Execution, 0, len(es.records))

	for _, execution := range es.records {
		executions = append(executions, execution)
	}

	return executions
}

func (es *ExecutionStore) ListByJobID(jobID int) []*Execution {
	es.mu.RLock()
	defer es.mu.RUnlock()

	executions := make([]*Execution, 0)

	for _, execution := range es.records {
		if execution.JobID == jobID {
			executions = append(executions, execution)
		}
	}
	return executions
}

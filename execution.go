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

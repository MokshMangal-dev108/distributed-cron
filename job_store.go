package main

import (
	"errors"
	"strings"
	"sync"
)

type JobStore struct {
	nextID int
	store  map[int]*Job
	mu     sync.Mutex
}

func NewJobStore() *JobStore {
	return &JobStore{
		nextID: 0,
		store:  make(map[int]*Job),
	}
}

func (st *JobStore) Create(job Job) (*Job, error) {
	if strings.TrimSpace(job.Name) == "" {
		return nil, errors.New("job name is required")
	}
	if job.Interval <= 0 {
		return nil, errors.New("job interval is required")
	}
	st.nextID++
	job.ID = st.nextID
	st.store[job.ID] = &job
	return &job, nil
}

func (st *JobStore) Get(id int) (*Job, bool) {
	job, ok := st.store[id]
	return job, ok
}

func (st *JobStore) List() []*Job {
	lst := make([]*Job, 0, len(st.store))
	for _, job := range st.store {
		lst = append(lst, job)
	}
	return lst
}

func (st *JobStore) Delete(id int) bool {
	_, exists := st.store[id]
	if exists {
		delete(st.store, id)
	}
	return exists
}

func (st *JobStore) TryClaim(id int) bool {
	st.mu.Lock()
	defer st.mu.Unlock()
	job, exists := st.store[id]
	if !exists || job.IsRunning {
		return false
	}
	job.IsRunning = true
	return true
}

func (st *JobStore) Release(id int) {
	st.mu.Lock()
	defer st.mu.Unlock()
	job, exists := st.store[id]
	if exists {
		job.IsRunning = false
	}
}

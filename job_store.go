package main

import (
	"errors"
	"strings"
)

type JobStore struct {
	nextID int
	store  map[int]*Job
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

package main

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

func (st *JobStore) Create(job Job) *Job {
	st.nextID++
	job.ID = st.nextID
	st.store[job.ID] = &job
	return &job
}

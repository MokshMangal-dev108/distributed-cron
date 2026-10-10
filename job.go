package main

import "time"

type Job struct {
	ID        int
	Name      string
	Interval  time.Duration
	NextRunAt time.Time
	Status    string
}

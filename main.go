package main

import (
	"fmt"
	"time"
)

func main() {
	store := NewJobStore()

	job1 := store.Create(Job{
		Name:     "send-report",
		Interval: 5 * time.Minute,
		Status:   "ACTIVE",
	})

	job2 := store.Create(Job{
		Name:     "cleanup-files",
		Interval: 10 * time.Second,
		Status:   "ACTIVE",
	})

	fmt.Println("Created:", job1.ID, job1.Name)
	fmt.Println("Created:", job2.ID, job2.Name)
}

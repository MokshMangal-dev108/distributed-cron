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

	fmt.Println("LIST:", store.List())
	jb, _ := store.Get(1)
	fmt.Println("ID 1:", jb)
	del := store.Delete(1)
	fmt.Println("DEL:", del)
	fmt.Println("LIST:", store.List())
}

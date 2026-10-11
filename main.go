package main

import (
	"fmt"
	"time"
)

func main() {
	store := NewJobStore()

	job1, err := store.Create(Job{
		Name:     "send-report",
		Interval: 5 * time.Minute,
		Status:   "ACTIVE",
	})
	if err != nil {
		fmt.Println("Error-> ", err)
		return
	}

	job2, err := store.Create(Job{
		Name:     "",
		Interval: 10 * time.Second,
		Status:   "ACTIVE",
	})
	if err != nil {
		fmt.Println("Error-> ", err)
		return
	}
	fmt.Println("Created:", job1.ID, job1.Name)
	fmt.Println("Created:", job2.ID, job2.Name)

	fmt.Println("LIST:", store.List())
	jb, _ := store.Get(1)
	fmt.Println("ID 1:", jb)
	del := store.Delete(1)
	fmt.Println("DEL:", del)
	fmt.Println("LIST:", store.List())
}

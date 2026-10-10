package main

import (
	"fmt"
	"time"
)

func main() {
	job := Job{
		ID:        1,
		Name:      "send-report",
		Interval:  5 * time.Minute,
		NextRunAt: time.Now().Add(5 * time.Minute),
		Status:    "ACTIVE",
	}
	fmt.Println("Job Name:", job.Name)
	fmt.Println("Interval:", job.Interval)
	fmt.Println("Next Run At:", job.NextRunAt)
	fmt.Println("Status:", job.Status)
}

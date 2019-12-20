package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/favicon.ico", http.StripPrefix("/favicon.ico", fs))
	http.HandleFunc("/", handleJob)
	log.Println("Server started on port 3000")
	http.ListenAndServe(":3000", nil)
}

//Payload ...
type Payload struct {
	Text1 string
	Text2 string
	Text3 string
}

//Job ..
type Job struct {
	Payload Payload
}

func doRequest(job Job) int {
	// api call
	time.Sleep(1 * time.Second)
	return 1
}

func handleJob(w http.ResponseWriter, r *http.Request) {

	jobs := []Job{}
	for j := 1; j <= 10000; j++ {
		jobs = append(jobs, Job{
			Payload{
				Text1: fmt.Sprintf("%d-1-test", j),
				Text2: fmt.Sprintf("%d-2-test", j),
				Text3: fmt.Sprintf("%d-3-test", j),
			},
		})
	}

	pool := newWorkerPool(1000)
	result := pool.Run(jobs, func(job Job) int {
		return doRequest(job)
	})

	js, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type workerPool struct {
	Jobs    chan Job
	Results chan int
	Func    func(Job) int
}

func newWorkerPool(workerCount int) *workerPool {
	return &workerPool{
		Results: make(chan int, workerCount),
	}
}

func (wp *workerPool) Run(jobs []Job, fn func(Job) int) int {
	wp.Jobs = make(chan Job, len(jobs))
	wp.Func = fn

	for w := 1; w <= cap(wp.Results); w++ {
		go wp.worker()
	}

	for _, job := range jobs {
		wp.Jobs <- job
	}

	close(wp.Jobs)
	result := 0
	for range jobs {
		result += <-wp.Results
	}
	return result
}

func (wp *workerPool) worker() {
	for job := range wp.Jobs {
		wp.Results <- wp.Func(job)
	}
}

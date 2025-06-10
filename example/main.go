package main

import (
	"fmt"
	"workerpool/workerpool"
)

func main() {
	var number_of_greetings = 10

	hello_func := func(name string) error {
		fmt.Printf("Hello from job number %s!\n", name)
		return nil
	}
	pool := workerpool.New(10, hello_func)

	for range 3 {
		pool.AddWorker()
	}

	for i := range number_of_greetings {
		pool.AddJob(fmt.Sprint(i))
	}

	pool.StopWait()
}

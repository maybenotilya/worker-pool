package workerpool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TODO: add panic recover for handler purposes
type WorkerPool struct {
	jobs      chan string
	handler   func(string) error // Handler MUST NOT panic!!!
	timeout   time.Duration      // Delay to try to add new task in full queue
	stopChans map[int]chan bool
	mutex     sync.Mutex
	nextID    int
	wg        sync.WaitGroup
	jobs_wg   sync.WaitGroup
	isStopped atomic.Bool
}

func New(bufferSize int, timeout time.Duration, handler func(string) error) *WorkerPool {
	return &WorkerPool{
		jobs:      make(chan string, bufferSize),
		handler:   handler,
		timeout:   timeout,
		stopChans: make(map[int]chan bool),
	}
}

func (pool *WorkerPool) runWorker(id int, stopChan chan bool) {
	defer pool.wg.Done()
	for {
		select {
		case job, ok := <-pool.jobs:
			if !ok {
				return
			}

			err := pool.handler(job)
			if err != nil {
				fmt.Printf("[Worker %d] error: %v\n", id, err)
			} else {
				fmt.Printf("[Worker %d] processed job\n", id)
			}
			pool.jobs_wg.Done()

		case <-stopChan:
			return
		}
	}
}

func (pool *WorkerPool) AddWorker() (int, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.isStopped.Load() {
		return -1, errors.New("failed to add worker: worker pool is stopped")
	}

	stopChan := make(chan bool)
	id := pool.nextID
	pool.nextID++

	pool.stopChans[id] = stopChan

	pool.wg.Add(1)

	go pool.runWorker(id, stopChan)

	return id, nil
}

func (pool *WorkerPool) RemoveWorker(id int) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.isStopped.Load() {
		return errors.New("failed to remove worker: worker pool is stopped")
	}
	stopChan, ok := pool.stopChans[id]
	if !ok {
		return errors.New("failed to remove worker: worker not found")
	}

	close(stopChan)
	delete(pool.stopChans, id)

	return nil
}

func (pool *WorkerPool) AddJob(job string) error {
	if pool.isStopped.Load() {
		return errors.New("failed to add job: worker pool is stopped")
	}

	pool.jobs_wg.Add(1)
	select {
	case pool.jobs <- job:
		return nil
	case <-time.After(pool.timeout):
		pool.jobs_wg.Done()
		return errors.New("failed to add job: job queue is full")
	}
}

func (pool *WorkerPool) Stop() {
	pool.mutex.Lock()

	if pool.isStopped.Load() {
		pool.mutex.Unlock()
		return
	}

	pool.isStopped.Store(true)

	close(pool.jobs)

	for _, stopChan := range pool.stopChans {
		close(stopChan)
	}

	for id := range pool.stopChans {
		delete(pool.stopChans, id)
	}

	pool.mutex.Unlock()

	pool.wg.Wait()
}

// Waits till all jobs are done and stops after
func (pool *WorkerPool) StopWait() {
	pool.mutex.Lock()

	if pool.isStopped.Load() {
		pool.mutex.Unlock()
		return
	}

	pool.isStopped.Store(true)

	close(pool.jobs)

	pool.jobs_wg.Wait()

	for _, stopChan := range pool.stopChans {
		close(stopChan)
	}

	for id := range pool.stopChans {
		delete(pool.stopChans, id)
	}

	pool.mutex.Unlock()

	pool.wg.Wait()
}

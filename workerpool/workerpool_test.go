package workerpool

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"testing"
)

func TestAddRemove(t *testing.T) {
	pool := New(10, func(s string) error { return nil })
	id1, err1 := pool.AddWorker()
	id2, err2 := pool.AddWorker()

	if err1 != nil || err2 != nil {
		t.Fatalf("AddWorker failed")
	}
	if id1 != 0 && id2 != 1 {
		t.Errorf("Expected id1 = 0, id2 = 1, got id1 = %d, id2 = %d", id1, id2)
	}

	err2 = pool.RemoveWorker(id2)
	err1 = pool.RemoveWorker(id1)

	if err1 != nil || err2 != nil {
		t.Error("RemoveWorker failed")
	}

	pool.Stop()
}

func TestJobs(t *testing.T) {
	var counter atomic.Int32
	num_workers := 10
	num_runs := 100

	pool := New(5, func(s string) error {
		counter.Add(1)
		return nil
	})

	for range num_workers {
		_, err := pool.AddWorker()
		if err != nil {
			pool.Stop()
			t.Fatal(err)
		}
	}

	for range num_runs {
		err := pool.AddJob("Job")
		if err != nil {
			pool.Stop()
			t.Fatal(err)
		}
	}

	pool.StopWait()

	if counter.Load() != int32(num_runs) {
		t.Fatalf("Processed %d jobs, expected %d", counter.Load(), num_runs)
	}
}

func TestErrorsHandle(t *testing.T) {
	var counter atomic.Int32
	num_workers := 10
	num_errors := 50
	num_runs := num_errors * 2

	pool := New(10, func(s string) error {
		if s == "error" {
			counter.Add(1)
			return errors.New("error")
		}
		return nil
	})

	for range num_workers {
		_, err := pool.AddWorker()
		if err != nil {
			pool.Stop()
			t.Fatal(err)
		}
	}

	for i := range num_runs {
		var err error
		if i%2 == 0 {
			err = pool.AddJob("error")
		} else {
			err = pool.AddJob("not error")
		}

		if err != nil {
			pool.Stop()
			t.Fatal(err)
		}
	}

	pool.StopWait()

	if counter.Load() != int32(num_errors) {
		t.Fatalf("Processed %d errors, expected %d", counter.Load(), num_errors)
	}

}

func TestRemoveNonExistential(t *testing.T) {
	pool := New(10, func(s string) error { return nil })
	_, err := pool.AddWorker()
	if err != nil {
		pool.Stop()
		t.Fatal(err)
	}

	err = pool.RemoveWorker(999)
	if err == nil || err.Error() != "failed to remove worker: worker not found" {
		t.Errorf("Expected 'worker not found' error, got: %v", err)
	}
	pool.Stop()
}

func TestRemoveAfterStop(t *testing.T) {
	pool := New(10, func(s string) error { return nil })
	id, err := pool.AddWorker()
	if err != nil {
		pool.Stop()
		t.Fatal(err)
	}

	pool.Stop()
	err = pool.RemoveWorker(id)
	if err == nil || err.Error() != "failed to remove worker: worker pool is stopped" {
		t.Errorf("Expected 'worker pool is stopped' error, got: %v", err)
	}
	pool.Stop()
}

func TestAddAfterStop(t *testing.T) {
	pool := New(10, func(s string) error { return nil })
	pool.Stop()

	_, err := pool.AddWorker()
	if err == nil || err.Error() != "failed to add worker: worker pool is stopped" {
		t.Fatalf("Expected 'worker pool is stopped' error, got: %v", err)
	}
}

func TestJobAfterStop(t *testing.T) {
	pool := New(10, func(s string) error { return nil })
	pool.Stop()

	err := pool.AddJob("")

	if err == nil || err.Error() != "failed to add job: worker pool is stopped" {
		t.Fatalf("Expected 'worker pool is stopped' error, got: %v", err)
	}
}

func TestAddJobWithoutWorkers(t *testing.T) {
	pool := New(10, func(s string) error {
		t.Error("Handler shouldn't be called without workers")
		return nil
	})

	for range 3 {
		err := pool.AddJob("Job")
		if err != nil {
			pool.Stop()
			t.Fatal(err)
		}
	}

	pool.Stop()
}

func TestMultipleStop(t *testing.T) {
	pool := New(10, func(s string) error { return nil })

	for range 7 {
		pool.Stop()
	}
}

func TestRandomOperations(t *testing.T) {
	num_workers := 10
	num_runs := 100
	SEED := 42
	rnd := rand.New(rand.NewSource(int64(SEED)))

	available_ids := []int{}

	pool := New(3, func(s string) error { return nil })

	for range num_workers {
		id, err := pool.AddWorker()
		if err != nil {
			t.Error(err)
		}
		available_ids = append(available_ids, id)
		t.Logf("Worker %d added", id)
	}

	for range num_runs {
		x := rnd.Int()
		if x%3 == 0 {
			if len(available_ids) == 0 {
				continue
			}
			err := pool.AddJob("Job")
			if err != nil {
				t.Error(err)
			}

			t.Log("Job processed")
		} else if x%3 == 1 {
			id, err := pool.AddWorker()
			if err != nil {
				t.Error(err)
			}
			available_ids = append(available_ids, id)

			t.Logf("Worker %d added", id)
		} else {
			if len(available_ids) == 0 {
				continue
			}

			id := available_ids[0]
			available_ids = available_ids[1:]
			err := pool.RemoveWorker(id)
			if err != nil {
				t.Error(err)
			}
			t.Logf("Worker %d removed", id)
		}
	}

	pool.Stop()
}

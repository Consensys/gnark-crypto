package utils

import (
	"runtime"
	"sync"
)

// Would normally put this in internal/parallel; but it may be desirable for it to be accessible from gnark

type Task func(start, end int)

type job struct {
	start, end int
	task       Task
	done       *sync.WaitGroup
}

func worker(jobs <-chan job) {
	for j := range jobs {
		j.task(j.start, j.end)
		j.done.Done()
	}
}

/*type Work struct {
	NbIterations int
	Task Task
}*/

type WorkerPool struct {
	jobs      chan job
	nbWorkers int
}

func NewWorkerPool() (p WorkerPool) {
	p.nbWorkers = runtime.NumCPU()
	p.jobs = make(chan job, 8*p.nbWorkers)
	for i := 0; i < p.nbWorkers; i++ {
		go worker(p.jobs)
	}
	return
}

// Dispatch schedules the execution of independent tasks of equal length and difficulty
// the preference is to run each task on a single worker
func (p *WorkerPool) Dispatch(nbIterations int, minJobSize int, tasks ...Task) *sync.WaitGroup {

	nbAvailableWorkers := p.nbWorkers - len(p.jobs) // TODO Try setting nbAvailableWorkers to p.nbWorkers and see if that's better
	var done sync.WaitGroup
	for len(tasks) >= nbAvailableWorkers { // spread them evenly. INCORRECTLY assumes the currently outstanding tasks take the same amount of time
		done.Add(nbAvailableWorkers)
		for workerI := 0; workerI < nbAvailableWorkers; workerI++ {
			p.jobs <- job{
				start: 0,
				end:   nbIterations,
				task:  tasks[workerI],
				done:  &done,
			}
		}
		tasks = tasks[nbAvailableWorkers:]
		nbAvailableWorkers = p.nbWorkers
	}

	// the remainders get broken up
	nbRemainingIterations := nbIterations * len(tasks)
	jobLength := Max(minJobSize, // TODO: Experiment with other minimum job size enforcement methods
		int(DivCeiling(uint(nbRemainingIterations), uint(nbAvailableWorkers))),
	)
	firstTaskStart := 0
	for nbRemainingIterations > 0 {
		firstTaskEnd := Min(nbIterations, firstTaskStart+jobLength)
		done.Add(1)
		p.jobs <- job{
			start: firstTaskStart,
			end:   firstTaskEnd,
			task:  tasks[0],
			done:  &done,
		}
		nbRemainingIterations += firstTaskStart - firstTaskEnd
		if firstTaskEnd == nbIterations { // if we've exhausted the current task
			tasks = tasks[1:]
			firstTaskStart = 0
		}
	}

	return &done
}

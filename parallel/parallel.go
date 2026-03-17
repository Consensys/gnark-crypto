// Package parallel provides shared parallel execution primitives used
// throughout gnark-crypto, including generated code.
package parallel

import (
	"runtime"
	"sync"
)

// Execute process in parallel the work function
func Execute(nbIterations int, work func(int, int), maxCpus ...int) {

	nbTasks := runtime.NumCPU()
	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
		if nbTasks < 1 {
			nbTasks = 1
		} else if nbTasks > 512 {
			nbTasks = 512
		}
	}

	if nbTasks == 1 {
		// no go routines
		work(0, nbIterations)
		return
	}

	nbIterationsPerCpus := nbIterations / nbTasks

	// more CPUs than tasks: a CPU will work on exactly one iteration
	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
		nbTasks = nbIterations
	}

	var wg sync.WaitGroup

	extraTasks := nbIterations - (nbTasks * nbIterationsPerCpus)
	extraTasksOffset := 0

	for i := 0; i < nbTasks; i++ {
		_start := i*nbIterationsPerCpus + extraTasksOffset
		_end := _start + nbIterationsPerCpus
		if extraTasks > 0 {
			_end++
			extraTasks--
			extraTasksOffset++
		}
		wg.Go(func() {
			work(_start, _end)
		})
	}

	wg.Wait()
}

// ExecuteAligned is like Execute but keeps chunk boundaries aligned to the
// provided alignment, except for a possible tail on the last chunk.
func ExecuteAligned(nbIterations, alignment int, work func(int, int), maxCpus ...int) {
	nbTasks := runtime.NumCPU()
	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
		if nbTasks < 1 {
			nbTasks = 1
		} else if nbTasks > 512 {
			nbTasks = 512
		}
	}

	if nbTasks == 1 || nbIterations <= alignment {
		work(0, nbIterations)
		return
	}

	totalUnits := nbIterations / alignment
	leftover := nbIterations % alignment

	if nbTasks > totalUnits {
		nbTasks = totalUnits
	}
	if nbTasks <= 1 {
		work(0, nbIterations)
		return
	}

	unitsPerTask := totalUnits / nbTasks
	extraUnits := totalUnits % nbTasks

	var wg sync.WaitGroup
	start := 0
	for i := 0; i < nbTasks; i++ {
		units := unitsPerTask
		if i < extraUnits {
			units++
		}
		end := start + units*alignment
		if i == nbTasks-1 {
			end += leftover
		}
		_start, _end := start, end
		start = end
		wg.Go(func() {
			work(_start, _end)
		})
	}
	wg.Wait()
}

// Chunks returns a slice of [2]int where each element is a (start, end) range to be processed by a worker,
// exactly as Execute does.
func Chunks(nbIterations int, maxCpus ...int) [][2]int {

	var chunks [][2]int

	nbTasks := runtime.NumCPU()
	if len(maxCpus) == 1 {
		nbTasks = maxCpus[0]
		if nbTasks < 1 {
			nbTasks = 1
		} else if nbTasks > 512 {
			nbTasks = 512
		}
	}

	if nbTasks == 1 {
		// no go routines
		chunks = append(chunks, [2]int{0, nbIterations})
		return chunks
	}

	nbIterationsPerCpus := nbIterations / nbTasks

	// more CPUs than tasks: a CPU will work on exactly one iteration
	if nbIterationsPerCpus < 1 {
		nbIterationsPerCpus = 1
		nbTasks = nbIterations
	}

	extraTasks := nbIterations - (nbTasks * nbIterationsPerCpus)
	extraTasksOffset := 0

	for i := 0; i < nbTasks; i++ {
		_start := i*nbIterationsPerCpus + extraTasksOffset
		_end := _start + nbIterationsPerCpus
		if extraTasks > 0 {
			_end++
			extraTasks--
			extraTasksOffset++
		}
		chunks = append(chunks, [2]int{_start, _end})
	}

	return chunks
}

// Task represents a function that processes a range [start, end).
type Task func(start, end int)

type job struct {
	start, end int
	task       Task
	done       *sync.WaitGroup
}

// WorkerPool is a persistent pool of goroutines that process submitted work.
type WorkerPool struct {
	chJobs    chan job
	nbWorkers int
}

// NewWorkerPool creates a new WorkerPool with NumCPU+2 workers.
func NewWorkerPool() *WorkerPool {
	p := &WorkerPool{}
	p.nbWorkers = runtime.NumCPU() + 2
	p.chJobs = make(chan job, 40*p.nbWorkers)
	for i := 0; i < p.nbWorkers; i++ {
		go func() {
			for j := range p.chJobs {
				j.task(j.start, j.end)
				j.done.Done()
			}
		}()
	}
	return p
}

// NbWorkers returns the number of workers in the pool.
func (wp *WorkerPool) NbWorkers() int {
	return wp.nbWorkers
}

// Stop closes the job channel and frees the workers. It does not wait for
// in-flight jobs to complete.
func (wp *WorkerPool) Stop() {
	close(wp.chJobs)
}

// Submit distributes n iterations of work into chunks of minBlock size and
// returns a WaitGroup that completes when all chunks are done.
func (wp *WorkerPool) Submit(n int, work func(int, int), minBlock int) *sync.WaitGroup {
	var wg sync.WaitGroup

	for start := 0; start < n; start += minBlock {
		start := start
		end := min(start+minBlock, n)
		wg.Add(1)
		wp.chJobs <- job{
			task:  work,
			start: start,
			end:   end,
			done:  &wg,
		}
	}

	return &wg
}

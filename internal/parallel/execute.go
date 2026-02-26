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

// ExecuteAligned is like Execute but ensures that chunk boundaries are aligned to
// the given alignment value (except possibly the last chunk).
// This is useful when work functions use SIMD operations that process elements
// in fixed-size blocks (e.g., 16 for AVX512), to avoid per-chunk tail handling.
//
// Work is distributed evenly: tasks receive either k or k+alignment elements,
// where k is the largest multiple of alignment that fits. Any unaligned tail
// (nbIterations % alignment) is absorbed by the last task.
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

	// Distribute aligned units across tasks evenly.
	totalUnits := nbIterations / alignment
	leftover := nbIterations % alignment // unaligned tail (usually 0 for FFT)

	// Don't spawn more tasks than aligned units.
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

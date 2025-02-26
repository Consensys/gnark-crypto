// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package polynomial

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"runtime"
	"sort"
	"sync"
	"unsafe"
)

// Memory management for polynomials
// Thread-safe implementation of polynomial memory pool

type sizedPool struct {
	maxN int
	pool sync.Pool
}

type inUseData struct {
	allocatedFor []uintptr
	pool         *sizedPool
}

type Pool struct {
	inUse    sync.Map
	subPools []sizedPool
}

func (p *sizedPool) get(n int) *fr.Element {
	return p.pool.Get().(*fr.Element)
}

func (p *sizedPool) put(ptr *fr.Element) {
	p.pool.Put(ptr)
}

func NewPool(maxN ...int) (pool Pool) {
	sort.Ints(maxN)
	pool = Pool{
		subPools: make([]sizedPool, len(maxN)),
	}

	for i := range pool.subPools {
		subPool := &pool.subPools[i]
		subPool.maxN = maxN[i]
		subPool.pool = sync.Pool{
			New: func() interface{} {
				return getDataPointer(make([]fr.Element, 0, subPool.maxN))
			},
		}
	}
	return
}

func (p *Pool) findCorrespondingPool(n int) *sizedPool {
	poolI := 0
	for poolI < len(p.subPools) && n > p.subPools[poolI].maxN {
		poolI++
	}
	return &p.subPools[poolI] // out of bounds error here would mean that n is too large
}

func (p *Pool) Make(n int) []fr.Element {
	pool := p.findCorrespondingPool(n)
	ptr := pool.get(n)
	p.addInUse(ptr, pool)
	return unsafe.Slice(ptr, n)
}

// Dump dumps a set of polynomials into the pool
func (p *Pool) Dump(slices ...[]fr.Element) {
	for _, slice := range slices {
		ptr := getDataPointer(slice)
		if metadata, ok := p.inUse.LoadAndDelete(ptr); ok {
			metadata.(inUseData).pool.put(ptr)
		} else {
			panic("attempting to dump a slice not created by the pool")
		}
	}
}

func (p *Pool) addInUse(ptr *fr.Element, pool *sizedPool) {
	pcs := make([]uintptr, 2)
	n := runtime.Callers(3, pcs)

	// Use LoadOrStore to atomically check and store
	if actual, loaded := p.inUse.LoadOrStore(ptr, inUseData{
		allocatedFor: pcs[:n],
		pool:         pool,
	}); loaded {
		panic(fmt.Errorf("re-allocated non-dumped slice, previously allocated at %v", runtime.CallersFrames(actual.(inUseData).allocatedFor)))
	}
}

func printFrame(frame runtime.Frame) {
	fmt.Printf("\t%s line %d, function %s\n", frame.File, frame.Line, frame.Function)
}

func (p *Pool) printInUse() {
	fmt.Println("slices never dumped allocated at:")
	p.inUse.Range(func(_, pcs any) bool {
		fmt.Println("-------------------------")

		var frame runtime.Frame
		frames := runtime.CallersFrames(pcs.(inUseData).allocatedFor)
		more := true
		for more {
			frame, more = frames.Next()
			printFrame(frame)
		}
		return true
	})
}

func getDataPointer(slice []fr.Element) *fr.Element {
	return (*fr.Element)(unsafe.SliceData(slice))
}

func (p *Pool) Clone(slice []fr.Element) []fr.Element {
	res := p.Make(len(slice))
	copy(res, slice)
	return res
}

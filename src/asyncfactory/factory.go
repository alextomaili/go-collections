package asyncfactory

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type (
	Allocator func(bool) unsafe.Pointer

	AsyncFactory struct {
		spinCount           int
		startAllocThreshold uint64
		bufferSize          uint64
		buffer              []unsafe.Pointer
		allocator           Allocator
		head, tail          uint64

		mutex *sync.Mutex
		cond  *sync.Cond

		d1 int
	}
)

func NewAsyncFactory(bufferSize, spinCount int, allocator Allocator) *AsyncFactory {
	f := AsyncFactory{
		startAllocThreshold: uint64(2 * (bufferSize / 3)),
		bufferSize:          uint64(bufferSize),
		spinCount:           spinCount,
		buffer:              make([]unsafe.Pointer, bufferSize, bufferSize),
		allocator:           allocator,
		head:                0,
		tail:                0,
	}

	f.mutex = &sync.Mutex{}
	f.cond = sync.NewCond(f.mutex)

	for i := uint64(0); i < f.bufferSize; i++ {
		f.buffer[i] = f.allocator(true)
	}
	f.tail = f.bufferSize

	go func(f *AsyncFactory) {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		for {
			f.wait()
			f.allocNext(true)
		}
	}(&f)

	return &f
}

func (f *AsyncFactory) idx(i uint64) uint64 {
	return i % f.bufferSize
}

func (f *AsyncFactory) force() {
	f.cond.Signal()
}

func (f *AsyncFactory) wait() {
	f.mutex.Lock()
	f.cond.Wait()
	f.mutex.Unlock()
}

func (f *AsyncFactory) allocNext(async bool) {
	for true {
		t := atomic.LoadUint64(&f.tail)
		h := atomic.LoadUint64(&f.head)

		if t-h == f.bufferSize {
			f.d1++
			return
		}

		for t-h < f.bufferSize {
			f.buffer[f.idx(t)] = f.allocator(async)
			t = atomic.AddUint64(&f.tail, 1)
			h = atomic.LoadUint64(&f.head)
		}

		runtime.Gosched()
	}

}

func (f *AsyncFactory) Alloc() unsafe.Pointer {
	c := 0
	for c < f.spinCount {

		h := atomic.LoadUint64(&f.head)
		t := atomic.LoadUint64(&f.tail)
		r := f.buffer[f.idx(h)]

		if h < t {
			if atomic.CompareAndSwapUint64(&f.head, h, h+1) {
				return r
			}
		}

		if c == 0 {
			c++
			f.force() //force allocation thread
			runtime.Gosched()
		} else if c%5 == 0 {
			c++
			runtime.Gosched()
		} else {
			c++
		}
	}
	return f.allocator(false)
}

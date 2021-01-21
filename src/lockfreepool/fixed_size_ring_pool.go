package lockfreepool

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

const (
	versionMask int64 = 0xFFFFFFFF
	lockedMask  int64 = 0x0100000000
	hasDataMask int64 = 0x0200000000
)

type (
	slot struct {
		typ   unsafe.Pointer
		data  unsafe.Pointer
		state int64
	}

	FixedSizeRingPool struct {
		size    int64
		buffer  []slot
		produce int64
		consume int64
	}
)

func NewFixedSizeRingPool(size int) *FixedSizeRingPool {
	p := &FixedSizeRingPool{
		size:    int64(size),
		buffer:  make([]slot, size, size),
		produce: -1,
		consume: -1,
	}
	return p
}

func (f *FixedSizeRingPool) idx(i int64) int64 {
	return i % f.size
}

func (f *FixedSizeRingPool) Put(v interface{}) bool {
	var ptr, attempts, prevState, newState, idx int64

	attempts = 0
	for true {
		if attempts >= f.size {
			return false
		}
		attempts++

		ptr = atomic.LoadInt64(&f.produce)
		if !atomic.CompareAndSwapInt64(&f.produce, ptr, ptr+1) {
			continue
		}
		ptr = ptr + 1

		idx = f.idx(ptr)
		prevState = atomic.LoadInt64(&f.buffer[idx].state)

		if prevState&hasDataMask > 0 || prevState&lockedMask > 0 {
			continue
		}

		newState = ((prevState + 1) & versionMask) | lockedMask
		if !atomic.CompareAndSwapInt64(&f.buffer[idx].state, prevState, newState) {
			continue
		}
		prevState = newState

		vp := (*slot)(unsafe.Pointer(&v))
		f.buffer[idx].typ = vp.typ
		f.buffer[idx].data = vp.data

		newState = ((prevState + 1) & versionMask) | hasDataMask & ^lockedMask
		if !atomic.CompareAndSwapInt64(&f.buffer[idx].state, prevState, newState) {
			panic("uups, this idx must be owned by me")
		}

		break
	}
	return true
}

func (f *FixedSizeRingPool) Get() (v interface{}) {
	var ptr, prodPtr, attempts, prevState, newState, idx int64

	attempts = 0
	for true {
		if attempts >= f.size {
			return
		}
		attempts++

		prodPtr = atomic.LoadInt64(&f.produce)
		ptr = atomic.LoadInt64(&f.consume)

		if ptr >= prodPtr {
			return
		}

		if !atomic.CompareAndSwapInt64(&f.consume, ptr, ptr+1) {
			continue
		}
		ptr = ptr + 1

		idx = f.idx(ptr)
		prevState = atomic.LoadInt64(&f.buffer[idx].state)

		if prevState&hasDataMask == 0 || prevState&lockedMask > 0 {
			continue
		}

		newState = ((prevState + 1) & versionMask) | lockedMask
		if !atomic.CompareAndSwapInt64(&f.buffer[idx].state, prevState, newState) {
			continue
		}
		prevState = newState

		vp := (*slot)(unsafe.Pointer(&v))
		vp.typ = f.buffer[idx].typ
		f.buffer[idx].typ = nil
		vp.data = f.buffer[idx].data
		f.buffer[idx].data = nil

		newState = ((prevState + 1) & versionMask) & ^hasDataMask & ^lockedMask
		if !atomic.CompareAndSwapInt64(&f.buffer[idx].state, prevState, newState) {
			panic("uups, this idx must be owned by me")
		}

		break
	}
	return
}

func (f *FixedSizeRingPool) State() string {
	c := atomic.LoadInt64(&f.consume)
	p := atomic.LoadInt64(&f.produce)
	return fmt.Sprintf("FixedSizeRingPool: { c: %v, p: %v, idx-c: %v, idx-p: %v}",
		c, p, f.idx(c), f.idx(p))
}

package lockfreepool

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type (
	ifaceWords struct {
		typ  unsafe.Pointer
		data unsafe.Pointer
	}

	FixedSizeRingPool struct {
		size   int64
		buffer []ifaceWords

		//produceA int64
		produce int64
		consume int64
	}
)

func NewFixedSizeRingPool(size int) *FixedSizeRingPool {
	p := &FixedSizeRingPool{
		size:    int64(size),
		buffer:  make([]ifaceWords, size, size),
		produce: 0,
		consume: 0,
	}
	return p
}

func (f *FixedSizeRingPool) idx(i int64) int64 {
	return i % f.size
}

func (f *FixedSizeRingPool) Put(v interface{}) bool {
	var c, p int64

	c = atomic.LoadInt64(&f.consume)
	p = atomic.LoadInt64(&f.produce)

	if p < c {
		panic("FixedSizeRingPool.Put() panic: p < c")
	} else if p-c > f.size {
		panic("FixedSizeRingPool.Put() panic: p-c > f.size")
	} else if p-c == f.size {
		return false
	}

	idx := f.idx(p)
	vp := (*ifaceWords)(unsafe.Pointer(&v))
	atomic.StorePointer(&f.buffer[idx].typ, vp.typ)
	atomic.StorePointer(&f.buffer[idx].data, vp.data)
	if vp.typ == nil || vp.data == nil {
		panic(fmt.Sprintf("FixedSizeRingPool.Put() panic: vp.typ == nil || vp.data == nil { c: %v, p: %v, vp.typ: %v, vp.data: %v}",
			c, p, vp.typ, vp.data))
	}

	if atomic.CompareAndSwapInt64(&f.produce, p, p+1) {
		return true
	}
	return false
}

func (f *FixedSizeRingPool) Get() (v interface{}) {
	var c, p int64

	for true {
		c = atomic.LoadInt64(&f.consume)
		p = atomic.LoadInt64(&f.produce)

		if c == p {
			return nil
		} else if c > p {
			panic(fmt.Sprintf("FixedSizeRingPool.Get() panic: c > p { c: %v, p: %v }", c, p))
		}

		if atomic.CompareAndSwapInt64(&f.consume, c, c+1) {
			break
		}
	}

	idx := f.idx(c)
	vp := (*ifaceWords)(unsafe.Pointer(&v))
	vp.typ = atomic.LoadPointer(&f.buffer[idx].typ)
	vp.data = atomic.LoadPointer(&f.buffer[idx].data)
	atomic.StorePointer(&f.buffer[idx].typ, nil)
	atomic.StorePointer(&f.buffer[idx].data, nil)
	if vp.typ == nil || vp.data == nil {
		panic(fmt.Sprintf("FixedSizeRingPool.Get() panic: vp.typ == nil || vp.data == nil { c: %v, p: %v, vp.typ: %v, vp.data: %v}",
			c, p, vp.typ, vp.data))
	}

	return
}

func (f *FixedSizeRingPool) State() string {
	c := atomic.LoadInt64(&f.consume)
	p := atomic.LoadInt64(&f.produce)
	return fmt.Sprintf("FixedSizeRingPool: { c: %v, p: %v, idx-c: %v, idx-p: %v}",
		c, p, f.idx(c), f.idx(p))
}

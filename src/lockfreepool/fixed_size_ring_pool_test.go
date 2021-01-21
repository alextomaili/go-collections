package lockfreepool

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type (
	tstS struct {
		c int64
	}

	tstPool interface {
		Get() interface{}
		Put(v interface{}) bool
		State() string
	}

	stdPool struct {
		sync.Pool
	}

	syncPool struct {
		lock   sync.Mutex
		buffer []interface{}
		top    int
	}
)

func (t *tstS) use(p tstPool) {
	if !atomic.CompareAndSwapInt64(&t.c, 0, 1) {
		panic("tstS.use() panic: !atomic.CompareAndSwapInt64(&t.c, 0, 1), " + p.State())
	}
}

func (t *tstS) clear(p tstPool) {
	if !atomic.CompareAndSwapInt64(&t.c, 1, 0) {
		panic("tstS.clear() panic: !atomic.CompareAndSwapInt64(&t.c, 1, 0), " + p.State())
	}
}

func (p *stdPool) Put(v interface{}) bool {
	p.Pool.Put(v)
	return true
}

func (p *stdPool) Get() interface{} {
	return p.Pool.Get()
}

func (p *stdPool) State() string {
	return "std-pool"
}

func newSyncPool(size int) *syncPool {
	return &syncPool{
		lock:   sync.Mutex{},
		buffer: make([]interface{}, size, size),
		top:    0,
	}
}

func (p *syncPool) Put(v interface{}) bool {
	r := false
	p.lock.Lock()
	if p.top < len(p.buffer)-1 {
		p.buffer[p.top] = v
		p.top++
		r = true
	}
	p.lock.Unlock()
	return r
}

func (p *syncPool) Get() (v interface{}) {
	p.lock.Lock()
	i := p.top - 1
	if i >= 0 {
		v = p.buffer[i]
		p.top--
	}
	p.lock.Unlock()
	return
}

func (p *syncPool) State() string {
	return "sync-pool"
}

func test(p tstPool, testThreadCount int, testIterationCount int) string {
	var (
		createCount int64
		getCount    int64
		putCount    int64
	)

	do := make(chan bool)
	wg := sync.WaitGroup{}

	for i := 0; i < testThreadCount; i++ {
		wg.Add(1)
		go func() {
			<-do
			for j := 0; j < testIterationCount; j++ {
				i := p.Get()
				if i == nil {
					v := &tstS{}
					atomic.AddInt64(&createCount, 1)

					if p.Put(v) {
						atomic.AddInt64(&putCount, 1)
					}
				} else {
					v := i.(*tstS)
					atomic.AddInt64(&getCount, 1)

					v.use(p)
					runtime.Gosched()
					v.clear(p)

					if p.Put(v) {
						atomic.AddInt64(&getCount, 1)
					}
				}
				runtime.Gosched()
			}
			wg.Done()
		}()
	}

	close(do)
	wg.Wait()

	return fmt.Sprintf("passed, createCount: %v, getCount: %v, putCount: %v",
		atomic.LoadInt64(&createCount), atomic.LoadInt64(&getCount), atomic.LoadInt64(&putCount))
}

func TestDoNotUseSameObject(t *testing.T) {
	testThreadCount := 1024
	testIterationCount := 1024 * 10

	testPoolSize := 1024
	p := NewFixedSizeRingPool(testPoolSize)

	msg := test(p, testThreadCount, testIterationCount)
	log.Println(msg)
}

var msg string

func BenchmarkWithDoNotUseSameObject(b *testing.B) {
	b.StopTimer()

	testThreadCount := 1024
	testIterationCount := 1024 * 10

	testPoolSize := 1024
	p := NewFixedSizeRingPool(testPoolSize)
	syncP := newSyncPool(testPoolSize)
	stdP := &stdPool{}

	b.Run("std-pool", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			msg = test(stdP, testThreadCount, testIterationCount)
		}
		//log.Println(msg)
	})

	b.Run("sync-pool", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			msg = test(syncP, testThreadCount, testIterationCount)
		}
		//log.Println(msg)
	})

	b.Run("fsz-ring-pool", func(b *testing.B) {
		b.StartTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			msg = test(p, testThreadCount, testIterationCount)
		}
		//log.Println(msg)
	})
}

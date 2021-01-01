package asyncfactory

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type (
	tstStruct struct {
		number  uint64
		a, b, c int
		d, e, f string
		g, h, i float64
		j, k, l int64
	}
	tstStructMapWithIntKey         map[int64]tstStruct
	tstStructMapWithIntKeyProvider func() *tstStructMapWithIntKey
)

var (
	blackHole int64
)

// functional tests:

func TestDoNotReturnSameObject(tst *testing.T) {
	factoryBufferSize := 64
	factorySpinCount := 1000
	testIterCount := 1000
	testThreadCount := runtime.NumCPU()

	var (
		tstStructCounter    uint64
		tstSyncAllocCounter uint64
	)

	af := NewAsyncFactory(factoryBufferSize, factorySpinCount, func(async bool) unsafe.Pointer {
		if !async {
			atomic.AddUint64(&tstSyncAllocCounter, 1)
			//tst.Errorf("can't obtain object after %v spins", factorySpinCount)
			//tst.FailNow()
		}
		r := &tstStruct{
			number: atomic.AddUint64(&tstStructCounter, 1),
		}
		return unsafe.Pointer(r)
	})

	do := make(chan bool)
	wg := sync.WaitGroup{}
	allocations := make(chan *tstStruct, testThreadCount*testIterCount+1)

	for i := 0; i < testThreadCount; i++ {
		wg.Add(1)
		go func() {
			<-do
			for i := 0; i < testIterCount; i++ {
				p := af.Alloc()
				allocations <- (*tstStruct)(p)
			}
			wg.Done()
		}()
	}
	close(do)
	wg.Wait()
	close(allocations)

	m := make(map[uint64]uint64)
	for t := range allocations {
		if _, f := m[t.number]; f {
			tst.Errorf("Same oject was returned more than one time, %v", t.number)
			tst.FailNow()
		} else {
			m[t.number] = 1
		}
	}

	tst.Logf("passed, allocated: %v, sync_allocations: %v", len(m), tstSyncAllocCounter)
}

//benchmarks:

func workloadA(m *tstStructMapWithIntKey, maxKey int) int64 {
	t := time.Now().UnixNano()
	for i := 0; i < maxKey; i++ {
		k := int64(i) //rand.Int63n(int64(maxKey))
		if v, f := (*m)[k]; f {
			v.number++
			t++
			v.j = t
			(*m)[k] = v
		} else {
			(*m)[k] = tstStruct{
				number: 1,
				j:      t,
			}
		}
	}

	r := int64(0)
	for _, mv := range *m {
		r += mv.j
	}

	return r
}

func bench(b *testing.B, maxKey, testThreadCount int, mp tstStructMapWithIntKeyProvider) {
	b.ReportAllocs()
	b.StopTimer()

	do := make(chan bool)
	wg := sync.WaitGroup{}

	//additional garbage generator --
	var c uint64 = 100
	wg.Add(1)
	go func() {
		for i := 0; i < b.N*testThreadCount; i++ {
			payload := make([]byte, b.N, b.N)
			atomic.AddUint64(&c, uint64(len(payload)))
			blackHole += int64(c) //prevent dce
		}
		wg.Done()
	}()
	//--

	for i := 0; i < testThreadCount; i++ {
		wg.Add(1)
		go func() {
			<-do
			for i := 0; i < b.N; i++ {
				m := mp()
				r := workloadA(m, maxKey)
				blackHole += r //prevent dce
			}
			wg.Done()
		}()
	}

	b.StartTimer()
	close(do)
	wg.Wait()
}

//GODEBUG=gctrace=1
func BenchmarkAsyncFactory(b *testing.B) {
	//workload:
	maxMapKeys := 2000
	testThreadCount := 1024

	//factory:
	factorySpinCount := 1000
	factoryBufferSize := testThreadCount * 5

	var (
		tstAsyncAllocCounter uint64
		tstSyncAllocCounter  uint64
	)

	af := NewAsyncFactory(factoryBufferSize, factorySpinCount, func(async bool) unsafe.Pointer {
		if !async {
			atomic.AddUint64(&tstSyncAllocCounter, 1)
		} else {
			atomic.AddUint64(&tstAsyncAllocCounter, 1)
		}
		r := make(tstStructMapWithIntKey, maxMapKeys)
		return unsafe.Pointer(&r)
	})

	b.Run("_std", func(b *testing.B) {
		bench(b, maxMapKeys, testThreadCount, func() *tstStructMapWithIntKey {
			r := make(tstStructMapWithIntKey, maxMapKeys)
			return &r
		})

	})

	b.Run("async_alloc", func(b *testing.B) {
		bench(b, maxMapKeys, testThreadCount, func() *tstStructMapWithIntKey {
			p := af.Alloc()
			return (*tstStructMapWithIntKey)(p)
		})

		h := atomic.LoadUint64(&af.head)
		t := atomic.LoadUint64(&af.tail)

		b.Logf("passed, async_allocated: %v, sync_allocations: %v, h: %v, t: %v, d1: %v",
			tstAsyncAllocCounter, tstSyncAllocCounter, h, t, af.d1)
	})
}

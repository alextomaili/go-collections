package compactmap

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	minId = 4000000
	maxId = 8000000
)

var (
	blackHole int64
)

func bench(b *testing.B, maxKey, testThreadCount int, addGcPressure bool, workload func(maxKey int) int64) {
	b.ReportAllocs()
	b.StopTimer()

	do := make(chan bool)
	wg := sync.WaitGroup{}

	//additional garbage generator --
	if addGcPressure {
		var c uint64 = 100
		wg.Add(1)
		go func() {
			<-do
			for i := 0; i < b.N*testThreadCount; i++ {
				payload := make([]byte, b.N, b.N)
				atomic.AddUint64(&c, uint64(len(payload)))
				blackHole += int64(c) //prevent dce
			}
			wg.Done()
		}()
	}
	//--

	for i := 0; i < testThreadCount; i++ {
		wg.Add(1)
		go func() {
			<-do
			for i := 0; i < b.N; i++ {
				r := workload(maxKey)
				blackHole += r //prevent dce
			}
			wg.Done()
		}()
	}

	b.StartTimer()
	close(do)
	wg.Wait()
}

func BenchmarkCompactMap(b *testing.B) {
	b.StopTimer()

	//workload:
	additionalGc := false
	maxMapKeys := 2000
	mapCapacity := maxMapKeys / 2
	testThreadCount := 1024

	rand.Seed(time.Now().Unix())
	data := make([]tstDataA, 0)
	for i := 0; i < maxMapKeys; i++ {
		key := KeyType(rand.Int31n(maxId-minId) + minId)
		data = append(data, tstDataA{
			key: key,
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	mapsPool := sync.Pool{New: func() interface{} {
		return NewIntKeyMap(emptyTstStructA.Size(), mapCapacity)
	}}

	b.Run("std_map____", func(b *testing.B) {
		bench(b, maxMapKeys, testThreadCount, additionalGc, func(maxKey int) int64 {
			var (
				r int64
				_ = tstStructA{}
			)

			m := make(map[KeyType]tstStructA, mapCapacity)
			r = int64(len(m))

			for i, _ := range data {
				m[data[i].key] = data[i].value
			}

			for i, _ := range data {
				s, f := m[data[i].key]
				if !f {
					panic("key expected")
				}
				r += int64(s.x)
			}

			return r
		})
	})

	b.Run("intk_map____", func(b *testing.B) {
		bench(b, maxMapKeys, testThreadCount, additionalGc, func(maxKey int) int64 {
			var (
				r int64
				b = tstStructA{}
			)

			m := NewIntKeyMap(emptyTstStructA.Size(), mapCapacity)
			r = int64(m.Len())

			for i, _ := range data {
				m.Put(data[i].key, &data[i].value)
			}

			for i, _ := range data {
				f := m.Get(data[i].key, &b)
				if !f {
					panic("key expected")
				}
				r += int64(b.x)
			}

			return r
		})
	})

	b.Run("intk_map_pool", func(b *testing.B) {
		bench(b, maxMapKeys, testThreadCount, additionalGc, func(maxKey int) int64 {
			var (
				r int64
				b = tstStructA{}
			)

			m := mapsPool.Get().(*IntKeyMap)
			r = int64(m.Len())

			for i, _ := range data {
				m.Put(data[i].key, &data[i].value)
			}

			for i, _ := range data {
				f := m.Get(data[i].key, &b)
				if !f {
					panic("key expected")
				}
				r += int64(b.x)
			}

			mapsPool.Put(m)
			return r
		})
	})
}

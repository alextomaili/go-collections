package compactset

import (
	"testing"
	"fmt"
	"runtime"
)

const (
	first_key_to_test = 0
	keys_to_test = 80000

	first_key_to_bench = 0
	keys_to_bench = 80000

	mem_bench_sets_count int = 1000
	mem_bench_steps_count int = 16
)

// -------------------------------------------------------------------------
// functional tests
// -------------------------------------------------------------------------

func verify(t *testing.T, s CompactIntegerSet, value KeyType)  {
	s.Add(value)
	if !s.Contains(value) {
		t.Errorf("Set should contains %d", value)
	}
}

func _testA(t *testing.T, s CompactIntegerSet)  {
	values := []uint32 {0, 17, 63, 64, 12234, 65536}

	fmt.Printf("verify: \n")
	for _, i := range values {
		verify(t, s, KeyType(i))
		fmt.Printf("verified %d\n", i)
	}

	fmt.Printf("iterate: \n")
	s.Iterate(func(value valueType) {fmt.Printf("iterated: %d\n", value)})
}

func _testB(t *testing.T, s CompactIntegerSet)  {
	startKey := first_key_to_test
	maxKey := keys_to_test

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			s.Add(KeyType(i))
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			if !s.Contains(KeyType(i)) {
				t.Errorf("Set must contain %d", i)
			}
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 != 0 {
			if s.Contains(KeyType(i)) {
				t.Errorf("Set mustn't contain %d", i)
			}
		}
	}
}


func TestA_BitmapIntegerSet(t *testing.T) {
	_testA(t, NewBitmapIntegerSet())
}

func TestA_LinearAddressingBitmapIntegerSet(t *testing.T) {
	_testA(t, NewLinearAddressingBitmapIntegerSet())
}

func TestB_BitmapIntegerSet(t *testing.T) {
	_testB(t, NewBitmapIntegerSet())
}

func TestB_LinearAddressingBitmapIntegerSet(t *testing.T) {
	_testB(t, NewLinearAddressingBitmapIntegerSet())
}

// -------------------------------------------------------------------------
// benchmarks
// -------------------------------------------------------------------------
var holder int = 0
func _benchmarkGetKeys(b *testing.B, s CompactIntegerSet, evenOnly bool) {
	b.StopTimer()

	//fill
	startKey := first_key_to_bench
	maxKey := keys_to_bench

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 || !evenOnly {
			s.Add(KeyType(i))
		}
	}

	for i := 0; i < b.N; i++ {

		b.StartTimer()
		for i := startKey; i<=maxKey; i++ {
			if i % 2 == 0 || !evenOnly {
				if !s.Contains(KeyType(i)) {
					panic(fmt.Errorf("Set must contain %d", i))
				}
				holder++
			}
		}

		if evenOnly {
			for i := startKey; i <= maxKey; i++ {
				if i % 2 != 0 {
					if s.Contains(KeyType(i)) {
						panic(fmt.Errorf("Set mustn't contain %d", i))
					}
				}
				holder++
			}
		}
		b.StopTimer()
	}
}

func BenchmarkGetKeys_BitmapIntegerSetA(b *testing.B) {
	_benchmarkGetKeys(b, NewBitmapIntegerSet(), true)
}

func BenchmarkGetKeys_LinearAddressingBitmapIntegerSetA(b *testing.B) {
	_benchmarkGetKeys(b, NewLinearAddressingBitmapIntegerSet(), true)
}

func BenchmarkGetKeys_BitmapIntegerSetB(b *testing.B) {
	_benchmarkGetKeys(b, NewBitmapIntegerSet(), false)
}

func BenchmarkGetKeys_LinearAddressingBitmapIntegerSetB(b *testing.B) {
	_benchmarkGetKeys(b, NewLinearAddressingBitmapIntegerSet(), false)
}

// ------------------------------------------------------------------------
// memory allocation
// ------------------------------------------------------------------------
func _worm_up() {
	var stats1 runtime.MemStats
	var stats2 runtime.MemStats
	var stats3 runtime.MemStats

	for i := 0; i < 10; i++ {
		runtime.GC()
		runtime.ReadMemStats(&stats1);
		runtime.GC()
		runtime.ReadMemStats(&stats2);
		runtime.GC()
		runtime.ReadMemStats(&stats3);
	}
	fmt.Printf("bytes %d %d\n", stats2.HeapAlloc - stats1.HeapAlloc, stats2.Alloc - stats1.Alloc);
	fmt.Printf("bytes %d %d\n", stats3.HeapAlloc - stats2.HeapAlloc, stats3.Alloc - stats2.Alloc);
}

func _alloc() uint64 {
	var stats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}


func _runSetTest(factory func() CompactIntegerSet) []CompactIntegerSet  {
	_worm_up()
	n := mem_bench_sets_count

	hs := make([]CompactIntegerSet, 0, int(16 * 8 * n))

	before := _alloc()
	for i := 0; i < n; i++ {
		h := factory()
		hs = append(hs, h)
	}
	after := _alloc()
	emptyPerMap := float64(after - before) / float64(n)
	fmt.Printf("Bytes used for %d empty maps: %d, bytes/map %.1f\n", n, after - before, emptyPerMap)

	k := 1
	cnt := 1
	for p := 1; p < mem_bench_steps_count; p++ {
		before = _alloc()
		for i := 0; i < n; i++ {
			h := factory()
			for j := 0; j < k; j++ {
				h.Add(KeyType(j))
				cnt++
			}
			hs = append(hs, h)
		}
		after = _alloc()
		fullPerMap := float64(after - before) / float64(n)
		fmt.Printf("Bytes used for %d maps with %d entries: %d, bytes/map %.1f\n", n, k, after - before, fullPerMap)
		fmt.Printf("Bytes per entry %.1f\n", (fullPerMap - emptyPerMap) / float64(k))
		k *= 2
	}
	fmt.Println("TOTAL ENTRIES:", cnt)

	return hs
}

func TestA_MemoryFor_BitmapIntegerSet(t *testing.T) {
	hs := _runSetTest(func() CompactIntegerSet {
		return NewBitmapIntegerSet()
	})
	holder +=len(hs)
}

func TestA_MemoryFor_LinearAddressingBitmapIntegerSet(t *testing.T) {
	hs := _runSetTest(func() CompactIntegerSet {
		return NewLinearAddressingBitmapIntegerSet()
	})
	holder +=len(hs)
}
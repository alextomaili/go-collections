package targeting

import (
	"github.com/alextomaili/go-collections/src/compactset"
	"math/rand"
	"testing"
)

func BenchmarkRandomAccessRead(b *testing.B) {
	b.StopTimer()

	set := Uint32Set{}
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set[idx] = struct{}{}
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contain(idx) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessRead64(b *testing.B) {
	b.StopTimer()

	set := Uint32Set{}
	for i := 0; i < 64; i++ {
		idx := uint32(rand.Int31n(64))
		set[idx] = struct{}{}
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		if set.Contain(idx) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadBitmapIntegerSet(b *testing.B) {
	b.StopTimer()

	set := compactset.NewBitmapIntegerSet()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set.Add(compactset.KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contains(compactset.KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadBitmapIntegerSet64(b *testing.B) {
	b.StopTimer()

	set := compactset.NewBitmapIntegerSet()
	for i := 0; i < 64; i++ {
		idx := uint32(rand.Int31n(64))
		set.Add(compactset.KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		if set.Contains(compactset.KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadOABitmapIntegerSet(b *testing.B) {
	b.StopTimer()

	set := compactset.NewOABitmapIntegerSet()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set.Add(compactset.KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contains(compactset.KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadOABitmapIntegerSet64(b *testing.B) {
	b.StopTimer()

	set := compactset.NewOABitmapIntegerSet()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		set.Add(compactset.KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		if set.Contains(compactset.KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	//fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

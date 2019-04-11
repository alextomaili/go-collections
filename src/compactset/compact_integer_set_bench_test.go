package compactset

import (
	"testing"
	"math/rand"
	"fmt"
)

func BenchmarkRandomAccessReadOABitmapIntegerSet(b *testing.B) {
	b.StopTimer()

	set := NewOABitmapIntegerSet()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set.Add(KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contains(KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadBitmapIntegerSet(b *testing.B) {
	b.StopTimer()

	set := NewBitmapIntegerSet()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set.Add(KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contains(KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessReadBitmapIntegerSet64(b *testing.B) {
	b.StopTimer()

	set := NewBitmapIntegerSet()
	for i:=0; i < 64; i++ {
		idx := uint32(rand.Int31n(64))
		set.Add(KeyType(idx))
	}

	cnt := 0
	b.StartTimer()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		if set.Contains(KeyType(idx)) {
			cnt++
		}
	}
	b.StopTimer()

	fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}
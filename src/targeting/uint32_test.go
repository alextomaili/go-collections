package targeting

import (
	"testing"
	"math/rand"
	"fmt"
)

func BenchmarkRandomAccessRead(b *testing.B) {
	b.StopTimer()

	set := Uint32Set{}
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		set[idx] = struct{}{}
	}

	cnt := 0
	b.StartTimer()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(int32(b.N)))
		if set.Contain(idx) {
			cnt++
		}
	}
	b.StopTimer()

	fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}

func BenchmarkRandomAccessRead64(b *testing.B) {
	b.StopTimer()

	set := Uint32Set{}
	for i:=0; i < 64; i++ {
		idx := uint32(rand.Int31n(64))
		set[idx] = struct{}{}
	}

	cnt := 0
	b.StartTimer()
	for i:=0; i < b.N; i++ {
		idx := uint32(rand.Int31n(64))
		if set.Contain(idx) {
			cnt++
		}
	}
	b.StopTimer()

	fmt.Printf("b.N: %d, hit count: %d\n", b.N, cnt)
}
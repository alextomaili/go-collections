package compactmap

import (
	"testing"
	"unsafe"
)

type counterType uint32

// value type
func (d *counterType) Size() int {
	return 4
}

func (d *counterType) WriteTo(p unsafe.Pointer) {
	*(*counterType)(p) = *d
}

func (d *counterType) ReadFrom(p unsafe.Pointer) {
	*d = *(*counterType)(p)
}

func BenchmarkIntKeyMapVsMap(b *testing.B) {
	b.Run("IntKeyMap", func(b *testing.B) {
		m := NewIntKeyMap(4, 100)

		v := counterType(256)
		for i := 0; i < b.N; i++ {
			for k := 0; k < 1000; k++ {
				m.Put(KeyType(k%10), &v)
			}
		}
	})

	b.Run("map[]", func(b *testing.B) {
		m := make(map[KeyType]counterType)

		v := counterType(256)
		for i := 0; i < b.N; i++ {
			for k := 0; k < 1000; k++ {
				m[KeyType(k%10)] = v
			}
		}
	})
}

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

func TestIntKeyMapWorksFine(t *testing.T) {
	im := NewIntKeyMap(4, 100)
	one := counterType(1)
	for k := 0; k < 1000; k++ {
		p, isNew := im.UpsertAndReturnPointer(KeyType(k % 10))
		if isNew {
			one.WriteTo(p)
		} else {
			*(*counterType)(p)++
		}
	}

	m := make(map[KeyType]counterType)
	for k := 0; k < 1000; k++ {
		m[KeyType(k%10)]++
	}

	var ic counterType
	for k := 0; k < 1000; k++ {
		key := KeyType(k % 10)

		has := im.Get(key, &ic)
		if !has {
			t.Error("no key")
		}
		if ic != m[key] {
			t.Error("diff", ic, "/", m[key])
		}
	}
}

func BenchmarkIntKeyMapVsMap(b *testing.B) {
	b.Run("IntKeyMap", func(b *testing.B) {
		b.ReportAllocs()
		m := NewIntKeyMap(4, 100)
		one := counterType(1)
		for i := 0; i < b.N; i++ {
			for k := 0; k < 1000; k++ {
				p, isNew := m.UpsertAndReturnPointer(KeyType(k % 10))
				if isNew {
					one.WriteTo(p)
				} else {
					*(*counterType)(p)++
				}
			}
		}
	})

	b.Run("map[]", func(b *testing.B) {
		b.ReportAllocs()
		m := make(map[KeyType]counterType)
		for i := 0; i < b.N; i++ {
			for k := 0; k < 1000; k++ {
				m[KeyType(k%10)]++
			}
		}
	})
}

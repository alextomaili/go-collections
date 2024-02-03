package lhmap

import (
	"math/rand"
	"testing"
)

type (
	benchDataR struct {
		key   tstKeyR
		nKey  tstKeyR
		value tstValueR
	}
)

var (
	tstValueRExample tstValueR
	blackHole        float64
)

func benchSet(s int) ([]benchDataR, float64) {
	r := make([]benchDataR, 0, s)
	cs := float64(0)

	for i := 0; i < s; i++ {
		k := tstKeyR{a: rand.Uint32(), b: rand.Uint32()}
		nKey := tstKeyR{
			a: k.a + 1,
			b: k.b - 1,
		}
		r = append(r, benchDataR{
			key:   k,
			nKey:  nKey,
			value: tstValueR(rand.Float64()),
		})
	}

	for _, dv := range r {
		cs = cs + float64(dv.value)
	}

	return r, cs
}

func BenchmarkLhMapVsMap(b *testing.B) {
	b.StopTimer()
	bs, cs := benchSet(4096)

	b.Run("LhMap", func(b *testing.B) {
		b.StopTimer()
		m := NewLhMap(func() KeyType { return &tstKeyR{} }, tstValueRExample.Size(), len(bs))
		for _, d := range bs {
			m.Put(&d.key, &d.value)
		}
		b.StartTimer()

		var v tstValueR
		for i := 0; i < b.N; i++ {
			//check positive keys
			blackHole = 0
			for _, d := range bs {
				if f := m.Get(&d.key, &v); f {
					blackHole = blackHole + float64(v)
				}
			}
			if blackHole != cs {
				b.Error("Upps, wrong data into map")
			}

			//negative keys (LhMap may be worse here die to collision resolution method)
			for _, d := range bs {
				if f := m.Get(&d.nKey, &v); f {
					blackHole = blackHole + float64(v)
				}
			}

		}
	})

	b.Run("Map", func(b *testing.B) {
		b.StopTimer()
		m := make(map[tstKeyR]tstValueR, len(bs))
		for _, d := range bs {
			m[d.key] = d.value
		}
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			//check positive keys
			blackHole = 0
			for _, d := range bs {
				if v, f := m[d.key]; f {
					blackHole = blackHole + float64(v)
				}
			}
			if blackHole != cs {
				b.Error("Upps, wrong data into map")
			}

			//negative keys (LhMap may be worse here die to collision resolution method)
			for _, d := range bs {
				if v, f := m[d.nKey]; f {
					blackHole = blackHole + float64(v)
				}
			}

		}
	})
}

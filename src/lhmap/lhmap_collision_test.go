package lhmap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestPutGetDeletePutAgainWithCollision(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: &tstKeyK{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyK{} }, emptyTstStructA.Size(), 10, )

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
	}

	if m.Len() != len(data) {
		t.Error(fmt.Sprintf("Invalid len actual:%v but expectd %v", m.Len(), len(data)))
	}

	b := tstStructA{}
	for i, _ := range data {
		k := data[i].key
		if !m.Get(k, &b) {
			t.Error(fmt.Sprintf("map must contains key: %v", k))
		}
		if b != data[i].value {
			t.Error(fmt.Sprintf("map must contains data for key: %v, actual: [%v], expected: [%v]", k, b, data[i].value))
		}
	}

	deleted := 0
	for i, _ := range data {
		k := data[i].key
		if k.Hash()%2 == 0 {
			if !m.Del(k) {
				t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
			}
			deleted++
		}
	}

	if m.Len() != len(data) - deleted {
		t.Error(fmt.Sprintf("Invalid len actual:%v but expectd %v", m.Len(), (len(data) - deleted)))
	}

	for i, _ := range data {
		k := data[i].key
		if k.Hash()%2 == 0 {
			if m.Get(k, nil) {
				t.Error(fmt.Sprintf("map must't contains key: %v, i: %v", k, i))
			}
		}
	}

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
	}
	if m.Len() != len(data) {
		t.Error(fmt.Sprintf("Invalid len actual:%v but expectd %v", m.Len(), len(data)))
	}

	for i, _ := range data {
		k := data[i].key
		if !m.Get(k, nil) {
			t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
		}
	}
}


package compactmap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSimplePutAndThenGet(t *testing.T) {
	data := []tstDataA{
		{
			key: KeyType(rand.Uint32()),
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		},
	}

	m := NewIntKeyMap(emptyTstStructA.Size(), 10)

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
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
}

func TestPutGetDelete(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: KeyType(rand.Uint32()),
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewIntKeyMap(emptyTstStructA.Size(), 10)

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
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

	for i, _ := range data {
		k := data[i].key
		if k%2 == 0 {
			if !m.Del(k) {
				t.Error(fmt.Sprintf("map must contains key: %v", k))
			}
		} else if i+1 < len(data) {
			m.Put(k, &data[i+1].value)
		}
	}

	for i, _ := range data {
		k := data[i].key
		if k%2 == 0 {
			if m.Get(k, nil) {
				t.Error(fmt.Sprintf("map must't contains key: %v", k))
			}
		} else if i+1 < len(data) {
			if !m.Get(k, &b) {
				t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
			}
			if b != data[i+1].value {
				t.Error(fmt.Sprintf("map must contains data for key: %v, actual: [%v], expected: [%v]", k, b, data[i].value))
			}
		}
	}
}

func TestPutGetClear(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: KeyType(rand.Uint32()),
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewIntKeyMap(emptyTstStructA.Size(), 10)

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
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

	m.Clear()

	for i, _ := range data {
		k := data[i].key
		if m.Get(k, nil) {
			t.Error(fmt.Sprintf("map must't contains key: %v", k))
		}
	}

	for m.generation > 1 {
		m.Clear()
	}

	for i, _ := range data {
		k := data[i].key
		if m.Get(k, nil) {
			t.Error(fmt.Sprintf("map must't contains key: %v", k))
		}
	}
}

func TestPutAndThenGetWithRehash(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 1024; i++ {
		data = append(data, tstDataA{
			key: KeyType(rand.Uint32()),
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewIntKeyMap(emptyTstStructA.Size(), 10)

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
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
}

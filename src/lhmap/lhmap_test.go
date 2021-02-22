package lhmap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

func TestSimplePutAndThenGet(t *testing.T) {
	data := []tstDataA{
		{
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		},
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

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
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

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
		if k.Hash()%2 == 0 {
			if !m.Del(k) {
				t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
			}
		} else if i+1 < len(data) {
			m.Put(k, &data[i+1].value)
		}
	}

	for i, _ := range data {
		k := data[i].key
		if k.Hash()%2 == 0 {
			if m.Get(k, nil) {
				t.Error(fmt.Sprintf("map must't contains key: %v, i: %v", k, i))
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
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

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
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

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

func TestPutDelVisitAll(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
	}

	vm := make(map[tstKeyA]struct{})
	for i, _ := range data {
		k := data[i].key
		if k.Hash()%2 == 0 {
			if !m.Del(k) {
				t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
			}
		} else {
			tk := k.(*tstKeyA)
			vm[*tk] = struct{}{}
		}
	}

	k := tstKeyA{}
	m.VisitAll(func(idx int, key unsafe.Pointer, p unsafe.Pointer) {
		k.ReadFrom(key)
		delete(vm, k)
	})

	if len(vm) > 0 {
		t.Error(fmt.Sprintf("all key should be visited!"))
	}
}

func TestPutDelVisit(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

	for i, _ := range data {
		m.Put(data[i].key, &data[i].value)
	}

	vm := make(map[tstKeyA]int)
	for i, _ := range data {
		k := data[i].key
		if k.Hash()%2 == 0 {
			if !m.Del(k) {
				t.Error(fmt.Sprintf("map must contains key: %v, i: %v", k, i))
			}
		} else {
			tk := k.(*tstKeyA)
			vm[*tk] = i
		}
	}

	k := tstKeyA{}
	start := 0
	for {
		start = m.Visit(start, 9, func(idx int, key unsafe.Pointer, p unsafe.Pointer) {
			k.ReadFrom(key)
			delete(vm, k)
		})
		if start == 0 {
			break
		}
	}

	if len(vm) > 0 {
		t.Error(fmt.Sprintf("all key should be visited! [%v]", vm))
	}
}

func TestPutGetDeletePutAgain(t *testing.T) {
	data := make([]tstDataA, 0)

	for i := 0; i < 40; i++ {
		data = append(data, tstDataA{
			key: &tstKeyA{a: rand.Uint32(), b: rand.Uint32(), c: rand.Uint32()},
			value: tstStructA{
				t:   time.Now().Unix(),
				x:   rand.Int31(),
				f32: rand.Float32(),
				f64: rand.Float64(),
			},
		})
	}

	m := NewLhMap(func() KeyType { return &tstKeyA{} }, emptyTstStructA.Size(), 10, )

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

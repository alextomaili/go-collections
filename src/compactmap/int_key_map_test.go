package compactmap

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

type (
	tstStructA struct {
		t   int64
		x   int32
		f32 float32
		f64 float64
	}

	tstDataA struct {
		key   KeyType
		value tstStructA
	}
)

func (d *tstStructA) Size() int {
	return int(
		0 +
			unsafe.Sizeof(d.t) +
			unsafe.Sizeof(d.x) +
			unsafe.Sizeof(d.f32) +
			unsafe.Sizeof(d.f64) +
			0)
}

func (d *tstStructA) WriteTo(p unsafe.Pointer) {
	*(*int64)(p) = d.t
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.t))

	*(*int32)(p) = d.x
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.x))

	*(*float32)(p) = d.f32
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.f32))

	*(*float64)(p) = d.f64
}

func (d *tstStructA) ReadFrom(p unsafe.Pointer) {
	d.t = *(*int64)(p)
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.t))

	d.x = *(*int32)(p)
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.x))

	d.f32 = *(*float32)(p)
	p = unsafe.Pointer(uintptr(p) + unsafe.Sizeof(d.f32))

	d.f64 = *(*float64)(p)
}

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

	s := &tstStructA{}
	m := NewIntKeyMap(s.Size(), 10)

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

	s := &tstStructA{}
	m := NewIntKeyMap(s.Size(), 10)

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

	s := &tstStructA{}
	m := NewIntKeyMap(s.Size(), 10)

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

	s := &tstStructA{}
	m := NewIntKeyMap(s.Size(), 10)

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

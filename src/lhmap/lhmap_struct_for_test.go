package lhmap

import "unsafe"

type (
	tstKeyA struct {
		a uint32
		b uint32
		c uint32
	}

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

var (
	emptyTstStructA = &tstStructA{}
	emptyTstKeyA    = &tstKeyA{}
)

//value type
func (d *tstStructA) Size() int {
	return int(unsafe.Sizeof(tstStructA{}))
}

func (d *tstStructA) WriteTo(p unsafe.Pointer) {
	*(*tstStructA)(p) = *d
}

func (d *tstStructA) ReadFrom(p unsafe.Pointer) {
	*d = *(*tstStructA)(p)
}

// key type
func (d *tstKeyA) Size() int {
	return int(unsafe.Sizeof(tstKeyA{}))
}

func (d *tstKeyA) WriteTo(p unsafe.Pointer) {
	*(*tstKeyA)(p) = *d
}

func (d *tstKeyA) ReadFrom(p unsafe.Pointer) {
	*d = *(*tstKeyA)(p)
}

func (d *tstKeyA) Hash() int {
	return int(d.a ^ d.b ^ d.c)
}

func (d *tstKeyA) Equals(p unsafe.Pointer) bool {
	x := tstKeyA{}
	x.ReadFrom(p)
	return *d == x
}

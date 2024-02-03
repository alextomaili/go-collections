package lhmap

import "unsafe"

type (
	tstKeyR struct {
		a uint32
		b uint32
	}

	tstValueR float64
)

// key
func (d *tstKeyR) Size() int {
	return int(unsafe.Sizeof(tstKeyR{}))
}

func (d *tstKeyR) WriteTo(p unsafe.Pointer) {
	*(*tstKeyR)(p) = *d
}

func (d *tstKeyR) ReadFrom(p unsafe.Pointer) {
	*d = *(*tstKeyR)(p)
}

func (d *tstKeyR) Hash() int {
	return int(d.a ^ d.b)
}

func (d *tstKeyR) Equals(p unsafe.Pointer) bool {
	x := tstKeyR{}
	x.ReadFrom(p)
	return *d == x
}

// value
func (d *tstValueR) Size() int {
	return int(unsafe.Sizeof(float64(0)))
}

func (d *tstValueR) WriteTo(p unsafe.Pointer) {
	*(*tstValueR)(p) = *d
}

func (d *tstValueR) ReadFrom(p unsafe.Pointer) {
	*d = *(*tstValueR)(p)
}

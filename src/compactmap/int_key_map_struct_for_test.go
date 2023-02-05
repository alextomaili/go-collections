package compactmap

import "unsafe"

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

var (
	emptyTstStructA = &tstStructA{}
)

func (d *tstStructA) Size() int {
	return int(unsafe.Sizeof(tstStructA{}))
}

func (d *tstStructA) WriteTo(p unsafe.Pointer) {
	*(*tstStructA)(p) = *d
}

func (d *tstStructA) ReadFrom(p unsafe.Pointer) {
	*d = *(*tstStructA)(p)
}

/*
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
*/

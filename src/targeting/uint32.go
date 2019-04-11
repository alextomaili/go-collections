package targeting

import "encoding/json"

// A Uint32Set struct represents a set of the integers
type Uint32Set map[uint32]struct{}

// UnmarshalJSON parses JSON encoded data and stores value into current set of integers
func (c *Uint32Set) UnmarshalJSON(b []byte) (err error) {
	raw := []uint32{}
	if err = json.Unmarshal(b, &raw); err == nil {
		res := make(map[uint32]struct{}, len(raw))
		for _, value := range raw {
			res[value] = struct{}{}
		}
		*c = res
	}
	return
}

// Valid checks that set of integers exists and not nil
func (a Uint32Set) Valid() bool {
	return a != nil
}

// Contain checks that needed integer exists in the set
func (a Uint32Set) Contain(check uint32) bool {
	_, ok := a[check]
	return ok
}

// add adds each integer from set of integers into current set
func (a Uint32Set) add(b Uint32Set) Uint32Set {
	for v := range b {
		a[v] = struct{}{}
	}
	return a
}

// multiply makes intersection of sets
func (a Uint32Set) multiply(b Uint32Set) Uint32Set {
	for v := range a {
		if _, ok := b[v]; !ok {
			delete(a, v)
		}
	}
	return a
}

// subtract makes subtraction of sets
func (a Uint32Set) subtract(b Uint32Set) Uint32Set {
	for v := range a {
		if _, ok := b[v]; ok {
			delete(a, v)
		}
	}
	return a
}

func NewUint32Set(src []uint32) Uint32Set {
	res := make(Uint32Set, len(src))
	for _, v := range src {
		res[v] = struct{}{}
	}

	return res
}

// mergeUint32Set merges two collections of set of integers into one. Each collection is an includes and excludes sets
func mergeUint32Set(incA, excA, incB, excB Uint32Set) (include, exclude Uint32Set) {
	switch {
	case incA != nil && incB != nil:
		include = incA.multiply(incB)
	case incA == nil && incB != nil:
		include = incB
	case incA != nil && incB == nil:
		include = incA
	}
	switch {
	case excA != nil && excB != nil:
		exclude = excA.add(excB)
	case excA == nil && excB != nil:
		exclude = excB
	case excA != nil && excB == nil:
		exclude = excA
	}
	return
}

// Uint32SetCheck returns TRUE if value integer doesn't exist into includes checks or if it exists into excludes check.
// Otherwise it returns FALSE
func Uint32SetCheck(include, exclude Uint32Set, value uint32) bool {
	return include.Valid() && !include.Contain(value) || exclude.Valid() && exclude.Contain(value)
}


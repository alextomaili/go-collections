package lhmap

import (
	"unsafe"
)

type flagType uint16

const (
	generationMask          = 0x7FFF
	deletedFlag    flagType = 0x8000
)

const (
	//hash table to store slots ---------------------
	//we should protect system against to large array's allocation
	//linear addressing uses arrays, we must limit capacity
	initialPower    uint = 3
	maxPower        uint = 30
	initialCapacity      = int(1 << initialPower)
	maxCapacity          = int(1 << maxPower)

	//good value from java framework
	defaultLoadFactor = float32(0.75)

	//used to calculate hash from key, by the way key ^ (key >> hashShift)
	//intended to involve high and low bits to the hash calculation
	//this value is used in java8
	hashShift = 16
)

type (
	LhMap struct {
		loadFactor          float32
		threshold           int
		capacity            int
		itemSize            int
		dataSize            int
		keySize             int
		headerSize          int
		flagSize            int
		data                []byte
		liveItemsCount      int
		allocatedItemsCount int
		generation          flagType
		keyCtr              func() KeyType
		tmpKey              KeyType
	}

	KeyType interface {
		Size() int
		ReadFrom(p unsafe.Pointer)
		WriteTo(p unsafe.Pointer)
		Hash() int
		Equals(p unsafe.Pointer) bool
	}

	MapValue interface {
		Size() int
		ReadFrom(p unsafe.Pointer)
		WriteTo(p unsafe.Pointer)
	}

	Visitor func(idx int, key unsafe.Pointer, p unsafe.Pointer)
)

func calcThreshold(capacity int, loadFactor float32) int {
	return int(float32(capacity) * loadFactor)
}

func hash(k KeyType, tLen int) int {
	h := k.Hash()
	//from java.util.HashMap, java 1.8
	h = h ^ (h >> hashShift)
	//length must be a non-zero power of 2, faster than index % tableLen
	return int(h) & (tLen - 1)
}

func capacityToPowerOf2(capacity int) int {
	power := uint(initialCapacity)
	for ; (1 << power) < capacity; {
		power++
		if power > maxPower {
			panic("max capacity reached")
		}
	}
	return 1 << power
}

func NewLhMap(keyCtr func() KeyType, dataSize int, capacity int) *LhMap {
	capacity = capacityToPowerOf2(capacity)
	tmpKey := keyCtr()

	s := &LhMap{
		loadFactor: defaultLoadFactor,
		threshold:  calcThreshold(capacity, defaultLoadFactor),
		capacity:   capacity,

		keyCtr: keyCtr,
		tmpKey: tmpKey,

		keySize:  tmpKey.Size(),
		flagSize: int(unsafe.Sizeof(flagType(0))),

		dataSize: dataSize,

		liveItemsCount:      0,
		allocatedItemsCount: 0,
		generation:          1,
	}

	s.headerSize = s.keySize + s.flagSize
	s.itemSize = s.headerSize + s.dataSize
	size := s.capacity * s.itemSize
	s.data = make([]byte, size, size)

	return s
}

func (s *LhMap) shift(index int) int {
	return index * s.itemSize
}

func (s *LhMap) key(index int) KeyType {
	p := s.pKey(index)
	s.tmpKey.ReadFrom(p)
	return s.tmpKey
}

func (s *LhMap) setKey(index int, k KeyType) {
	p := s.pKey(index)
	k.WriteTo(p)
}

func (s *LhMap) flag(index int) flagType {
	return *(*flagType)(unsafe.Pointer(&s.data[s.shift(index)+s.keySize]))
}

func (s *LhMap) setFlag(index int, f flagType) {
	*(*flagType)(unsafe.Pointer(&s.data[s.shift(index)+s.keySize])) = f
}

func (s *LhMap) isEmptySlot(index int) bool {
	f := s.flag(index)
	deleted := f&deletedFlag > 0
	generation := f & generationMask
	return generation != s.generation && !deleted
}

func (s *LhMap) pKey(index int) unsafe.Pointer {
	return unsafe.Pointer(&s.data[s.shift(index)])
}

func (s *LhMap) pData(index int) unsafe.Pointer {
	return unsafe.Pointer(&s.data[s.shift(index)+s.headerSize])
}

func (s *LhMap) Clear() {
	s.generation = (s.generation + 1) & generationMask
	s.liveItemsCount = 0
	s.allocatedItemsCount = 0

	//if wrap around - make new slice and start with gen == 1 again
	if s.generation <= 0 {
		size := s.capacity * s.itemSize
		s.data = make([]byte, size, size)
		s.generation = 1
	}
}

// search until we either find the key, or find an empty slot.
func (s *LhMap) findSlotByLinearProbing(key KeyType) (int, bool) {
	index := hash(key, s.capacity) // compute hashcode

	for i := 0; i < s.capacity; i++ {
		deleted := s.flag(index)&deletedFlag > 0

		if !deleted && s.isEmptySlot(index) {
			return index, false
		}

		if !deleted && key.Equals(s.pKey(index)) {
			return index, true
		}

		//next probe
		index++
		if index >= s.capacity {
			index = 0
		}
	}
	return -1, false //nothing found, table is full
}

func (s *LhMap) ensureCapacity(newCount int) bool {
	if newCount > maxCapacity {
		return false
	}
	if newCount <= s.threshold {
		return true //already have enough capacity
	}
	//enlarge size
	s.rehash(s.capacity << 1)
	return true
}

func (s *LhMap) rehash(newCapacity int) {
	oldS := &LhMap{}
	*oldS = *s

	newSize := newCapacity * s.itemSize
	s.capacity = newCapacity
	s.data = make([]byte, newSize, newSize)
	s.threshold = calcThreshold(newCapacity, s.loadFactor)

	for i := 0; i < oldS.capacity; i++ {
		oldShift := oldS.shift(i)

		if !oldS.isEmptySlot(i) {
			mK := oldS.key(i)
			idx, _ := s.findSlotByLinearProbing(mK)
			shift := s.shift(idx)
			//copy memory
			for j := 0; j < s.itemSize; j++ {
				s.data[shift+j] = oldS.data[oldShift+j]
			}
		}
	}
}

func (s *LhMap) findOrInsertSlot(key KeyType) (int, bool) {
	if !s.ensureCapacity(s.allocatedItemsCount + 1) {
		panic("no more capacity")
	}
	i, found := s.findSlotByLinearProbing(key)
	if i < 0 {
		panic("internal error. shouldn't happens, ensureCapacity should provide empty slots")
	}
	return i, found
}

func (s *LhMap) Put(key KeyType, value MapValue) {
	if value == nil {
		panic("nil value is not allowed")
	}

	index, found := s.findOrInsertSlot(key)
	if !found {
		s.liveItemsCount++
		s.allocatedItemsCount++
		s.setKey(index, key)
		s.setFlag(index, s.generation & ^deletedFlag)
	}

	p := s.pData(index)
	value.WriteTo(p)
}

func (s *LhMap) Get(key KeyType, value MapValue) bool {
	index, found := s.findSlotByLinearProbing(key)
	if !found {
		return false
	}

	if value == nil {
		return true //just report what key is exist
	}

	p := s.pData(index)
	value.ReadFrom(p)
	return true
}

func (s *LhMap) Del(key KeyType) bool {
	index, found := s.findSlotByLinearProbing(key)
	if !found {
		return false
	}

	s.setFlag(index, s.flag(index)|deletedFlag)
	s.liveItemsCount--
	return true
}

func (s *LhMap) Len() int {
	return s.liveItemsCount
}

func (s *LhMap) VisitAll(visitor Visitor) {
	for i := 0; i < s.capacity; i++ {
		if !s.isEmptySlot(i) {
			k := s.pKey(i)
			p := s.pData(i)
			visitor(i, k, p)
		}
	}
}

func (s *LhMap) Visit(start, count int, visitor Visitor) (next int) {
	if start >= s.capacity || start < 0 {
		return 0
	}

	v := 0
	i := start
	for ; i < s.capacity && v < count; i++ {
		if !s.isEmptySlot(i) {
			k := s.pKey(i)
			p := s.pData(i)
			visitor(i, k, p)
			v++
		}
	}

	if i == s.capacity {
		return 0
	} else {
		return i
	}
}

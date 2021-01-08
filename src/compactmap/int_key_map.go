package compactmap

import "unsafe"

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
	KeyType  uint32
	flagType uint16

	IntKeyMap struct {
		loadFactor float32
		threshold  int
		capacity   int
		itemSize   int
		dataSize   int
		keySize    int
		headerSize int
		flagSize   int
		data       []byte
		itemsCount int
		generation flagType
	}

	MapValue interface {
		Size() int
		ReadFrom(p unsafe.Pointer)
		WriteTo(p unsafe.Pointer)
	}

	Visitor func(key KeyType, p unsafe.Pointer)
)

func calcThreshold(capacity int, loadFactor float32) int {
	return int(float32(capacity) * loadFactor)
}

func hash(h KeyType, tLen int) int {
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

func NewIntKeyMap(dataSize int, capacity int) *IntKeyMap {
	capacity = capacityToPowerOf2(capacity)

	s := &IntKeyMap{
		loadFactor: defaultLoadFactor,
		threshold:  calcThreshold(capacity, defaultLoadFactor),
		capacity:   capacity,

		keySize:  int(unsafe.Sizeof(KeyType(0))),
		flagSize: int(unsafe.Sizeof(flagType(0))),

		dataSize: dataSize,

		itemsCount: 0,
		generation: 1,
	}

	s.headerSize = s.keySize + s.flagSize
	s.itemSize = s.headerSize + s.dataSize
	size := s.capacity * s.itemSize
	s.data = make([]byte, size, size)

	return s
}

func (s *IntKeyMap) shift(index int) int {
	return index * s.itemSize
}

func (s *IntKeyMap) key(index int) KeyType {
	return *(*KeyType)(unsafe.Pointer(&s.data[s.shift(index)]))
}

func (s *IntKeyMap) setKey(index int, k KeyType) {
	*(*KeyType)(unsafe.Pointer(&s.data[s.shift(index)])) = k
}

func (s *IntKeyMap) flag(index int) flagType {
	return *(*flagType)(unsafe.Pointer(&s.data[s.shift(index)+s.keySize]))
}

func (s *IntKeyMap) setFlag(index int, f flagType) {
	*(*flagType)(unsafe.Pointer(&s.data[s.shift(index)+s.keySize])) = f
}

func (s *IntKeyMap) isEmptySlot(index int) bool {
	f := s.flag(index)
	return f != s.generation
}

func (s *IntKeyMap) pData(index int) unsafe.Pointer {
	return unsafe.Pointer(&s.data[s.shift(index)+s.headerSize])
}

func (s *IntKeyMap) Clear() {
	s.generation++
	s.itemsCount = 0

	//if wrap around - make new slice and start with gen == 1 again
	if s.generation <= 0 {
		size := s.capacity * s.itemSize
		s.data = make([]byte, size, size)
		s.generation = 1
	}
}

// search until we either find the key, or find an empty slot.
func (s *IntKeyMap) findSlotByLinearProbing(key KeyType) (int, bool) {
	index := hash(key, s.capacity) // compute hashcode

	for i := 0; i < s.capacity; i++ {
		if s.isEmptySlot(index) {
			return index, false
		}

		k := s.key(index)
		if k == key {
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

func (s *IntKeyMap) ensureCapacity(newCount int) bool {
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

func (s *IntKeyMap) rehash(newCapacity int) {
	oldS := &IntKeyMap{}
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

func (s *IntKeyMap) findOrInsertSlot(key KeyType) (int, bool) {
	if !s.ensureCapacity(s.itemsCount + 1) {
		panic("no more capacity")
	}
	i, found := s.findSlotByLinearProbing(key)
	if i < 0 {
		panic("internal error. shouldn't happens, ensureCapacity should provide empty slots")
	}
	return i, found
}

func (s *IntKeyMap) Put(key KeyType, value MapValue) {
	if value == nil {
		panic("nil value is not allowed")
	}

	index, found := s.findOrInsertSlot(key)
	if !found {
		s.itemsCount++
		s.setKey(index, key)
		s.setFlag(index, s.generation)
	}

	p := s.pData(index)
	value.WriteTo(p)
}

func (s *IntKeyMap) Get(key KeyType, value MapValue) bool {
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

func (s *IntKeyMap) Del(key KeyType) bool {
	index, found := s.findSlotByLinearProbing(key)
	if !found {
		return false
	}

	s.setFlag(index, 0)
	s.itemsCount--
	return true
}

func (s *IntKeyMap) Len() int {
	return s.itemsCount
}

func (s *IntKeyMap) VisitAll(visitor Visitor) {
	for i := 0; i < s.capacity; i++ {
		if !s.isEmptySlot(i) {
			k := s.key(i)
			p := s.pData(i)
			visitor(k, p)
		}
	}
}

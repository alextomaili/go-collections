package compactset

/*
    Implementation of Set for positive integer keys
    Intended to save required memory
    Stores keys in the sparse bit index

    slot1 [0] -> [uint32 mask]
    slot2 [1] -> [0 1 2 3 ...7][8 9 ... 15] ... [16 .. 31]
                      1         1                          <--- we have two id here: 32 and 40
    .....
    slotX [x] -> [....]

*/

const (
	//hash table to store slots ---------------------
	//we should protect system against to large array's allocation
	//linear addressing uses arrays, we must limit capacity
	initial_power uint = 3
	max_power uint = 30
	initial_length int = int(1 << initial_power)
	max_length int = int(1 << max_power)

	//good value from java framework
	default_load_factor = float32(0.75)
	// ----------------------------------------------

	//compact set -----------------------------------
	//set can holds numbers from 0 to 2 097 120
	//it uses 2 bytes for holds slots and 4 bytes for hold bit mask
	maxSlotNumber uint16 = 0xFFFF   //65 535 * 32 = 2 097 120
	dividerPower uint32 = 5        //1 << 5 = 32
	maskSize uint32 = 1 << dividerPower//32
	maxKey uint32 = uint32(maxSlotNumber) << dividerPower

	//used to calculate hash from key, by the way key ^ (key >> hashShift)
	//intended to involve high and low bits to the hash calculation
	//this value is used in java8
	hashShift = 16
	//-----------------------------------------------
)

type KeyType uint32
type slotKeyType uint16

const emptyKey slotKeyType = 0

type valueType uint32

const emptyValue valueType = 0

func calc(key KeyType) (slotIdx slotKeyType, mask valueType) {
	slotIdx = slotKeyType(key >> dividerPower)        //like (key / 32)
	bitNumber := key - KeyType(slotIdx) << dividerPower //like (key % 32)
	mask = valueType(1 << bitNumber)
	return
}

func iterate(slotIdx slotKeyType, slotValue valueType, iterator CompactIntegerSetIterator) {
	x := valueType(slotIdx) << dividerPower
	for i := uint32(0); i < maskSize; i++ {
		mask := valueType(1 << i)
		if (slotValue & mask) != 0 {
			v := x + valueType(i)
			iterator(v)
		}
	}
}

type CompactIntegerSetIterator func(value valueType)
type CompactIntegerSet interface {
	Add(key KeyType)
	Contains(key KeyType) bool
	Len() int
	Iterate(iterator CompactIntegerSetIterator)
}

// --------------------------------------------------------------------------
// based on std map[]
// --------------------------------------------------------------------------
type BitmapIntegerSet struct {
	slots map[slotKeyType]valueType
	count int
}

func NewBitmapIntegerSet() *BitmapIntegerSet {
	s := BitmapIntegerSet{
		slots: make(map[slotKeyType]valueType),
		count: 0}
	return &s
}

func (b *BitmapIntegerSet) Len() int {
	return b.count
}

func (b *BitmapIntegerSet) Iterate(iterator CompactIntegerSetIterator) {
	for slotIdx, slotValue := range b.slots {
		iterate(slotIdx, slotValue, iterator)
	}
}

func (b *BitmapIntegerSet) Contains(key KeyType) bool {
	slotIdx, mask := calc(key)

	slotValue, found := b.slots[slotIdx]
	if !found {
		return false;
	}
	return (slotValue & mask) != 0
}

func (b *BitmapIntegerSet) Add(key KeyType) {
	slotIdx, mask := calc(key)

	slotValue, found := b.slots[slotIdx]
	if !found {
		b.slots[slotIdx] = mask
		b.count++
	} else {
		if (slotValue & mask) == 0 {
			b.slots[slotIdx] = slotValue | mask
			b.count++
		}
	}
}

// --------------------------------------------------------------------------
// based on linear addressing, with linear probing collision resolution
// --------------------------------------------------------------------------
type LinearAddressingBitmapIntegerSet struct {
	loadFactor       float32
	threshold        int
	slotCount        int
	count            int
	hasEmptyKey      bool
	valueForEmptyKey valueType
	capacity         int
	keys             []slotKeyType
	values           []valueType
}

func calc_threshold(capacity int, load_factor float32) int {
	return int(float32(capacity) * load_factor);
}

func hash(h slotKeyType, tLen int) int {
	//from java.util.HashMap, java 1.8
	h = h ^ (h >> hashShift)
	//length must be a non-zero power of 2, faster than index % tableLen
	return int(h) & (tLen - 1)
}

func NewLinearAddressingBitmapIntegerSet() *LinearAddressingBitmapIntegerSet {
	s := LinearAddressingBitmapIntegerSet{
		loadFactor: default_load_factor,
		threshold: calc_threshold(initial_length, default_load_factor),
		slotCount: 0,
		count: 0,
		hasEmptyKey: false,
		valueForEmptyKey: 0,
		capacity: initial_length,
		keys: make([]slotKeyType, initial_length),
		values: make([]valueType, initial_length)}
	return &s
}

// search until we either find the key, or find an empty slot.
func (s *LinearAddressingBitmapIntegerSet) findSlotByLinearProbing(key slotKeyType) (int, bool) {
	index := hash(key, s.capacity) // compute hashcode

	for i := 0; i < s.capacity; i++ {
		k := s.keys[index];
		if k == emptyKey {
			return index, false
		}
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

func (s *LinearAddressingBitmapIntegerSet) ensureCapacity(newCount int) bool {
	if newCount > max_length {
		return false
	}
	if newCount <= s.threshold {
		return true //already have enough capacity
	}
	//enlarge size
	s.rehash(s.capacity << 1);
	return true
}

func (s *LinearAddressingBitmapIntegerSet) rehash(newCapacity int) {
	newTable := make([]slotKeyType, newCapacity)
	newValueTable := make([]valueType, newCapacity)

	for i := 0; i < s.capacity; i++ {
		mK := s.keys[i]
		mV := s.values[i]
		if mK != emptyKey {
			idx, _ := s.findSlotByLinearProbing(mK)
			newTable[idx] = mK
			newValueTable[idx] = mV
		}
	}

	s.capacity = newCapacity
	s.keys = newTable
	s.values = newValueTable
	s.threshold = calc_threshold(newCapacity, s.loadFactor)
}

func (s *LinearAddressingBitmapIntegerSet) findSlot(key slotKeyType) (int, bool) {
	if !s.ensureCapacity(s.slotCount + 1) {
		panic("no more capacity")
	}
	i, found := s.findSlotByLinearProbing(key)
	if i < 0 {
		panic("internal error. shouldn't happens, ensureCapacity should provide empty slots")
	}
	return i, found
}

func (s *LinearAddressingBitmapIntegerSet) putValue(key slotKeyType, value valueType) {
	if key == emptyKey {
		if !s.hasEmptyKey {
			s.hasEmptyKey = true
		}
		s.valueForEmptyKey = value
		return
	}

	i, found := s.findSlot(key)
	if !found {
		s.keys[i] = key
		s.slotCount++
	}
	s.values[i] = value
}

func (s *LinearAddressingBitmapIntegerSet) getValue(key slotKeyType) (value valueType, exist bool) {
	if key == emptyKey {
		if s.hasEmptyKey {
			return s.valueForEmptyKey, true
		} else {
			return valueType(0), false
		}
	}

	i, found := s.findSlotByLinearProbing(key)
	if found {
		return s.values[i], true
	} else {
		return valueType(0), false
	}
}

func (s *LinearAddressingBitmapIntegerSet) Contains(key KeyType) bool {
	slotIdx, mask := calc(key)

	slotValue, found := s.getValue(slotIdx)
	if !found {
		return false;
	}
	return (slotValue & mask) != 0
}

func (s *LinearAddressingBitmapIntegerSet) Add(key KeyType) {
	slotIdx, mask := calc(key)

	slotValue, found := s.getValue(slotIdx)
	if !found {
		s.putValue(slotIdx, mask)
		s.count++
	} else {
		if (slotValue & mask) == 0 {
			s.putValue(slotIdx, slotValue | mask)
			s.count++
		}
	}
}

func (s *LinearAddressingBitmapIntegerSet) Len() int {
	return s.count
}

func (s *LinearAddressingBitmapIntegerSet) Iterate(iterator CompactIntegerSetIterator) {
	if s.hasEmptyKey {
		iterate(emptyKey, s.valueForEmptyKey, iterator)
	}

	for i, slotIdx := range s.keys {
		if slotIdx == emptyKey {
			continue
		}
		slotValue := s.values[i]
		iterate(slotIdx, slotValue, iterator)
	}
}

//clear before return to pool
func (s *LinearAddressingBitmapIntegerSet) Clear() {
	s.count = 0

	s.hasEmptyKey = false
	s.valueForEmptyKey = emptyValue

	s.slotCount = 0
	for i, _ := range s.keys {
		s.keys[i] = emptyKey
		s.values[i] = emptyValue
	}
}
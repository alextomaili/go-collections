package concurrentmap

import (
	"unsafe"
	"sync"
	"sync/atomic"
	"fmt"
)

const (
	//hash table to store slots
	initial_power uint = 3
	max_power uint = 30
	initial_length int = int(1 << initial_power)
	max_length int = int(1 << max_power)

	//good value from java framework
	default_load_factor = float32(0.75)

	//used to calculate hash from key
	hashShift = 16

	decreaseCapacityRate = 2

	//key types
	KT_INT = 0
	KT_STRING = 1
)


type Data interface{}

type entry struct {
	intKey  int
	strKey  string
	value   Data
	deleted bool
}

func buildNewEntry(keyKind byte, intKey int, strKey *string, value Data, deleted bool) *entry  {
	newEntry := &entry {
		value: value,
		deleted: deleted }

	switch keyKind {
	case KT_INT:
		newEntry.intKey = intKey;
	case KT_STRING:
		newEntry.intKey = intKey;
		newEntry.strKey = *strKey;
	default:
		panic(fmt.Errorf("Unsuported keyKind: %d", keyKind))
	}

	return newEntry
}

type mapData struct {
	loadFactor     float32
	threshold      int
	capacity       int
	estimatedCount int32            //used by atomic operations
	data           []unsafe.Pointer // data[0]->(*entry), data[1]->(*entry)
}

type CMap struct {
	keyKind byte
	data         unsafe.Pointer // -> mapData
	mutatorsLock sync.RWMutex
}

func calc_threshold(capacity int, load_factor float32) int {
	return int(float32(capacity) * load_factor);
}

func hash(h int, tLen int) int {
	//from java.util.HashMap, java 1.8
	h = h ^ (h >> hashShift)
	//length must be a non-zero power of 2, faster than index % tableLen
	return int(h) & (tLen - 1)
}

func (m *CMap) kGet(intKey  int, strKey  *string) Data {
	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
	map_capacity := map_data.capacity
	isIntKey := m.keyKind == KT_INT
	index := hash(intKey, map_capacity) // compute hashcode


	for i := 0; i < map_capacity; i++ {
		//volatile red, other thread may write in same time
		entryPtr := atomic.LoadPointer(&map_data.data[index])
		oldEntry := (*entry)(entryPtr)

		if oldEntry == nil {
			return nil
		}

		if !oldEntry.deleted && oldEntry.intKey == intKey && (isIntKey || oldEntry.strKey == *strKey) {
			return oldEntry.value
		}

		//next probe
		index++
		if index >= map_capacity {
			index = 0
		}
	}
	return nil
}

func (m *CMap) kDel(intKey  int, strKey  *string) Data {
	m.mutatorsLock.RLock()
	defer m.mutatorsLock.RUnlock()

	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
	map_capacity := map_data.capacity
	isIntKey := m.keyKind == KT_INT
	index := hash(intKey, map_capacity) // compute hashcode

	for i := 0; i < map_capacity; i++ {
		//volatile red, other thread may write in same time
		entryPtr := atomic.LoadPointer(&map_data.data[index])
		oldEntry := (*entry)(entryPtr)

		if oldEntry == nil {
			return nil
		}

		if !oldEntry.deleted && oldEntry.intKey == intKey && (isIntKey || oldEntry.strKey == *strKey) {
			newEntry := buildNewEntry(m.keyKind, intKey, strKey, nil, true)
			if atomic.CompareAndSwapPointer(&map_data.data[index], entryPtr, unsafe.Pointer(newEntry)) {
				atomic.AddInt32(&map_data.estimatedCount, -1)
			}
			return oldEntry.value
		}

		//next probe
		index++
		if index >= map_capacity {
			index = 0
		}
	}
	return nil
}

func (m *CMap) kPut(intKey  int, strKey  *string, value Data) Data {
	//prepare new entry
	newEntry := buildNewEntry(m.keyKind, intKey, strKey, value, false)
	//W-LOCK (possible): if grow is really needed - write lock used inside to protect it
	m.ensureCapacity()

	//R-LOCK: read lock used inside, try concurrent insert
	data, upsert := m.put(newEntry)
	if upsert {
		map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
		atomic.AddInt32(&map_data.estimatedCount, 1)
		return data
	}

	//W-LOCK: exclusive insert, we can't insert anyway, try grow again and insert with write lock
	m.mutatorsLock.Lock()
	defer m.mutatorsLock.Unlock()

	new_map_data := m.rehash()
	d, done := _put(new_map_data, newEntry, m.keyKind)
	if !done {
		panic("Can't insert entry, no capacity after grow")
	}
	atomic.StorePointer(&m.data, unsafe.Pointer(new_map_data))

	return d
}

func (m *CMap) GetCount() int {
	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
	count := atomic.LoadInt32(&map_data.estimatedCount)
	if count < 0 {
		return 0
	} else {
		return int(count)
	}
}

func (m *CMap) ensureCapacity() {
	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
	estimatedCount := atomic.LoadInt32(&map_data.estimatedCount)
	if int(estimatedCount) + 1 <= map_data.threshold {
		return //already have enough capacity
	}

	m.mutatorsLock.Lock()
	defer m.mutatorsLock.Unlock()

	new_map_data := m.rehash()
	atomic.StorePointer(&m.data, unsafe.Pointer(new_map_data))
}

//always calls under W-LOCK
func (m *CMap) rehash() *mapData {
	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read

	//check, may be we will have necessary amount of free space without grow size ----------------
	occupied := 0
	for i := 0; i < map_data.capacity; i++ {
		//we are om W-LOCK and owns map_data after volatile read here, we can read directly
		oldEntry := (*entry)(map_data.data[i])
		if oldEntry != nil && !oldEntry.deleted {
			occupied++
		}
	}
	occupied++
	new_capacity := map_data.capacity
	if occupied > map_data.threshold {
		//ok we don't have necessary free space, grow
		new_capacity = map_data.capacity << 1
		if new_capacity > max_length {
			panic("no more capacity")
		}
	} else if occupied < int(map_data.threshold / decreaseCapacityRate) {
		//its time to decrease capacity, we have too many deleted items
		new_capacity = roundToMinimalPowerOf2(int(float32(occupied) / map_data.loadFactor))
	}
	// -------------------------------------------------------------------------------------------


	new_map_data := &mapData{
		loadFactor: default_load_factor,
		threshold: calc_threshold(new_capacity, default_load_factor),
		capacity: new_capacity,
		estimatedCount: 0,
		data: make([]unsafe.Pointer, new_capacity)}

	for i := 0; i < map_data.capacity; i++ {
		//we are om W-LOCK and owns map_data after volatile read here, we can read directly
		oldEntry := (*entry)(map_data.data[i])
		if oldEntry != nil && !oldEntry.deleted {
			_put(new_map_data, oldEntry, m.keyKind)
			new_map_data.estimatedCount++
		}
	}

	return new_map_data
}

func (m *CMap) put(newEntry *entry) (Data, bool) {
	m.mutatorsLock.RLock()
	defer m.mutatorsLock.RUnlock()

	map_data := (*mapData)(atomic.LoadPointer(&m.data)) //volatile read
	return _put(map_data, newEntry, m.keyKind)
}

func _put(map_data *mapData, newEntry *entry, keyKind byte) (Data, bool) {
	map_capacity := map_data.capacity
	isIntKey := keyKind == KT_INT
	index := hash(newEntry.intKey, map_capacity) // compute hashcode

	for i := 0; i < map_capacity; i++ {
		//try store if not exist
		entryPtr := atomic.LoadPointer(&map_data.data[index])
		oldEntry := (*entry)(entryPtr)
		if oldEntry == nil {
			if atomic.CompareAndSwapPointer(&map_data.data[index], entryPtr, unsafe.Pointer(newEntry)) {
				return nil, true
			}
		}

		//ok, slot is occupied may be this is our value, in this case we must write here,
		//if value is deleted - ok it will be actual again
		entryPtr = atomic.LoadPointer(&map_data.data[index])
		oldEntry = (*entry)(entryPtr)

		if  oldEntry.intKey == newEntry.intKey && (isIntKey || oldEntry.strKey == newEntry.strKey) {
			//we must override this value, because this is our key
			atomic.StorePointer(&map_data.data[index], unsafe.Pointer(newEntry))
			if oldEntry.deleted {
				return nil, true
			} else {
				return oldEntry.value, true
			}

		}

		//next probe
		index++
		if index >= map_capacity {
			index = 0
		}
	}
	return nil, false
}

func roundToMinimalPowerOf2(capacity int) int {
	power := uint(initial_power)
	for ; (1 << power) < capacity; {
		power++
	}
	return 1 << power;
}

//--------------------------------------------------------------------------------------
// map with integer key
//--------------------------------------------------------------------------------------
type CIntKeyMap struct {
	CMap
}

//empty strKey
var emptyStringKey string = ""

func NewCIntKeyMap() *CIntKeyMap {
	map_data := &mapData{
		loadFactor: default_load_factor,
		threshold: calc_threshold(initial_length, default_load_factor),
		capacity: initial_length,
		estimatedCount: 0,
		data: make([]unsafe.Pointer, initial_length)}
	m := CIntKeyMap{CMap{keyKind: KT_INT, mutatorsLock: sync.RWMutex{}}}
	atomic.StorePointer(&m.data, unsafe.Pointer(map_data))
	return &m
}

func (m *CIntKeyMap) Get(key int) Data {
	return m.kGet(key, &emptyStringKey);
}

func (m *CIntKeyMap) Del(key int) Data {
	return m.kDel(key, &emptyStringKey);
}

func (m *CIntKeyMap) Put(key int, value Data) Data {
	return m.kPut(key, &emptyStringKey, value)
}

//--------------------------------------------------------------------------------------
// map with string key
//--------------------------------------------------------------------------------------
type CStrKeyMap struct {
	CMap
}

func NewCStrKeyMap() *CStrKeyMap {
	map_data := &mapData{
		loadFactor: default_load_factor,
		threshold: calc_threshold(initial_length, default_load_factor),
		capacity: initial_length,
		estimatedCount: 0,
		data: make([]unsafe.Pointer, initial_length)}
	m := CStrKeyMap{CMap{keyKind: KT_STRING, mutatorsLock: sync.RWMutex{}}}
	atomic.StorePointer(&m.data, unsafe.Pointer(map_data))
	return &m
}

/* java implementation
   The value 31 was chosen because it is an odd prime. If it were even and the multiplication overflowed, information
   would be lost, as multiplication by 2 is equivalent to shifting. The advantage of using a prime is less clear,
   but it is traditional. A nice property of 31 is that the multiplication can be replaced by a shift and a
   subtraction for better performance: 31 * i == (i << 5) - i. Modern VMs do this sort of optimization automatically.
   from pprof disasm:
      20ms       20ms     471ed1: LEAQ 0x1(CX), SI
         .          .     471ed5: INCQ BX
      70ms       70ms     471ed8: MOVZX 0(CX), DI
         .          .     471edb: MOVQ DX, R8
      40ms       40ms     471ede: SHLQ $0x5, DX
      30ms       30ms     471ee2: SUBQ R8, DX
     150ms      150ms     471ee5: ADDQ DI, DX
      70ms       70ms     471ee8: MOVQ SI, CX
      10ms       10ms     471eeb: CMPQ AX, BX
         .          .     471eee: JL 0x471ed1
*/
func strHash(key *string) int  {
	var h int = 0;
	for _, c := range []byte(*key) {
		h = 31 * h + int(c);
	}
	return h
}

func (m *CStrKeyMap) Get(key string) Data {
	return m.kGet(strHash(&key), &key);
}

func (m *CStrKeyMap) Del(key string) Data {
	return m.kDel(strHash(&key), &key);
}

func (m *CStrKeyMap) Put(key string, value Data) Data {
	return m.kPut(strHash(&key), &key, value)
}

//--------------------------------------------------------------------------------------
//debug only  code below:
//--------------------------------------------------------------------------------------
func VerifyForDoubledValuesCIntKeyMap(m *CIntKeyMap) {
	m.mutatorsLock.Lock()
	defer m.mutatorsLock.Unlock()

	map_data := (*mapData)(atomic.LoadPointer(&m.data))
	mp := make(map[int]Data, len(map_data.data))

	for i := 0; i < map_data.capacity; i++ {
		entryPtr := atomic.LoadPointer(&map_data.data[i])
		cEntry := (*entry)(entryPtr)

		if cEntry == nil {
			continue
		}

		d, found := mp[cEntry.intKey]
		if !found {
			mp[cEntry.intKey] = cEntry.value
		} else {
			panic(fmt.Sprintf("two value for key fund, key %d, value %v, value %v", cEntry.intKey, d, cEntry.value))
		}
	}
}

func VerifyForDoubledValuesCStrKeyMap(m *CStrKeyMap) {
	m.mutatorsLock.Lock()
	defer m.mutatorsLock.Unlock()

	map_data := (*mapData)(atomic.LoadPointer(&m.data))
	mp := make(map[string]Data, len(map_data.data))

	for i := 0; i < map_data.capacity; i++ {
		entryPtr := atomic.LoadPointer(&map_data.data[i])
		cEntry := (*entry)(entryPtr)

		if cEntry == nil {
			continue
		}

		d, found := mp[cEntry.strKey]
		if !found {
			mp[cEntry.strKey] = cEntry.value
		} else {
			panic(fmt.Sprintf("two value for key fund, key %s, value %v, value %v", cEntry.strKey, d, cEntry.value))
		}
	}
}


func _dump(md *mapData)  {
	fmt.Printf("map_data[loadFactor: %f, threshold: %d, capacity: %d, estimatedCount: %d, len(data) %d]\n",
		md.loadFactor, md.threshold, md.capacity, md.estimatedCount, len(md.data))
}

func DebugCMap() {
	m := NewCIntKeyMap()

	fmt.Printf("old value: %s\n", m.Put(0, Data("v0-0")))

	fmt.Printf("old value: %s\n", m.Put(1, Data("v1-0")))
	fmt.Printf("get value: %d\n", m.Get(1))

	fmt.Printf("old value: %s\n", m.Put(1, Data("v1-1")))
	fmt.Printf("get value: %d\n", m.Get(1))

	fmt.Printf("get value: %d\n", m.Get(100))
}

func DebugStrKeyMap(m *CStrKeyMap, key string, value Data, expectedOldValue Data)  {
	oldValue := m.Put("key#0", "val-0")
	if expectedOldValue != nil && oldValue != expectedOldValue {
		fmt.Printf("Old value expected %d, but found %d\n", expectedOldValue, oldValue)
	}
	v := m.Get(key)
	if v != value {
		fmt.Printf("expected value %s not found, but found: %s\n", value, v)
	}
}
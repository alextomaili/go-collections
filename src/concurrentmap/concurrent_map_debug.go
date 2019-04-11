package concurrentmap

import (
	"fmt"
	"sync/atomic"
)

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

# go-collections

## compactset - package 

Implementation of Set for positive integer keys
Intended to save required memory
Stores keys in the sparse bit index

```bash
slot1 [0] -> [uint32 mask]
slot2 [1] -> [0 1 2 3 ...7][8 9 ... 15] ... [16 .. 31]
                  1         1                          <--- we have two id here: 32 and 40
.....
slotX [x] -> [....]
```

BitmapIntegerSet - compactset based on standard go map

OABitmapIntegerSet - compactset based on open addressing map, has less memeory overhead than standard go map

## concurrentmap - package

Concurrent map implemented without sync.RWMutex and based on atomic package    
Supports two types of keys: "int" and "string" without overhead of boxing/unboxong 

CIntKeyMap - map with "int" key

CStrKeyMap - map with "string" key

CIntKeyMap 4th times fatser than standard map[] protected by RWMutex on read from many threads test  

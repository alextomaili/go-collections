package compactmap

const (
	//hash table to store slots ---------------------
	//we should protect system against to large array's allocation
	//linear addressing uses arrays, we must limit capacity
	initialPower  uint = 3
	maxPower      uint = 30
	initialLength int  = int(1 << initialPower)
	maxLength     int  = int(1 << maxPower)

	//good value from java framework
	defaultLoadFactor = float32(0.75)

	//used to calculate hash from key, by the way key ^ (key >> hashShift)
	//intended to involve high and low bits to the hash calculation
	//this value is used in java8
	hashShift = 16
)

type (
	KeyType uint32

	Entry struct {
		Key        KeyType
		generation uint32
	}

	IntKeyMap struct {
		loadFactor float32
		threshold  int
	}
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

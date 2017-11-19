package concurrentmap

import (
	"testing"
	"fmt"
	"runtime"
	"sync/atomic"
	"math/rand"
	"strconv"
	"time"
	"sync"
)

const (
	first_key_to_test = 0
	keys_to_test = 80000

	random_test_keys_bound = 32
	random_test_verify_count = 1024
	random_test_sleep_time_milliseconds = 10
)

func verifyCIntKeyMap(t *testing.T, m *CIntKeyMap, key int, value Data, expectedOldValue Data ) {
	oldValue := m.Put(key, value)
	if expectedOldValue != nil && oldValue != expectedOldValue {
		t.Errorf("Old value expected %d, but found %d\n", expectedOldValue, oldValue)
	}

	v := m.Get(key)
	if v != value {
		t.Errorf("expected value %d not found, but found: %d\n", value, v)
	}
}

func verifyDeleteCIntKeyMap(t *testing.T, m *CIntKeyMap, key int, expectedOldValue Data)  {
	v := m.Del(key)
	if v != expectedOldValue {
		t.Errorf("verifyDelete: actual value %d is not equals expected %d\n", v, expectedOldValue)
	}
}

func TestA(t *testing.T)  {
	m := NewCIntKeyMap()

	verifyCIntKeyMap(t, m, 0, "v0-0", nil)
	verifyCIntKeyMap(t, m, 0, "v0-1", "v0-0")

	verifyCIntKeyMap(t, m, 5, "v5-0", nil)
	verifyCIntKeyMap(t, m, 5, "v5-1", "v5-0")

	verifyCIntKeyMap(t, m, 100, "v100-0", nil)
	verifyCIntKeyMap(t, m, 100, "v100-1", "v100-0")

	verifyCIntKeyMap(t, m, 5, "v5-2", "v5-1")

	verifyDeleteCIntKeyMap(t, m, 100, "v100-1")
	verifyDeleteCIntKeyMap(t, m, 100, nil)
}

func _testB(t *testing.T, m *CIntKeyMap)  {
	startKey := first_key_to_test
	maxKey := keys_to_test

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			m.Put(i, i)
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			v := m.Get(i)
			if v == nil || v.(int) != i {
				t.Errorf("Set must contain %d", i)
			}
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 != 0 {
			v := m.Get(i)
			if v != nil {
				t.Errorf("Set mustn't contain %d, but contains %d", i, v.(int))
			}
		}
	}
}

func TestB(t *testing.T) {
	m := NewCIntKeyMap();
	_testB(t, m)
}

func TestRandomAccess(t *testing.T)  {
	m := NewCIntKeyMap();
	threadCount := runtime.NumCPU()
	do := make(chan bool)
	done := make(chan bool)
        finish := int32(0)

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			rnd := rand.New(rand.NewSource(time.Now().Unix()))
			for ;atomic.LoadInt32(&finish) == 0;{
				action := rnd.Int31n(3)
				key := rnd.Intn(random_test_keys_bound)
				value := "value" + strconv.Itoa(int(key));
				switch action {
				case 0:
					m.Put(key, value)
				case 1:
					m.Get(key)
				case 2:
					m.Del(key)
				}
				//fmt.Printf("test step: action %d, value %d\n", action, value)
			}
			done<-true
			//fmt.Printf("--> done test thread\n")
		}()
	}
	close(do) //allow all threads to work

	for i:=0; i<random_test_verify_count; i++ {
		VerifyForDoubledValuesCIntKeyMap(m)
		fmt.Println("verified ... ")
		time.Sleep(random_test_sleep_time_milliseconds * time.Millisecond)
	}

	//stop test
	atomic.StoreInt32(&finish, 1)
	for i := 0; i < threadCount; i++ {
		<-done
	}
}

var holder int = 0

func BenchmarkCIntKeyMap_GetKeys(b *testing.B) {
	b.StopTimer()
	do := make(chan bool)
	done := make(chan bool)
	//threadCount := 8
	threadCount := runtime.NumCPU()
	cnt := b.N
	//cnt = BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", b.N, runtime.NumCPU())

	tSet := NewCIntKeyMap()
	for i := 0; i <= cnt; i++ {
		tSet.Put(i, i)
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				v :=  tSet.Get(i)
				holder += v.(int)
			}
			done<-true
		}()
	}

	b.StartTimer()
	close(do)
	for i := 0; i < threadCount; i++ {
		<-done
	}
	b.StopTimer()
}


func BenchmarkRwLock_GetKeys(b *testing.B) {
	b.StopTimer()
	do := make(chan bool)
	done := make(chan bool)
	//threadCount := 8
	threadCount := runtime.NumCPU()
	cnt := b.N
	//cnt = BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", b.N, runtime.NumCPU())

	mu := new(sync.RWMutex);
	tSet := make(map[int]interface{}, cnt)
	for i := 0; i <= cnt; i++ {
		tSet[i] = i
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				mu.RLock();
				v :=  tSet[i]
				mu.RUnlock();
				holder += v.(int)
			}
			done<-true
		}()
	}

	b.StartTimer()
	close(do)
	for i := 0; i < threadCount; i++ {
		<-done
	}
	b.StopTimer()
}

func BenchmarkSyncMap_GetKeys(b *testing.B) {
	b.StopTimer()
	do := make(chan bool)
	done := make(chan bool)
	//threadCount := 8
	threadCount := runtime.NumCPU()
	cnt := b.N
	//cnt = BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", b.N, runtime.NumCPU())

	tSet := &sync.Map{}
	for i := 0; i <= cnt; i++ {
		tSet.Store(i, i)
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				v, _ :=  tSet.Load(i)
				holder += v.(int)
			}
			done<-true
		}()
	}

	b.StartTimer()
	close(do)
	for i := 0; i < threadCount; i++ {
		<-done
	}
	b.StopTimer()
}

// ---

var BENCHMARK_KEYS_COUNT int = 1000000;
func TestTimeProf(t *testing.T)  {
	do := make(chan bool)
	done := make(chan bool)
	threadCount := runtime.NumCPU()
	cnt := BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", cnt, runtime.NumCPU())

	tSet := NewCIntKeyMap()
	for i := 0; i <= cnt; i++ {
		tSet.Put(i, i)
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				v :=  tSet.Get(i)
				holder += v.(int)
			}
			done<-true
		}()
	}

	close(do)
	for i := 0; i < threadCount; i++ {
		<-done
	}
	fmt.Printf("holder -> %d, count -> %d\n", holder, tSet.GetCount());
}

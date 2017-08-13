package concurrentmap

import (
	"testing"
	"strconv"
	"time"
	"sync/atomic"
	"fmt"
	"runtime"
	"math/rand"
	"sync"
)


func verifyCStrKeyMap(t *testing.T, m *CStrKeyMap, key string, value Data, expectedOldValue Data) {
	oldValue := m.Put(key, value)
	if expectedOldValue != nil && oldValue != expectedOldValue {
		t.Errorf("Old value expected %d, but found %d\n", expectedOldValue, oldValue)
	}

	v := m.Get(key)
	if v != value {
		t.Errorf("expected value %s not found, but found: %s\n", value, v)
	}
}

func verifyDeleteCStrKeyMap(t *testing.T, m *CStrKeyMap, key string, expectedOldValue Data)  {
	v := m.Del(key)
	if v != expectedOldValue {
		t.Errorf("verifyDelete: actual value %s is not equals expected %s\n", v, expectedOldValue)
	}
}

func TestSkA(t *testing.T)  {
	m := NewCStrKeyMap()

	verifyCStrKeyMap(t, m, "key#0", "v0-0", nil)
	verifyCStrKeyMap(t, m, "key#0", "v0-1", "v0-0")

	verifyCStrKeyMap(t, m, "key#5", "v5-0", nil)
	verifyCStrKeyMap(t, m, "key#5", "v5-1", "v5-0")

	verifyCStrKeyMap(t, m, "key#100", "v100-0", nil)
	verifyCStrKeyMap(t, m, "key#100", "v100-1", "v100-0")

	verifyCStrKeyMap(t, m, "key#5", "v5-2", "v5-1")

	verifyDeleteCStrKeyMap(t, m, "key#100", "v100-1")
	verifyDeleteCStrKeyMap(t, m, "key#100", nil)
}

func _testSkB(t *testing.T, m *CStrKeyMap)  {
	startKey := first_key_to_test
	maxKey := keys_to_test

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			m.Put("key"+strconv.Itoa(i), i)
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 == 0 {
			v := m.Get("key"+strconv.Itoa(i))
			if v == nil || v.(int) != i {
				t.Errorf("Set must contain %d", i)
			}
		}
	}

	for i := startKey; i<=maxKey; i++ {
		if i % 2 != 0 {
			v := m.Get("key"+strconv.Itoa(i))
			if v != nil {
				t.Errorf("Set mustn't contain %d, but contains %d", i, v.(int))
			}
		}
	}
}

func TestSkB(t *testing.T) {
	m := NewCStrKeyMap();
	_testSkB(t, m)
}

func TestRandomAccessSk(t *testing.T)  {
	m := NewCStrKeyMap();
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
					m.Put("key"+strconv.Itoa(key), value)
				case 1:
					m.Get("key"+strconv.Itoa(key))
				case 2:
					m.Del("key"+strconv.Itoa(key))
				}
				//fmt.Printf("test step: action %d, value %d\n", action, value)
			}
			done<-true
			//fmt.Printf("--> done test thread\n")
		}()
	}
	close(do) //allow all threads to work

	for i:=0; i<random_test_verify_count; i++ {
		VerifyForDoubledValuesCStrKeyMap(m)
		fmt.Println("verified ... ")
		time.Sleep(random_test_sleep_time_milliseconds * time.Millisecond)
	}

	//stop test
	atomic.StoreInt32(&finish, 1)
	for i := 0; i < threadCount; i++ {
		<-done
	}
}

func BenchmarkCStrKeyMap_GetKeys(b *testing.B) {
	b.StopTimer()
	do := make(chan bool)
	done := make(chan bool)
	//threadCount := 8
	threadCount := runtime.NumCPU()
	cnt := b.N
	//cnt = BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", b.N, runtime.NumCPU())

	keys := make([]string, cnt+1)
	tSet := NewCStrKeyMap()
	for i := 0; i <= cnt; i++ {
		keys[i] = "key-123456-#"+strconv.Itoa(i)
		tSet.Put(keys[i], i)
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				v :=  tSet.Get(keys[i])
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

func BenchmarkRwLock_GetStrKeys(b *testing.B) {
	b.StopTimer()
	do := make(chan bool)
	done := make(chan bool)
	//threadCount := 8
	threadCount := runtime.NumCPU()
	cnt := b.N
	//cnt = BENCHMARK_KEYS_COUNT
	fmt.Printf("b.N --> %d, numCPU --> %d\n", b.N, runtime.NumCPU())

	mu := new(sync.RWMutex);
	keys := make([]string, cnt+1)
	tSet := make(map[string]interface{}, cnt)
	for i := 0; i <= cnt; i++ {
		keys[i] = "key-123456-#"+strconv.Itoa(i)
		tSet[keys[i]] = i
	}

	for i := 0; i<threadCount; i++ {
		go func() {
			<-do
			for i := 1; i <= cnt; i++ {
				mu.RLock();
				v :=  tSet[keys[i]]
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


func Benchmark_StdStrMapAccess(b *testing.B) {
	b.StopTimer()

	key := "key-123456-#10000"
	tSet := make(map[string]interface{}, 32)
	tSet[key] = 10

	b.StartTimer()
	for i := 1; i <= b.N; i++ {
		v :=  tSet[key]
		holder += v.(int)
	}
	b.StopTimer()
}


func Benchmark_StrMapAccess(b *testing.B) {
	b.StopTimer()

	key := "key-123456-#10000"
	tSet := NewCStrKeyMap()
	tSet.Put(key, 10)

	b.StartTimer()
	for i := 1; i <= b.N; i++ {
		v :=  tSet.Get(key)
		holder += v.(int)
	}
	b.StopTimer()
}

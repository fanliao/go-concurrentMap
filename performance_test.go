package concurrent

import (
	"fmt"
	//"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

var (
	listN  int
	n      int
	list   [][]interface{}
	readCM *ConcurrentMap
	readLM *lockMap
	readM  map[interface{}]interface{}
)

func init() {
	MAXPROCS := runtime.NumCPU()
	runtime.GOMAXPROCS(MAXPROCS)
	listN = MAXPROCS + 1
	n = 100000
	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", n, "\n")

	list = make([][]interface{}, listN, listN)
	for i := 0; i < listN; i++ {
		list1 := make([]interface{}, 0, n)
		for j := 0; j < n; j++ {
			list1 = append(list1, j+(i)*n/10)
		}
		list[i] = list1
	}

	readCM = NewConcurrentMap()
	readM = make(map[interface{}]interface{})
	readLM = newLockMap()
	for i := range list[0] {
		readCM.Put(i, i)
		readLM.put(i, i)
		readM[i] = i
	}
}

type lockMap struct {
	m  map[interface{}]interface{}
	rw *sync.RWMutex
}

func (t *lockMap) put(k interface{}, v interface{}) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.m[k] = v
}

func (t *lockMap) putIfNotExist(k interface{}, v interface{}) (ok bool) {
	t.rw.Lock()
	defer t.rw.Unlock()
	if _, ok = t.m[k]; !ok {
		t.m[k] = v
	}
	return
}

func (t *lockMap) get(k interface{}) (v interface{}, ok bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	v, ok = t.m[k]
	return
}

func (t *lockMap) len() int {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return len(t.m)

}

func newLockMap() *lockMap {
	return &lockMap{make(map[interface{}]interface{}), new(sync.RWMutex)}
}

func BenchmarkLockMapPut(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMapPut(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[j] = j
			}
			//wg.Done()
		}
	}
}

func BenchmarkConcurrentMapPut(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(j, j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLockMapPut2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(strconv.Itoa(j.(int)), j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMapPut2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				cm[strconv.Itoa(j.(int))] = j
			}
			//wg.Done()
		}
	}
}

func BenchmarkConcurrentMapPut2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(strconv.Itoa(j.(int)), j)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkLockMapGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			go func() {
				//itr := NewMapIterator(cm)
				//for itr.HasNext() {
				//	entry := itr.NextEntry()
				//	k := entry.key.(string)
				//	v := entry.value.(int)
				for k := range list[0] {
					_, _ = readLM.get(k)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMapGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for k := range list[0] {
				_, _ = readM[k]
			}
			//wg.Done()
		}
	}
}

func BenchmarkConcurrentMapGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			go func() {
				for k := range list[0] {
					_, _ = readCM.Get(k)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

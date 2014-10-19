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
	listN int
	n     int
	list  [][]interface{}
)

func init() {
	MAXPROCS := 4
	runtime.GOMAXPROCS(MAXPROCS)
	listN = 5
	n = 10000
	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", n, "\n")

	list = make([][]interface{}, listN, listN)
	for i := 0; i < listN; i++ {
		list1 := make([]interface{}, 0, n)
		for j := 0; j < n; j++ {
			list1 = append(list1, j+(i)*n/10)
		}
		list[i] = list1
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

func BenchmarkLockMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.put(strconv.Itoa(j.(int)), j)
					v, ok := cm.get(strconv.Itoa(j.(int)))
					if v != j || !ok {
						fmt.Println("fail in lockMap!!!!!!!!!!")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				k := strconv.Itoa(j.(int))
				cm[k] = j
				v := cm[strconv.Itoa(j.(int))]
				if v != j {
					fmt.Println("fail in Map2!!!!!!!!!!")
				}
			}
			//wg.Done()
		}
	}
}

func BenchmarkConcurrentMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(strconv.Itoa(j.(int)), j)
					v, _ := cm.Get(strconv.Itoa(j.(int)))
					if v != j {
						fmt.Println("fail in ConcurrentMap!!!!!!!!!!")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

// test project main.go
package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	cm := NewConcurrentMap()
	fmt.Println("put 1 get", cm.Put(1, "aa"))
	fmt.Println("get 1 with", cm.Get(1))

	preformance()
	fmt.Println("hello")
}

//concurrentMap
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

func duration(title string, f func()) {
	s := time.Now()
	f()
	fmt.Println(title, " spends ", time.Since(s).Nanoseconds()/1000, " micro s")
}

func preformance() {
	MAXPROCS := 4
	runtime.GOMAXPROCS(MAXPROCS)
	listN := 5
	n := 100000
	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", n)

	list := make([][]interface{}, listN, listN)
	for i := 0; i < listN; i++ {
		list1 := make([]interface{}, 0, n)
		for j := 0; j < n; j++ {
			list1 = append(list1, j+(i-1)*n/100)
		}
		list[i] = list1
	}

	l := 0
	duration("LockMap", func() {
		cm := newLockMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.putIfNotExist(strconv.Itoa(j.(int)), j)
					v, ok := cm.get(strconv.Itoa(j.(int)))
					if v != j || !ok {
						fmt.Println("fail in lockMap!!!!!!!!!!")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		l = cm.len()
	})
	fmt.Println("len=", l)

	duration("LockMap2", func() {
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
						fmt.Println("fail in lockMap2!!!!!!!!!!")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		l = cm.len()
	})
	fmt.Println("len=", l)

	var concurrentMap *ConcurrentMap
	duration("ConcurrentMap2", func() {
		cm := NewConcurrentMap()

		wg := new(sync.WaitGroup)
		wg.Add(listN)
		for i := 0; i < listN; i++ {
			k := i
			go func() {
				for _, j := range list[k] {
					cm.Put(strconv.Itoa(j.(int)), j)
					v := cm.Get(strconv.Itoa(j.(int)))
					if v != j {
						fmt.Println("fail in ConcurrentMap2!!!!!!!!!!")
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		l = int(cm.Size())
		concurrentMap = cm
	})
	fmt.Println("len=", l)

	duration("Map", func() {
		cm := make(map[interface{}]interface{})

		//wg := new(sync.WaitGroup)
		//wg.Add(listN)
		for i := 0; i < listN; i++ {
			for _, j := range list[i] {
				k := strconv.Itoa(j.(int))
				if _, ok := cm[k]; !ok {
					cm[k] = j
				}
				v := cm[strconv.Itoa(j.(int))]
				if v != j {
					fmt.Println("fail in Map!!!!!!!!!!")
				}
			}
			//wg.Done()
		}
		//wg.Wait()
		l = len(cm)
	})
	fmt.Println("len=", l)

	duration("Map2", func() {
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
		//wg.Wait()
		l = len(cm)
	})
	fmt.Println("len=", l)

}

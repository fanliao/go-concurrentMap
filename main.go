// test project main.go
package concurrent

import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

func main1() {
	a := [3]int8{1, 2, 4}
	v := reflect.ValueOf(a)
	fmt.Println(v.Type().Size())
	fmt.Println(v.Index(0).Interface())

	var va interface{} = a
	ui := (*emptyInterface)(unsafe.Pointer(&va))
	fmt.Println(ui.word)
	fmt.Printf("%x\n", uintptr(ui.word)>>8)

	//fmt.Println(bytesForArray(a, v.Type().Size()))

	//preformance()
	//m1 := make(map[int]int)
	//m2 := make(map[int]int)
	//fmt.Println("map equals?", m1 == m2)

	testHash()
	fmt.Println("hello")
}

////concurrentMap
//type lockMap struct {
//	m  map[interface{}]interface{}
//	rw *sync.RWMutex
//}

//func (t *lockMap) put(k interface{}, v interface{}) {
//	t.rw.Lock()
//	defer t.rw.Unlock()
//	t.m[k] = v
//}

//func (t *lockMap) putIfNotExist(k interface{}, v interface{}) (ok bool) {
//	t.rw.Lock()
//	defer t.rw.Unlock()
//	if _, ok = t.m[k]; !ok {
//		t.m[k] = v
//	}
//	return
//}

//func (t *lockMap) get(k interface{}) (v interface{}, ok bool) {
//	t.rw.RLock()
//	defer t.rw.RUnlock()
//	v, ok = t.m[k]
//	return
//}

//func (t *lockMap) len() int {
//	t.rw.RLock()
//	defer t.rw.RUnlock()
//	return len(t.m)

//}

//func newLockMap() *lockMap {
//	return &lockMap{make(map[interface{}]interface{}), new(sync.RWMutex)}
//}

//func duration(title string, f func()) {
//	s := time.Now()
//	f()
//	fmt.Println(title, " spends ", time.Since(s).Nanoseconds()/1000, " micro s")
//}

//func preformance() {
//	MAXPROCS := 4
//	runtime.GOMAXPROCS(MAXPROCS)
//	listN := 5
//	n := 10000
//	fmt.Println("MAXPROCS is ", MAXPROCS, ", listN is", listN, ", n is ", n, "\n")

//	list := make([][]interface{}, listN, listN)
//	for i := 0; i < listN; i++ {
//		list1 := make([]interface{}, 0, n)
//		for j := 0; j < n; j++ {
//			list1 = append(list1, j+(i)*n/10)
//		}
//		list[i] = list1
//	}

//	l := 0
//	duration("LockMap", func() {
//		cm := newLockMap()

//		wg := new(sync.WaitGroup)
//		wg.Add(listN)
//		for i := 0; i < listN; i++ {
//			k := i
//			go func() {
//				for _, j := range list[k] {
//					cm.put(strconv.Itoa(j.(int)), j)
//					v, ok := cm.get(strconv.Itoa(j.(int)))
//					if v != j || !ok {
//						fmt.Println("fail in lockMap!!!!!!!!!!")
//					}
//				}
//				wg.Done()
//			}()
//		}
//		wg.Wait()
//		l = cm.len()
//	})
//	fmt.Println("len=", l)

//	var concurrentMap *ConcurrentMap
//	duration("ConcurrentMap", func() {
//		cm := NewConcurrentMap()

//		wg := new(sync.WaitGroup)
//		wg.Add(listN)
//		for i := 0; i < listN; i++ {
//			k := i
//			go func() {
//				for _, j := range list[k] {
//					cm.Put(strconv.Itoa(j.(int)), j)
//					v := cm.Get(strconv.Itoa(j.(int)))
//					if v != j {
//						fmt.Println("fail in ConcurrentMap!!!!!!!!!!")
//					}
//				}
//				wg.Done()
//			}()
//		}
//		wg.Wait()
//		l = int(cm.Size())
//		concurrentMap = cm
//	})
//	fmt.Println("len=", l)

//	duration("Map", func() {
//		cm := make(map[interface{}]interface{})

//		//wg := new(sync.WaitGroup)
//		//wg.Add(listN)
//		for i := 0; i < listN; i++ {
//			for _, j := range list[i] {
//				k := strconv.Itoa(j.(int))
//				cm[k] = j
//				v := cm[strconv.Itoa(j.(int))]
//				if v != j {
//					fmt.Println("fail in Map2!!!!!!!!!!")
//				}
//			}
//			//wg.Done()
//		}
//		//wg.Wait()
//		l = len(cm)
//	})
//	fmt.Println("len=", l)

//	itr := NewHashIterator(concurrentMap)
//	for itr.HasNext() {
//		_ = itr.NextEntry()
//		//fmt.Println(e.key, e.value, e.hash)
//	}

//	for i := 0; i < n+(listN-1)*n/10; i++ {
//		if v := concurrentMap.Get(strconv.Itoa(i)); v != i {
//			fmt.Println("Not matched,", i, v)
//		}
//	}

//	fmt.Println("map test is done")

//}

type types struct {
	Boolean      bool
	Byte         byte
	Rune         rune
	Text         string
	Date         time.Time
	Int          int
	Int8         int8
	Int16        int16
	Int32        int32
	Int64        int64
	Uint         uint
	Uint8        uint8
	Uint16       uint16
	Uint32       uint32
	Uint64       uint64
	Float32      float32
	Float64      float64
	Complex64    complex64
	Complex128   complex128
	Uintptr      uintptr
	String       string
	SliceInt     []int
	ArrayInt1    [2]byte
	ArrayString1 [4]string
	MapIntString map[int]string
	Channel1     chan int
	Function1    func()
	Itab         sizeTester1
	Struct1      sizeTester1Impl
	Ei           interface{}
	End          bool
}

type sizeTester1 interface {
	Get() string
}

type sizeTester1Impl struct {
}

func (this sizeTester1Impl) Get() string {
	return "hello"
}

func testHash() {
	o1 := &types{}
	o2 := &types{}
	_ = o2
	k1, k2 := hashi(*o1), hashi(o1)

	fmt.Println(k1, k2)
}

package concurrent

import (
	"fmt"
	c "github.com/smartystreets/goconvey/convey"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

func TestNil(t *testing.T) {
	c.Convey("Nil cannot be as key", t, func() {
		cm := NewConcurrentMap()
		_, err := cm.Put(nil, 1)
		c.So(err, c.ShouldNotBeNil)

		var nilVal interface{} = nil
		_, err = cm.Put(nilVal, 1)
		c.So(err, c.ShouldNotBeNil)

		var nilPtr *string = nil
		nilVal = nilPtr
		_, err = cm.Put(nilVal, 1)
		c.So(err, c.ShouldNotBeNil)

		_, err = cm.Put(1, nil)
		c.So(err, c.ShouldNotBeNil)

		_, err = cm.Put(1, nilVal)
		c.So(err, c.ShouldNotBeNil)

		nilVal = nilPtr
		_, err = cm.Put(1, nilVal)
		c.So(err, c.ShouldNotBeNil)

		_, err = cm.Get(nil)
		c.So(err, c.ShouldNotBeNil)

		_, err = cm.Get(nilVal)
		c.So(err, c.ShouldNotBeNil)

		nilVal = nilPtr
		_, err = cm.Get(nilVal)
		c.So(err, c.ShouldNotBeNil)
	})
}

//// negative zero is a good test because:
////  1) 0 and -0 are equal, yet have distinct representations.
////  2) 0 is represented as all zeros, -0 isn't.
//// I'm not sure the language spec actually requires this behavior,
//// but it's what the current map implementation does.
//func TestNegativeZero(t *testing.T) {
//	m := NewConcurrentMap1(0)
//	var zero float64 = +0.0
//	var nzero float64 = math.Copysign(0.0, -1.0)

//	m.Put(zero, true)
//	m.Put(nzero, true) // should overwrite +0 entry

//	if m.Size() != 1 {
//		t.Error("length wrong", m.Size())
//	}

//	itr := NewHashIterator(m)
//	for {
//		if itr.HasNext() {
//			e := itr.NextEntry()
//			if math.Copysign(1.0, e.key.(float64)) > 0 {
//				t.Error("wrong sign")
//			}
//		} else {
//			break
//		}
//	}

//	m = NewConcurrentMap1(0)

//	m.Put(nzero, true)
//	m.Put(zero, true) // should overwrite -0.0 entry

//	if m.Size() != 1 {
//		t.Error("length wrong")
//	}

//	itr = NewHashIterator(m)
//	for {
//		if itr.HasNext() {
//			e := itr.NextEntry()
//			if math.Copysign(1.0, e.key.(float64)) < 0 {
//				t.Error("wrong sign")
//			}
//		} else {
//			break
//		}
//	}
//}

// nan is a good test because nan != nan, and nan has
// a randomized hash value.
func TestNan(t *testing.T) {
	m := NewConcurrentMap1(0) //make(map[float64]int, 0)
	nan := math.NaN()
	m.Put(nan, 1)
	m.Put(nan, 2)
	m.Put(nan, 4)
	if m.Size() != 3 {
		t.Error("length wrong")
	}
	s := 0
	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k, v := entry.key.(float64), entry.value.(int)
		if k == k {
			t.Error("nan disappeared")
		}
		if (v & (v - 1)) != 0 {
			t.Error("value wrong")
		}
		s |= v
	}
	if s != 7 {
		t.Error("values wrong")
	}
}

func TestGrowWithNaN(t *testing.T) {
	m := NewConcurrentMap1(0) //make(map[float64]int, 0)
	nan := math.NaN()
	m.Put(nan, 1)
	m.Put(nan, 2)
	m.Put(nan, 4)
	cnt := 0
	s := 0
	growflag := true

	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k, v := entry.key.(float64), entry.value.(int)
		if growflag {
			// force a hashtable resize
			for i := 0; i < 100; i++ {
				m.Put(float64(i), i)
			}
			growflag = false
		}
		if k != k {
			cnt++
			s |= v
		}
	}
	if cnt != 3 {
		t.Error("NaN keys lost during grow")
	}
	if s != 7 {
		t.Error("NaN values lost during grow")
	}
}

type FloatInt struct {
	x float64
	y int
}

func TestGrowWithNegativeZero(t *testing.T) {
	negzero := math.Copysign(0.0, -1.0)
	m := make(map[FloatInt]int, 4)
	m[FloatInt{0.0, 0}] = 1
	m[FloatInt{0.0, 1}] = 2
	m[FloatInt{0.0, 2}] = 4
	m[FloatInt{0.0, 3}] = 8
	growflag := true
	s := 0
	cnt := 0
	negcnt := 0
	// The first iteration should return the +0 key.
	// The subsequent iterations should return the -0 key.
	// I'm not really sure this is required by the spec,
	// but it makes sense.
	// TODO: are we allowed to get the first entry returned again???
	for k, v := range m {
		if v == 0 {
			continue
		} // ignore entries added to grow table
		cnt++
		if math.Copysign(1.0, k.x) < 0 {
			if v&16 == 0 {
				t.Error("key/value not updated together 1")
			}
			negcnt++
			s |= v & 15
		} else {
			if v&16 == 16 {
				t.Error("key/value not updated together 2", k, v)
			}
			s |= v
		}
		if growflag {
			// force a hashtable resize
			for i := 0; i < 100; i++ {
				m[FloatInt{3.0, i}] = 0
			}
			// then change all the entries
			// to negative zero
			m[FloatInt{negzero, 0}] = 1 | 16
			m[FloatInt{negzero, 1}] = 2 | 16
			m[FloatInt{negzero, 2}] = 4 | 16
			m[FloatInt{negzero, 3}] = 8 | 16
			growflag = false
		}
	}
	if s != 15 {
		t.Error("entry missing", s)
	}
	if cnt != 4 {
		t.Error("wrong number of entries returned by iterator", cnt)
	}
	if negcnt != 3 {
		t.Error("update to negzero missed by iteration", negcnt)
	}
}

func TestIterGrowAndDelete(t *testing.T) {
	m := make(map[int]int, 4)
	for i := 0; i < 100; i++ {
		m[i] = i
	}
	growflag := true
	for k := range m {
		//t.Log("k ad growflag", k, growflag)
		if growflag {
			// grow the table
			for i := 100; i < 1000; i++ {
				m[i] = i
			}
			// delete all odd keys
			for i := 1; i < 1000; i += 2 {
				delete(m, i)
			}
			growflag = false
		} else {
			if k&1 == 1 {
				t.Error("odd value returned")
			}
		}
	}
}

func TestIterGrowAndDelete1(t *testing.T) {
	m := NewConcurrentMap1(4) //	make(map[int]int, 4)
	for i := 0; i < 100; i++ {
		m.Put(i, i)
	}
	growflag := true
	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k := entry.key
		//t.Log("k ad growflag111111", k, growflag)
		if growflag {
			// grow the table
			for i := 100; i < 1000; i++ {
				m.Put(i, i)
			}
			// delete all odd keys
			for i := 1; i < 1000; i += 2 {
				m.Remove(i)
			}
			growflag = false
		} else {
			if k.(int)&1 == 1 {
				itr := NewHashIterator(m)
				for itr.HasNext() {
					entry := itr.NextEntry()
					if entry.key.(int)&1 == 1 {
						t.Error("odd value returned by itr")
					}
				}
				//ConcurrentMap cannot iterate the values changed outside iterator
				//t.Error("odd value returned")
			}
		}
	}
}

// make sure old bucket arrays don't get GCd while
// an iterator is still using them.
func TestIterGrowWithGC(t *testing.T) {
	m := NewConcurrentMap1(4) //	make(map[int]int, 4)
	for i := 0; i < 16; i++ {
		m.Put(i, i)
	}
	growflag := true
	bitmask := 0
	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k := entry.key.(int)
		if k < 16 {
			bitmask |= 1 << uint(k)
		}
		if growflag {
			// grow the table
			for i := 100; i < 1000; i++ {
				m.Put(i, i)
			}
			// trigger a gc
			runtime.GC()
			growflag = false
		}
	}
	if bitmask != 1<<16-1 {
		t.Error("missing key", bitmask)
	}
}

func testConcurrentReadsAfterGrowth(t *testing.T, useReflect bool) {
	if runtime.GOMAXPROCS(-1) == 1 {
		defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(16))
	}
	numLoop := 10
	numGrowStep := 250
	numReader := 16
	if testing.Short() {
		numLoop, numGrowStep = 2, 500
	}
	for i := 0; i < numLoop; i++ {
		m := NewConcurrentMap() //	make(map[int]int, 0)
		for gs := 0; gs < numGrowStep; gs++ {
			m.Put(gs, gs)
			var wg sync.WaitGroup
			wg.Add(numReader * 2)
			for nr := 0; nr < numReader; nr++ {
				go func() {
					defer wg.Done()
					itr := NewHashIterator(m)
					for itr.HasNext() {
						_ = itr.NextEntry()
					}
				}()
				go func() {
					defer wg.Done()
					for key := 0; key < gs; key++ {
						_, _ = m.Get(key)
					}
				}()
			}
			wg.Wait()
		}
	}
}

func TestConcurrentReadsAfterGrowth(t *testing.T) {
	testConcurrentReadsAfterGrowth(t, false)
}

func TestConcurrentReadsAfterGrowthReflect(t *testing.T) {
	testConcurrentReadsAfterGrowth(t, true)
}

func TestBigItems(t *testing.T) {
	var key [256]string
	for i := 0; i < 256; i++ {
		key[i] = "foo"
	}
	m := NewConcurrentMap1(4) //make(map[[256]string][256]string, 4)
	for i := 0; i < 100; i++ {
		key[37] = fmt.Sprintf("string%02d", i)
		m.Put(key, key) //m[key] = key
	}
	var keys [100]string
	var values [100]string
	i := 0
	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k := entry.key.([256]string)
		v := entry.value.([256]string)
		//for k, v := range m {
		keys[i] = k[37]
		values[i] = v[37]
		i++
	}
	sort.Strings(keys[:])
	sort.Strings(values[:])
	for i := 0; i < 100; i++ {
		if keys[i] != fmt.Sprintf("string%02d", i) {
			t.Errorf("#%d: missing key: %v", i, keys[i])
		}
		if values[i] != fmt.Sprintf("string%02d", i) {
			t.Errorf("#%d: missing value: %v", i, values[i])
		}
	}
}

type empty struct {
}

func TestEmptyKeyAndValue(t *testing.T) {
	//a := make(map[int]empty, 4)
	//b := make(map[empty]int, 4)
	//c := make(map[empty]empty, 4)
	a := NewConcurrentMap1(4)
	b := NewConcurrentMap1(4)
	c := NewConcurrentMap1(4)
	a.Put(0, empty{})       //a[0] = empty{}
	b.Put(empty{}, 0)       //b[empty{}] = 0
	b.Put(empty{}, 1)       //b[empty{}] = 1
	c.Put(empty{}, empty{}) //c[empty{}] = empty{}

	if a.Size() != 1 { // len(a) != 1 {
		t.Errorf("empty value insert problem")
	}
	if v, err := b.Get(empty{}); v != 1 || err != nil { //} b[empty{}] != 1 {
		t.Errorf("empty key returned wrong value")
	}
}

// Tests a map with a single bucket, with same-lengthed short keys
// ("quick keys") as well as long keys.
func TestSingleBucketMapStringKeys_DupLen(t *testing.T) {
	testMapLookups(t, NewConcurrentMapFromMap(map[interface{}]interface{}{
		"x":    "x1val",
		"xx":   "x2val",
		"foo":  "fooval",
		"bar":  "barval", // same key length as "foo"
		"xxxx": "x4val",
		strings.Repeat("x", 128): "longval1",
		strings.Repeat("y", 128): "longval2",
	}))
}

// Tests a map with a single bucket, with all keys having different lengths.
func TestSingleBucketMapStringKeys_NoDupLen(t *testing.T) {
	testMapLookups(t, NewConcurrentMapFromMap(map[interface{}]interface{}{
		"x":                      "x1val",
		"xx":                     "x2val",
		"foo":                    "fooval",
		"xxxx":                   "x4val",
		"xxxxx":                  "x5val",
		"xxxxxx":                 "x6val",
		strings.Repeat("x", 128): "longval",
	}))
}

func testMapLookups(t *testing.T, m *ConcurrentMap) {
	itr := NewHashIterator(m)
	for itr.HasNext() {
		entry := itr.NextEntry()
		k := entry.key.(string)
		v := entry.value.(string)
		//for k, v := range m {
		if v1, err := m.Get(k); v1 != v || err != nil {
			t.Fatalf("m[%q] = %q; want %q", k, v1, v)
		}
	}
}

//// Tests whether the iterator returns the right elements when
//// started in the middle of a grow, when the keys are NaNs.
//func TestMapNanGrowIterator(t *testing.T) {
//	m := make(map[float64]int)
//	nan := math.NaN()
//	const nBuckets = 16
//	// To fill nBuckets buckets takes LOAD * nBuckets keys.
//	nKeys := int(nBuckets * *runtime.HashLoad)

//	// Get map to full point with nan keys.
//	for i := 0; i < nKeys; i++ {
//		m[nan] = i
//	}
//	// Trigger grow
//	m[1.0] = 1
//	delete(m, 1.0)

//	// Run iterator
//	found := make(map[int]struct{})
//	for _, v := range m {
//		if v != -1 {
//			if _, repeat := found[v]; repeat {
//				t.Fatalf("repeat of value %d", v)
//			}
//			found[v] = struct{}{}
//		}
//		if len(found) == nKeys/2 {
//			// Halfway through iteration, finish grow.
//			for i := 0; i < nBuckets; i++ {
//				delete(m, 1.0)
//			}
//		}
//	}
//	if len(found) != nKeys {
//		t.Fatalf("missing value")
//	}
//}

func TestMapIterOrder(t *testing.T) {
	for _, n := range [...]int{3, 7, 9, 15} {
		// Make m be {0: true, 1: true, ..., n-1: true}.
		m := make(map[int]bool)
		for i := 0; i < n; i++ {
			m[i] = true
		}
		// Check that iterating over the map produces at least two different orderings.
		ord := func() []int {
			var s []int
			for key := range m {
				s = append(s, key)
			}
			return s
		}
		first := ord()
		ok := false
		for try := 0; try < 100; try++ {
			if !reflect.DeepEqual(first, ord()) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("Map with n=%d elements had consistent iteration order: %v", n, first)
		}
	}
}

//func TestMapStringBytesLookup(t *testing.T) {
//	// Use large string keys to avoid small-allocation coalescing,
//	// which can cause AllocsPerRun to report lower counts than it should.
//	m0 := map[string]int{
//		"1000000000000000000000000000000000000000000000000": 1,
//		"2000000000000000000000000000000000000000000000000": 2,
//	}
//	m1 := map[interface{}]interface{}{
//		"1000000000000000000000000000000000000000000000000": 1,
//		"2000000000000000000000000000000000000000000000000": 2,
//	}
//	_ = m1
//	m := NewConcurrentMapFromMap(m1)
//	buf := []byte("1000000000000000000000000000000000000000000000000")
//	if x, err := m.Get(string(buf)); x != 1 || err != nil { // m[string(buf)]; x != 1 {
//		t.Errorf(`m[string([]byte("1"))] = %d, want 1`, x)
//	}
//	buf[0] = '2'
//	if x, err := m.Get(string(buf)); x != 2 || err != nil { //x := m[string(buf)]; x != 2 {
//		t.Errorf(`m[string([]byte("2"))] = %d, want 2`, x)
//	}

//	var x int
//	n := testing.AllocsPerRun(100, func() {
//		_, _ = m.Get(string(buf))
//		//_ = m0[string(buf)]   //n will be 0
//		//_ = m1[string(buf)]   //n will be 2
//		//x += v.(int) //m[string(buf)]
//	})
//	if n != 0 {
//		t.Errorf("AllocsPerRun for m[string(buf)] = %v, want 0", n)
//	}

//	x = 0
//	n = testing.AllocsPerRun(100, func() {
//		y, err := m.Get(string(buf))
//		//y, ok := m[string(buf)]
//		if err != nil {
//			panic("!ok")
//		}
//		x += y.(int)
//	})
//	if n != 0 {
//		t.Errorf("AllocsPerRun for x,ok = m[string(buf)] = %v, want 0", n)
//	}
//}

/*-------------test each type of key------------------------*/

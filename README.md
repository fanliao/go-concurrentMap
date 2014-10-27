go-concurrentMap
================

go-concurrentMap is a concurrent Map implement, it is copying from java.util.ConcurrentHashMap.

## Quick start

```go
m := concurrent.NewConcurrentMap()

previou, err := m.Put(1, 10)                   //return nil, nil
previou, err = m.PutIfAbsent(1, 20)            //return 10, nil

val, err := m.Get(1)                           //return 10, nil
s := m.Size()                                  //return 1

m.PutAll(map[interface{}]interface{}{
	1: 100,
	2: 200,
})

ok, err := m.RemoveEntry(1, 100)               //return true, nil

previou, err = m.Replace(2, 20)                //return 200, nil
ok, err = m.CompareAndReplace(2, 200, 20)      //return false, nil

previou, err = m.Remove(2)                     //return 20, nil

m.Clear()
s = m.Size()                                   //return 0

//iterate ConcurrentMap
itr := m.Iterator()
for itr.HasNext() {
	entry := itr.NextEntry()
	k, v := entry.Key(), entry.Value()
}
```

## Performance

The CPU information for performance testing: Xeon E3-1230V3 3.30GHZ

Max number of procs is 8，number of goroutines is 9，every goroutines will put or get 100,000 key-value pairs.

I used LockMap to compare the performance, it is a implement that uses the RWMutex to synchronize. The below are the test results:

* LockMap Put ------------------------- 386.822120 ms/op 

* ConcurrentMap Put ------------------ 99.955715 ms/op

* LockMap Get -------------------------- 45.642610 ms/op 

* ConcurrentMap Get ------------------ 60.583466 ms/op

* LockMap PutAndGet---------------- 620.546950 ms/op 

* ConcurrentMap PutAndGet ------- 138.507920 ms/op

## License

go-concurrentMap is licensed under the MIT Licence, (http://www.apache.org/licenses/LICENSE-2.0.html).

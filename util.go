package concurrent

import (
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"reflect"
	"sync/atomic"
	"unsafe"
)

const (
	intSize   = unsafe.Sizeof(1)
	ptrSize   = unsafe.Sizeof((*int)(nil))
	bigEndian = false
)

var (
	hasherT           = reflect.TypeOf((*Hashable)(nil)).Elem()
	defaultEqualsfunc func(k1 interface{}, k2 interface{}) bool
	hasherEng         *hashEnginer
	boolEng           *hashEnginer
	intEng            *hashEnginer
	int8Eng           *hashEnginer
	int16Eng          *hashEnginer
	int32Eng          *hashEnginer
	int64Eng          *hashEnginer
	uintEng           *hashEnginer
	uint8Eng          *hashEnginer
	uint16Eng         *hashEnginer
	uint32Eng         *hashEnginer
	uint64Eng         *hashEnginer
	uintptrEng        *hashEnginer
	float32Eng        *hashEnginer
	float64Eng        *hashEnginer
	complex64Eng      *hashEnginer
	complex128Eng     *hashEnginer
	stringEng         *hashEnginer
	engMap            map[reflect.Kind]*hashEnginer
)

func init() {
	hasherEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			w.Write(k.(Hashable).HashBytes())
		},
		equals: func(v1, v2 interface{}) bool {
			return v1.(Hashable).Equals(v2)
		},
	}
	boolEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(bool)
			w.Write((*((*[1]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	intEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(int)
			w.Write((*((*[intSize]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	int8Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(int8)
			w.Write((*((*[1]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	int16Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(int16)
			w.Write((*((*[2]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	int32Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(int32)
			w.Write((*((*[4]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	int64Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(int64)
			w.Write((*((*[8]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uintEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uint)
			w.Write((*((*[intSize]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uint8Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uint8)
			w.Write((*((*[1]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uint16Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uint16)
			w.Write((*((*[2]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uint32Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uint32)
			w.Write((*((*[4]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uint64Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uint64)
			w.Write((*((*[8]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	uintptrEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(uintptr)
			w.Write((*((*[intSize]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	float32Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(float32)
			//Nan != Nan, so use a rand number to generate hash code
			if k1 != k1 {
				k1 = rand.Float32()
			}
			w.Write((*((*[4]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	float64Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(float64)
			//Nan != Nan, so use a rand number to generate hash code
			if k1 != k1 {
				k1 = rand.Float64()
			}
			w.Write((*((*[8]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	complex64Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(complex64)
			w.Write((*((*[8]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	complex128Eng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(complex128)
			w.Write((*((*[128]byte)(unsafe.Pointer(&k1))))[:])
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	stringEng = &hashEnginer{
		hash: func(w io.Writer, k interface{}) {
			k1 := k.(string)
			w.Write([]byte(k1))
		},
		equals: func(v1, v2 interface{}) bool {
			return v1 == v2
		},
	}
	engMap = map[reflect.Kind]*hashEnginer{
		reflect.Bool:       boolEng,
		reflect.Int:        intEng,
		reflect.Int8:       int8Eng,
		reflect.Int16:      int16Eng,
		reflect.Int32:      int32Eng,
		reflect.Int64:      int64Eng,
		reflect.Uint:       uintEng,
		reflect.Uint8:      uint8Eng,
		reflect.Uint16:     uint16Eng,
		reflect.Uint32:     uint32Eng,
		reflect.Uint64:     uint64Eng,
		reflect.Uintptr:    uintptrEng,
		reflect.Float32:    float32Eng,
		reflect.Float64:    float64Eng,
		reflect.Complex64:  complex64Eng,
		reflect.Complex128: complex128Eng,
		reflect.String:     stringEng,
	}
}

func hashKey(key interface{}, m *ConcurrentMap, isRead bool) (hashCode uint32, err error) {
	h := fnv.New32a()

	switch v := key.(type) {
	case bool:
		h.Write((*((*[1]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case int:
		h.Write((*((*[intSize]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case int8:
		h.Write((*((*[1]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case int16:
		h.Write((*((*[2]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case int32:
		h.Write((*((*[4]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case int64:
		h.Write((*((*[8]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint:
		h.Write((*((*[intSize]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint8:
		h.Write((*((*[1]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint16:
		h.Write((*((*[2]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint32:
		h.Write((*((*[4]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint64:
		h.Write((*((*[8]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uintptr:
		h.Write((*((*[intSize]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case float32:
		//Nan != Nan, so use a rand number to generate hash code
		if v != v {
			v = rand.Float32()
		}
		h.Write((*((*[4]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case float64:
		//Nan != Nan, so use a rand number to generate hash code
		if v != v {
			v = rand.Float64()
		}
		h.Write((*((*[8]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case complex64:
		h.Write((*((*[8]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case complex128:
		h.Write((*((*[128]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case string:
		h.Write([]byte(v))
		hashCode = h.Sum32()
	default:
		//if key is not simple type
		if her, ok := key.(Hashable); ok {
			h.Write(her.HashBytes())
		} else {
			if err = m.parseKey(key); err != nil {
				return
			}
			if isRead {
				eng := (*hashEnginer)(atomic.LoadPointer(&m.eng))
				eng.hash(h, key)
			} else {
				eng := (*hashEnginer)(m.eng)
				eng.hash(h, key)
			}
			hashCode = h.Sum32()
		}
	}
	return
}

func equals(k1, k2 interface{}, m *ConcurrentMap, isRead bool) bool {
	//if h1, ok := k1.(Hashable); ok {
	//	return h1.Equals(k2)
	//} else {
	//	return k1 == k2
	//}

	switch v := k1.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		uintptr, float32, float64, complex64, complex128, string:
		_ = v
		return k1 == k2
	default:
		//if key is not simple type
		if h1, ok := k1.(Hashable); ok {
			return h1.Equals(k2)
		} else {
			if err := m.parseKey(k1); err != nil {
				return false
			}
			if isRead {
				eng := (*hashEnginer)(atomic.LoadPointer(&m.eng))
				return eng.equals(k1, k2)
			} else {
				eng := (*hashEnginer)(m.eng)
				return eng.equals(k1, k2)
			}
		}
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	k := rv.Type().Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

type keyInfo struct {
	isHasher      bool
	includeHasher bool
	/*-- kind of key type --*/
	kind reflect.Kind
	/*-- index of field if it is a field of struct --*/
	index []int
	/*-- field informations of struct --*/
	fields []*keyInfo
	/*-- element information of array --*/
	elementInfo *keyInfo
	size        int
}

//获取t对应的类型信息，不支持slice, function, map, pointer, interface, channel
func getKeyInfo(t reflect.Type) (ki *keyInfo, err error) {
	return getKeyInfoByParent(t, nil, make([]int, 0, 0))
}

//获取t对应的类型信息，不支持slice, function, map, pointer, interface, channel
//如果parentIdx的长度>0，则表示t是strut中的字段的类型信息
func getKeyInfoByParent(t reflect.Type, parent *keyInfo, parentIdx []int) (ki *keyInfo, err error) {
	ki = &keyInfo{}
	//判断是否实现了hasher接口
	if t.Implements(hasherT) {
		ki.isHasher = true
		return
	}
	ki.kind = t.Kind()

	if _, ok := engMap[ki.kind]; ok {
		//simple types
		ki.index = parentIdx
	} else {
		//some types can be used as key
		switch ki.kind {
		case reflect.Chan, reflect.Slice, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface:
			err = NonSupportKey
		case reflect.Struct:
			if parent == nil {
				//parent==nil present t is a top struct
				parent = ki
				ki.fields = make([]*keyInfo, 0, t.NumField())
			}

			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				//skip unexported field,
				if len(f.PkgPath) > 0 {
					continue
				}

				idx := make([]int, len(parentIdx), len(parentIdx)+1)
				copy(idx, parentIdx)
				idx = append(idx, i)
				if fi, e := getKeyInfoByParent(f.Type, parent, idx); e != nil {
					err = e
					return
				} else {
					//fi.index = i
					if fi.includeHasher || fi.isHasher {
						parent.includeHasher = true
					}
					parent.fields = append(ki.fields, fi)
				}
			}
		case reflect.Array:
			if ki.elementInfo, err = getKeyInfo(t.Elem()); err != nil {
				return
			}
			ki.size = t.Len()
			ki.index = parentIdx
		}
	}

	return

}

func getHashFunc(ki *keyInfo) func(w io.Writer, k interface{}) {
	if ki.isHasher {
		return hasherEng.hash
	}

	//Printf("getPutFunc, ki = %v\n", ki)
	if eng, ok := engMap[ki.kind]; ok {
		return eng.hash
	} else {
		if ki.kind == reflect.Struct {
			//Printf("getPutFunc, ki = %v, other case\n", ki)
			hashFunc := func(w io.Writer, k interface{}) {
				rv := reflect.ValueOf(k)
				for _, fieldInfo := range ki.fields {
					//深度遍历每个field，并将其[]byte写入hash函数
					hashF := getHashFunc(fieldInfo)
					//Printf("getPutFunc, ki = %#v, fieldInfo = %#v, %#v\n", ki, fieldInfo, rv.Interface())
					//Printf("getPutFunc, value = %v\n", rv.FieldByIndex(fieldInfo.index).Interface())
					hashF(w, rv.FieldByIndex(fieldInfo.index).Interface())
				}
			}
			//Printf("getPutFunc, ki=%v, putFunc = %v, other case\n", ki, putFunc)
			return hashFunc
		} else if ki.kind == reflect.Array {
			hashFunc := func(w io.Writer, k interface{}) {
				rv := reflect.ValueOf(k)
				hashF := getHashFunc(ki.elementInfo)
				for i := 0; i < ki.size; i++ {
					//遍历数组元素，并将其[]byte写入hash函数
					hashF(w, rv.Index(i).Interface())
				}
			}
			return hashFunc
		}
		Printf("getPutFunc, return nil")
	}
	return nil
}

func getEqualsFunc(ki *keyInfo) func(v1, v2 interface{}) bool {
	if ki.isHasher {
		return hasherEng.equals
	}

	//Printf("getPutFunc, ki = %v\n", ki)
	if eng, ok := engMap[ki.kind]; ok {
		return eng.equals
	} else {
		if ki.kind == reflect.Struct {
			if !ki.includeHasher {
				return func(v1, v2 interface{}) bool {
					return v1 == v2
				}
			}

			//Printf("getPutFunc, ki = %v, other case\n", ki)
			equalsFunc := func(v1, v2 interface{}) bool {
				rv1, rv2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
				for _, fieldInfo := range ki.fields {
					//深度遍历每个field，比较每个field的值
					equalsF := getEqualsFunc(fieldInfo)
					//Printf("getPutFunc, value = %v\n", rv.FieldByIndex(fieldInfo.index).Interface())
					if !equalsF(rv1.FieldByIndex(fieldInfo.index).Interface(),
						rv2.FieldByIndex(fieldInfo.index).Interface()) {
						return false
					}
				}
				return true
			}
			//Printf("getPutFunc, ki=%v, putFunc = %v, other case\n", ki, putFunc)
			return equalsFunc
		} else if ki.kind == reflect.Array {
			if !(ki.elementInfo.isHasher || ki.elementInfo.includeHasher) {
				return func(v1, v2 interface{}) bool {
					return v1 == v2
				}
			}

			equalsFunc := func(v1, v2 interface{}) bool {
				rv1, rv2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
				equalsF := getEqualsFunc(ki.elementInfo)
				for i := 0; i < ki.size; i++ {
					//比较每个数组元素
					if !equalsF(rv1.Index(i).Interface(), rv2.Index(i).Interface()) {
						return false
					}
				}
				return true
			}
			return equalsFunc
		}
		Printf("getequalsFunc, return nil")
	}
	return nil
}

func Printf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		return fmt.Printf(format, a...)
	} else {
		return 0, nil
	}
}

func Println(a ...interface{}) (n int, err error) {
	if Debug {
		return fmt.Println(a...)
	} else {
		return 0, nil
	}
}

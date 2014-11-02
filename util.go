package concurrent

import (
	"hash"
	"hash/fnv"
	"math/rand"
	"reflect"
	"unsafe"
)

const (
	intSize   = unsafe.Sizeof(1)
	ptrSize   = unsafe.Sizeof((*int)(nil))
	bigEndian = false
)

/**
 * Applies a supplemental hash function to a given hashCode, which
 * defends against poor quality hash functions.  This is critical
 * because ConcurrentHashMap uses power-of-two length hash tables,
 * that otherwise encounter collisions for hashCodes that do not
 * differ in lower or upper bits.
 */
func hash2(h uint32) uint32 {
	//// Spread bits to regularize both segment and index locations,
	//// using variant of single-word Wang/Jenkins hash.
	//h += (h << 15) ^ 0xffffcd7d
	//h ^= (h >> 10)
	//h += (h << 3)
	//h ^= (h >> 6)
	//h += (h << 2) + (h << 14)
	//return uint32(h ^ (h >> 16))

	//Now all hashcode is created by FNVa, it isn't a poor quality hash function
	//so I removes the hash operation for second time
	return h
}

//hash a interface using FNVa
func hashI(val interface{}) (hashCode uint32) {
	h := fnv.New32a()
	switch v := val.(type) {
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
		h.Write((*((*[1]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case uint8:
		h.Write((*((*[intSize]byte)(unsafe.Pointer(&v))))[:])
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
		//some types can be used as key, we can use equals to test
		_ = val == val

		//support array, struct, channel, interface, pointer
		//don't support slice, function, map
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Ptr:
			//ei.word stores the memory address of value that v points to, we use address to generate hash code
			ei := (*emptyInterface)(unsafe.Pointer(&val))
			hashCode = hashI(uintptr(ei.word))
		case reflect.Interface:
			//for interface, we use contained value to generate the hash code
			hashCode = hashI(rv.Elem())
		default:
			//for array, struct and chan, will get byte array to calculate the hash code
			hashMem(rv, h)
			hashCode = h.Sum32()
		}
	}
	return
}

//hashMem writes byte array of underlying value to hash function
func hashMem(i interface{}, hashFunc hash.Hash32) {
	size := reflect.ValueOf(i).Type().Size()
	ei := (*emptyInterface)(unsafe.Pointer(&i))

	//if size of underlying value is greater than pointer size, ei.word will store the pointer that point to underlying value
	//else ei.word will store underlying value
	if size > ptrSize {
		addr := ei.word
		hashPtrData(uintptr(addr), size, hashFunc)
	} else {
		data := ei.word
		hashData(uintptr(data), size, hashFunc)
	}
	return
}

func hashPtrData(ptr uintptr, size uintptr, hashFunc hash.Hash32) {
	idx := 0
	for {
		if size >= 32 {
			bytes := *(*[32]byte)(unsafe.Pointer(ptr))
			size -= 32
			ptr += 32
			idx += 32
			hashFunc.Write(bytes[:])
		} else if size >= 16 {
			bytes := *(*[16]byte)(unsafe.Pointer(ptr))
			size -= 16
			ptr += 16
			idx += 16
			hashFunc.Write(bytes[:])
		} else if size >= 8 {
			bytes := *(*[8]byte)(unsafe.Pointer(ptr))
			size -= 8
			ptr += 8
			idx += 8
			hashFunc.Write(bytes[:])
		} else if size >= 4 {
			bytes := *(*[4]byte)(unsafe.Pointer(ptr))
			size -= 4
			ptr += 4
			idx += 4
			hashFunc.Write(bytes[:])
		} else if size >= 2 {
			bytes := *(*[2]byte)(unsafe.Pointer(ptr))
			size -= 2
			ptr += 2
			idx += 2
			hashFunc.Write(bytes[:])
		} else if size == 1 {
			bytes := *(*[1]byte)(unsafe.Pointer(ptr))
			hashFunc.Write(bytes[:])
			return
		}
		if size == 0 {
			return
		}
	}
}

func hashData(data uintptr, size uintptr, hashFunc hash.Hash32) {
	bytes := (*((*[ptrSize]byte)(unsafe.Pointer(&data))))
	hashFunc.Write(bytes[0:size])
	return
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

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  uintptr
	word unsafe.Pointer
}

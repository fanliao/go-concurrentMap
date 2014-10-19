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

func hashi(val interface{}) (hashCode uint32) {
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
		if v != v { //Nan
			v = rand.Float32()
		}
		h.Write((*((*[4]byte)(unsafe.Pointer(&v))))[:])
		hashCode = h.Sum32()
	case float64:
		if v != v { //Nan
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
		//Will panic if type of interface{} cannot used as hash key
		_ = val == val

		//array, struct, channel, interface, pointer
		//don't support slice, function, map
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Ptr:
			ei := (*emptyInterface)(unsafe.Pointer(&val))
			hashCode = hashi(uintptr(ei.word))
		case reflect.Interface:
			hashCode = hashi(rv.Elem())
		case reflect.Array, reflect.Struct:
			hashMem(rv, h)
			hashCode = h.Sum32()
		case reflect.Chan:
			hashMem(rv, h)
		}
	}
	return
}

func hashMem(i interface{}, hashFunc hash.Hash32) {
	size := reflect.ValueOf(i).Type().Size()
	ei := (*emptyInterface)(unsafe.Pointer(&i))
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
		if size > 16 {
			bytes := *(*[16]byte)(unsafe.Pointer(ptr))
			size -= 16
			ptr += 16
			idx += 16
			hashFunc.Write(bytes[:])
		} else if size > 8 {
			bytes := *(*[8]byte)(unsafe.Pointer(ptr))
			size -= 8
			ptr += 8
			idx += 8
			hashFunc.Write(bytes[:])
		} else if size > 4 {
			bytes := *(*[4]byte)(unsafe.Pointer(ptr))
			size -= 4
			ptr += 4
			idx += 4
			hashFunc.Write(bytes[:])
		} else if size > 2 {
			bytes := *(*[2]byte)(unsafe.Pointer(ptr))
			size -= 2
			ptr += 2
			idx += 2
			hashFunc.Write(bytes[:])
		} else if size == 2 {
			bytes := *(*[2]byte)(unsafe.Pointer(ptr))
			hashFunc.Write(bytes[:])
			return
		} else if size == 1 {
			bytes := *(*[1]byte)(unsafe.Pointer(ptr))
			hashFunc.Write(bytes[:])
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

//package main

//import(
//	"fmt"
//	"unsafe"
//)

//func main(){
//	var f64 float64 = 99999999.999
//	var f32 float32 = 999991.1
//	var r rune = 'a'
//	b2 := 0x12345678
//	fmt.Println(*((*uint64)(unsafe.Pointer(&f32))), uint64(*((*uint32)(unsafe.Pointer(&f32)))))
//	fmt.Println(*((*uint32)(unsafe.Pointer(&f64))), *((*uint64)(unsafe.Pointer(&f64))), uint64(r))
//  //*((*[32]byte)(unsafe.Pointer(&b2)))这个指针转换是否不安全，b2只有4个字节，后面的28个字节貌似不安全？
//	fmt.Println(uint32(b2), *((*uint64)(unsafe.Pointer(&b2))), *((*[32]byte)(unsafe.Pointer(&b2))), *((*[4]byte)(unsafe.Pointer(&b2))))
//	fmt.Println(0x12, 0x34, 0x56, 0x78)
//	fmt.Println("Hello World")
//}

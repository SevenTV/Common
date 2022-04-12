package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"reflect"
	"time"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

//
// Util - Ternary:
// A golang equivalent to JS Ternary Operator
//
// It takes a condition, and returns a result depending on the outcome
//
func Ternary[T any](condition bool, whenTrue T, whenFalse T) T {
	if condition {
		return whenTrue
	}

	return whenFalse
}

//
// Util - Is Power Of Two
//
func IsPowerOfTwo(n int64) bool {
	return (n != 0) && ((n & (n - 1)) == 0)
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func B2S(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func S2B(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func DifferentArray[T ComparableType](a []T, b []T) bool {
	if len(a) != len(b) {
		return true
	}
	if len(a) == 0 {
		return false
	}
	aM := make(map[T]int)
	bM := make(map[T]int)
	for _, v := range a {
		aM[v] = 1
	}
	for _, v := range b {
		bM[v] = 1
		if _, ok := aM[v]; !ok {
			return true
		}
	}
	for k := range aM {
		if _, ok := bM[k]; !ok {
			return true
		}
	}
	return false
}

func IsSliceArray(v interface{}) bool {
	k := reflect.TypeOf(v).Kind()
	return k == reflect.Slice || k == reflect.Array
}

func IsSliceArrayPointer(v interface{}) bool {
	n := reflect.TypeOf(v)
	k := n.Kind()
	if k == reflect.Ptr {
		k = n.Elem().Kind()
		return k == reflect.Slice || k == reflect.Array
	}
	return false
}

func SliceIndexOf[T ComparableType](arr []T, val T) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}

	return -1
}

func Contains[T ComparableType](arr []T, val T) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func PrependSlice[T any](s []T, v T) []T {
	s = make([]T, len(s)+1)
	copy(s[1:], s)
	s[0] = v
	return s
}

func IsPointer(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Ptr
}

func PointerOf[T any](v T) *T {
	return &v
}

type Key string

func PanicHandler(handle func(err interface{})) {
	if err := recover(); err != nil {
		if handle != nil {
			handle(err)
		}
	}
}

func EmptyChannel[T any](ch chan T) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

func JitterTime(lower, upper time.Duration) time.Duration {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(upper-lower)))

	return time.Duration(n.Int64()) + lower
}

type ComparableType interface {
	*any | int | int8 | int16 | int32 | float32 | float64 | string | bool | chan any | primitive.ObjectID
}

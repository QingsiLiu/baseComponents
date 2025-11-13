// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package sets

import (
	"reflect"
	"sort"
)

// Empty is public since it is used by some internal API objects for conversions between external
// string arrays and internal sets, and conversion logic requires public types today.
type Empty struct{}

// ordered is a replacement for constraints.Ordered so we can keep the dependency surface minimal.
// It represents values that can be compared with < and > operators.
type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Set is a generic set implemented via map[T]Empty for minimal memory consumption.
type Set[T ordered] map[T]Empty

// New creates a Set from a list of values.
func New[T ordered](items ...T) Set[T] {
	ss := make(Set[T], len(items))
	ss.Insert(items...)
	return ss
}

// KeySet creates a Set from the keys of the provided map.
// If the value passed in is not actually a map, or if its keys cannot be asserted to T, this panics.
func KeySet[T ordered](theMap interface{}) Set[T] {
	v := reflect.ValueOf(theMap)
	if !v.IsValid() || v.Kind() != reflect.Map {
		panic("sets: KeySet requires a map input")
	}
	ret := make(Set[T], v.Len())
	for _, keyValue := range v.MapKeys() {
		key, ok := keyValue.Interface().(T)
		if !ok {
			panic("sets: map key type doesn't match set element type")
		}
		ret.Insert(key)
	}
	return ret
}

// Insert adds items to the set.
func (s Set[T]) Insert(items ...T) Set[T] {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

// Delete removes all items from the set.
func (s Set[T]) Delete(items ...T) Set[T] {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has returns true if and only if item is contained in the set.
func (s Set[T]) Has(item T) bool {
	_, contained := s[item]
	return contained
}

// HasAll returns true if and only if all items are contained in the set.
func (s Set[T]) HasAll(items ...T) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny returns true if any items are contained in the set.
func (s Set[T]) HasAny(items ...T) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Difference returns a set of objects that are not in s2.
func (s Set[T]) Difference(s2 Set[T]) Set[T] {
	result := New[T]()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union returns a new set which includes items in either s1 or s2.
func (s1 Set[T]) Union(s2 Set[T]) Set[T] {
	result := New[T]()
	for key := range s1 {
		result.Insert(key)
	}
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

// Intersection returns a new set which includes the items that are in BOTH s1 and s2.
func (s1 Set[T]) Intersection(s2 Set[T]) Set[T] {
	var walk, other Set[T]
	result := New[T]()
	if s1.Len() < s2.Len() {
		walk = s1
		other = s2
	} else {
		walk = s2
		other = s1
	}
	for key := range walk {
		if other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset returns true if and only if s1 is a superset of s2.
func (s1 Set[T]) IsSuperset(s2 Set[T]) bool {
	for item := range s2 {
		if !s1.Has(item) {
			return false
		}
	}
	return true
}

// Equal returns true if and only if s1 is equal (as a set) to s2.
func (s1 Set[T]) Equal(s2 Set[T]) bool {
	return s1.Len() == s2.Len() && s1.IsSuperset(s2)
}

// List returns the contents as a sorted slice.
func (s Set[T]) List() []T {
	res := make([]T, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res
}

// UnsortedList returns the slice with contents in random order.
func (s Set[T]) UnsortedList() []T {
	res := make([]T, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny returns a single element from the set.
func (s Set[T]) PopAny() (T, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue T
	return zeroValue, false
}

// Len returns the size of the set.
func (s Set[T]) Len() int {
	return len(s)
}

// Common aliases to preserve the previous API surface.
type (
	Byte   = Set[byte]
	Int    = Set[int]
	Int32  = Set[int32]
	Int64  = Set[int64]
	String = Set[string]
)

func NewByte(items ...byte) Byte       { return New(items...) }
func NewInt(items ...int) Int          { return New(items...) }
func NewInt32(items ...int32) Int32    { return New(items...) }
func NewInt64(items ...int64) Int64    { return New(items...) }
func NewString(items ...string) String { return New(items...) }

func ByteKeySet(theMap interface{}) Byte     { return KeySet[byte](theMap) }
func IntKeySet(theMap interface{}) Int       { return KeySet[int](theMap) }
func Int32KeySet(theMap interface{}) Int32   { return KeySet[int32](theMap) }
func Int64KeySet(theMap interface{}) Int64   { return KeySet[int64](theMap) }
func StringKeySet(theMap interface{}) String { return KeySet[string](theMap) }

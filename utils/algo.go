package utils

import (
	"sort"
)

const (
	// hashing constants
	FNV_OFFSET64 = 14695981039346656037
	FNV_PRIME64  = 1099511628211
)

// key hash function - avoid standard Hash interface overhead
func FNVSum64(key string) uint64 {
	var hash uint64 = FNV_OFFSET64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= FNV_PRIME64
	}

	return hash
}

// fast Knuth power
func Pow(a, b int) int {
	p := 1
	for b > 0 {
		if b&1 != 0 {
			p *= a
		}
		b >>= 1
		a *= a
	}
	return p
}

// boilerplate for tests
func CompareStringByteMaps(m1, m2 map[string][]byte) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		v2, ok := m2[k]
		if !ok || !CompareByteSlices(v1, v2) {
			return false
		}
	}
	return true
}

func CompareByteSlices(s1, s2 []byte) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func FindInSliceString(p string, s []string) bool {
	for _, e := range s {
		if p == e {
			return true
		}
	}
	return false
}

func CompareStringSlicesUnordered(s1, s2 []string) bool {
	sort.Strings(s1)
	sort.Strings(s2)
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

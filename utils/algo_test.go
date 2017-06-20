package utils

import (
	"testing"
)

type StringByteMap map[string][]byte
type StringByteMapTestCase struct {
	First, Second StringByteMap
	Comp          bool
}

type ByteSliceTestCase struct {
	First, Second []byte
	Comp          bool
}

type StringSliceTestCase struct {
	First, Second []string
	Comp          bool
}

func TestUtilsAlgoPower(t *testing.T) {
	testCases := [][]int{
		[]int{0, 0, 1},
		[]int{0, 1, 0},
		[]int{-1, 0, 1},
		[]int{1, -1, 0},
		[]int{1, 0, 1},
		[]int{1, 300, 1},
		[]int{2, 2, 4},
		[]int{3, 3, 27},
		[]int{2, 10, 1024},
	}

	for _, c := range testCases {
		if Pow(c[0], c[1]) != c[2] {
			t.Errorf("power '%d^%d' failure => expected: %d, get: %d", c[0], c[1], c[2], Pow(c[0], c[1]))
		}
	}
}

func TestUtilsAlgoCompareStringByteMaps(t *testing.T) {
	testCases := []StringByteMapTestCase{
		StringByteMapTestCase{
			First: StringByteMap{
				"k1": []byte("v1"),
				"k2": []byte("v2")},
			Second: StringByteMap{
				"k1": []byte("v1"),
				"k2": []byte("v2")},
			Comp: true},
		StringByteMapTestCase{
			First: StringByteMap{
				"k1": []byte("v1"),
				"k2": []byte("v2")},
			Second: StringByteMap{
				"k2": []byte("v2")},
			Comp: false},
		StringByteMapTestCase{
			First: StringByteMap{
				"k1": []byte("v1"),
				"k2": []byte("v2")},
			Second: StringByteMap{},
			Comp:   false},
		StringByteMapTestCase{
			First: StringByteMap{
				"k1": []byte("v1"),
				"k2": []byte("v2")},
			Second: StringByteMap{
				"k2": []byte("v2"),
				"k1": []byte("v1")},
			Comp: true},
	}

	for _, c := range testCases {
		if CompareStringByteMaps(c.First, c.Second) != c.Comp {
			t.Error("no match")
		}
	}
}

func TestUtilsAlgoCompareByteSlices(t *testing.T) {
	testCases := []ByteSliceTestCase{
		ByteSliceTestCase{
			First:  []byte("qwerty123"),
			Second: []byte("qwerty123"),
			Comp:   true},
		ByteSliceTestCase{
			First:  []byte("qwerty123"),
			Second: []byte("qwery123"),
			Comp:   false},
		ByteSliceTestCase{
			First:  []byte("qwerty123"),
			Second: []byte("qwerdy123"),
			Comp:   false},
		ByteSliceTestCase{
			First:  []byte("qwerty123"),
			Second: []byte("\"qwerty123"),
			Comp:   false},
		ByteSliceTestCase{
			First:  []byte("qwerty123"),
			Second: []byte(""),
			Comp:   false},
	}

	for _, c := range testCases {
		if CompareByteSlices(c.First, c.Second) != c.Comp {
			t.Error("no match")
		}
	}
}

func TestUtilsAlgoCompareStringSlices(t *testing.T) {
	testCases := []StringSliceTestCase{
		StringSliceTestCase{
			First:  []string{"sorrow", "not", "wise", "warrior", "it", "is", "better", "for", "a", "man", "to", "avenge", "his", "friend", "than", "much", "mourn"},
			Second: []string{"sorrow", "not", "wise", "warrior", "it", "is", "better", "for", "a", "man", "to", "avenge", "his", "friend", "than", "much", "mourn"},
			Comp:   true},
		StringSliceTestCase{
			First:  []string{"sorrow", "not", "wise", "warrior", "it", "is", "better", "for", "a", "man", "to", "avenge", "his", "friend", "than", "much", "mourn"},
			Second: []string{"wise", "sorrow", "not", "warrior", "is", "man", "to", "better", "for", "it", "a", "avenge", "his", "much", "mourn", "than", "friend"},
			Comp:   true},
		StringSliceTestCase{
			First:  []string{"sorrow", "not", "wise", "warrior", "it", "is", "better", "for", "a", "man", "to", "avenge", "his", "friend", "than", "much", "mourn"},
			Second: []string{"sorrow", "not", "wise", "warrior", "it", "is", "better", "for", "the", "man", "to", "avenge", "his", "friend", "than", "much", "mourn"},
			Comp:   false},
		StringSliceTestCase{
			First:  []string{"sorrow", "not", "wise", "warrior"},
			Second: []string{"sorrow", "not"},
			Comp:   false},
		StringSliceTestCase{
			First:  []string{"sorrow", "not", "wise", "warrior"},
			Second: []string{},
			Comp:   false},
	}

	for _, c := range testCases {
		if CompareStringSlicesUnordered(c.First, c.Second) != c.Comp {
			t.Error("no match")
		}
	}
}

package trace

import (
	"testing"
)

func TestFormat(t *testing.T) {
	cases := map[string]struct {
		value uint64
		bytes int
	}{
		"1234567890abcdef": {0x1234567890abcdef, 8},
		"34567890abcdef":   {0x1234567890abcdef, 7},
		"567890abcdef":     {0x1234567890abcdef, 6},
		"7890abcdef":       {0x1234567890abcdef, 5},
		"90abcdef":         {0x1234567890abcdef, 4},
		"abcdef":           {0x1234567890abcdef, 3},
		"cdef":             {0x1234567890abcdef, 2},
		"ef":               {0x1234567890abcdef, 1},
	}

	for expected, args := range cases {
		dst := make([]byte, args.bytes*2)
		remainder := format(dst, args.value, args.bytes)

		if len(remainder) != 0 {
			t.Fatalf("remainder should be empty. [remainder:%v]", remainder)
		}

		if actual := makeString(dst); actual != expected {
			t.Fatalf("invalid format result. [expected:%v] [actual:%v]", expected, actual)
		}
	}

	if expected, actual := "1234567890abcdef", hexString(0x1234567890abcdef); expected != actual {
		t.Fatalf("invalid hexString result. [expected:%v] [actual:%v]", expected, actual)
	}
}

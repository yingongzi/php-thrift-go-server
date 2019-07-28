package trace

import (
	"testing"
	"time"
)

func BenchmarkNewLogid(b *testing.B) {
	unixnano := time.Now().UnixNano()

	for i := 0; i < b.N; i++ {
		MakeLogid(unixnano)
	}
}

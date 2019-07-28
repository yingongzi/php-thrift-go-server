package trace

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkSpanid(b *testing.B) {
	unixnano := time.Now().UnixNano()
	rnd := rand.New(rand.NewSource(unixnano))

	for i := 0; i < b.N; i++ {
		makeSpanid(unixnano, rnd.Int63())
	}
}

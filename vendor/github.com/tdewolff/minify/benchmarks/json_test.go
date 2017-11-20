package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/json"
)

var jsonSamples = []string{
	"sample_large.json",
	"sample_testsuite.json",
	"sample_twitter.json",
}

func init() {
	for _, sample := range jsonSamples {
		load(sample)
	}
}

func BenchmarkJSON(b *testing.B) {
	for _, sample := range jsonSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				json.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}

package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/js"
)

var jsSamples = []string{
	"sample_ace.js",
	"sample_dot.js",
	"sample_jquery.js",
	"sample_jqueryui.js",
	"sample_moment.js",
}

func init() {
	for _, sample := range jsSamples {
		load(sample)
	}
}

func BenchmarkJS(b *testing.B) {
	for _, sample := range jsSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				js.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}

package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/html"
)

var htmlSamples = []string{
	"sample_amazon.html",
	"sample_bbc.html",
	"sample_blogpost.html",
	"sample_es6.html",
	"sample_stackoverflow.html",
	"sample_wikipedia.html",
}

func init() {
	for _, sample := range htmlSamples {
		load(sample)
	}
}

func BenchmarkHTML(b *testing.B) {
	for _, sample := range htmlSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				html.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}

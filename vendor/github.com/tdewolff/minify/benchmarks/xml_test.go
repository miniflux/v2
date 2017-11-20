package benchmarks

import (
	"testing"

	"github.com/tdewolff/minify/xml"
)

var xmlSamples = []string{
	"sample_books.xml",
	"sample_catalog.xml",
	"sample_omg.xml",
}

func init() {
	for _, sample := range xmlSamples {
		load(sample)
	}
}

func BenchmarkXML(b *testing.B) {
	for _, sample := range xmlSamples {
		b.Run(sample, func(b *testing.B) {
			b.SetBytes(int64(r[sample].Len()))

			for i := 0; i < b.N; i++ {
				r[sample].Reset()
				w[sample].Reset()
				xml.Minify(m, w[sample], r[sample], nil)
			}
		})
	}
}

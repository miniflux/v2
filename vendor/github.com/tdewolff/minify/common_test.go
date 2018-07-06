package minify // import "github.com/tdewolff/minify"

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/tdewolff/test"
)

func TestMediatype(t *testing.T) {
	mediatypeTests := []struct {
		mediatype string
		expected  string
	}{
		{"text/html", "text/html"},
		{"text/html; charset=UTF-8", "text/html;charset=utf-8"},
		{"text/html; charset=UTF-8 ; param = \" ; \"", "text/html;charset=utf-8;param=\" ; \""},
		{"text/html, text/css", "text/html,text/css"},
	}
	for _, tt := range mediatypeTests {
		t.Run(tt.mediatype, func(t *testing.T) {
			mediatype := Mediatype([]byte(tt.mediatype))
			test.Minify(t, tt.mediatype, nil, string(mediatype), tt.expected)
		})
	}
}

func TestDataURI(t *testing.T) {
	dataURITests := []struct {
		dataURI  string
		expected string
	}{
		{"data:,text", "data:,text"},
		{"data:text/plain;charset=us-ascii,text", "data:,text"},
		{"data:TEXT/PLAIN;CHARSET=US-ASCII,text", "data:,text"},
		{"data:text/plain;charset=us-asciiz,text", "data:;charset=us-asciiz,text"},
		{"data:;base64,dGV4dA==", "data:,text"},
		{"data:text/svg+xml;base64,PT09PT09", "data:text/svg+xml;base64,PT09PT09"},
		{"data:text/xml;version=2.0,content", "data:text/xml;version=2.0,content"},
		{"data:text/xml; version = 2.0,content", "data:text/xml;version=2.0,content"},
		{"data:,=====", "data:,%3D%3D%3D%3D%3D"},
		{"data:,======", "data:;base64,PT09PT09"},
		{"data:text/x,<?x?>", "data:text/x,%3C%3Fx%3F%3E"},
	}
	m := New()
	m.AddFunc("text/x", func(_ *M, w io.Writer, r io.Reader, _ map[string]string) error {
		b, _ := ioutil.ReadAll(r)
		test.String(t, string(b), "<?x?>")
		w.Write(b)
		return nil
	})
	for _, tt := range dataURITests {
		t.Run(tt.dataURI, func(t *testing.T) {
			dataURI := DataURI(m, []byte(tt.dataURI))
			test.Minify(t, tt.dataURI, nil, string(dataURI), tt.expected)
		})
	}
}

func TestDecimal(t *testing.T) {
	numberTests := []struct {
		number   string
		expected string
	}{
		{"0", "0"},
		{".0", "0"},
		{"1.0", "1"},
		{"0.1", ".1"},
		{"+1", "1"},
		{"-1", "-1"},
		{"-0.1", "-.1"},
		{"10", "10"},
		{"100", "100"},
		{"1000", "1000"},
		{"0.001", ".001"},
		{"0.0001", ".0001"},
		{"0.252", ".252"},
		{"1.252", "1.252"},
		{"-1.252", "-1.252"},
		{"0.075", ".075"},
		{"789012345678901234567890123456789e9234567890123456789", "789012345678901234567890123456789e9234567890123456789"},
		{".000100009", ".000100009"},
		{".0001000009", ".0001000009"},
		{".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009", ".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"},
		{"E\x1f", "E\x1f"}, // fuzz
	}
	for _, tt := range numberTests {
		t.Run(tt.number, func(t *testing.T) {
			number := Decimal([]byte(tt.number), -1)
			test.Minify(t, tt.number, nil, string(number), tt.expected)
		})
	}
}

func TestDecimalTruncate(t *testing.T) {
	numberTests := []struct {
		number   string
		truncate int
		expected string
	}{
		{"0.1", 1, ".1"},
		{"0.0001", 1, "0"},
		{"0.111", 1, ".1"},
		{"0.111", 0, "0"},
		{"0.075", 1, ".1"},
		{"0.025", 1, "0"},
		{"9.99", 1, "10"},
		{"8.88", 1, "8.9"},
		{"8.88", 0, "9"},
		{"8.00", 0, "8"},
		{".88", 0, "1"},
		{"1.234", 1, "1.2"},
		{"33.33", 0, "33"},
		{"29.666", 0, "30"},
		{"1.51", 1, "1.5"},
		{"1.01", 1, "1"},
	}
	for _, tt := range numberTests {
		t.Run(tt.number, func(t *testing.T) {
			number := Decimal([]byte(tt.number), tt.truncate)
			test.Minify(t, tt.number, nil, string(number), tt.expected, "truncate to", tt.truncate)
		})
	}
}

func TestNumber(t *testing.T) {
	numberTests := []struct {
		number   string
		expected string
	}{
		{"0", "0"},
		{".0", "0"},
		{"1.0", "1"},
		{"0.1", ".1"},
		{"+1", "1"},
		{"-1", "-1"},
		{"-0.1", "-.1"},
		{"10", "10"},
		{"100", "100"},
		{"1000", "1e3"},
		{"0.001", ".001"},
		{"0.0001", "1e-4"},
		{"100e1", "1e3"},
		{"1.1e+1", "11"},
		{"1.1e6", "11e5"},
		{"1.1e", "1.1e"},   // broken number, don't parse
		{"1.1e+", "1.1e+"}, // broken number, don't parse
		{"0.252", ".252"},
		{"1.252", "1.252"},
		{"-1.252", "-1.252"},
		{"0.075", ".075"},
		{"789012345678901234567890123456789e9234567890123456789", "789012345678901234567890123456789e9234567890123456789"},
		{".000100009", "100009e-9"},
		{".0001000009", ".0001000009"},
		{".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009", ".0001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000009"},
		{".6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003e-9", ".6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003e-9"},
		{"E\x1f", "E\x1f"}, // fuzz
		{"1e9223372036854775807", "1e9223372036854775807"},
		{"11e9223372036854775807", "11e9223372036854775807"},
		{".01e-9223372036854775808", ".01e-9223372036854775808"},
		{".011e-9223372036854775808", ".011e-9223372036854775808"},

		{".12345e8", "12345e3"},
		{".12345e7", "1234500"},
		{".12345e6", "123450"},
		{".12345e5", "12345"},
		{".012345e6", "12345"},
		{".12345e4", "1234.5"},
		{"-.12345e4", "-1234.5"},
		{".12345e0", ".12345"},
		{".12345e-1", ".012345"},
		{".12345e-2", ".0012345"},
		{".12345e-3", "12345e-8"},
		{".12345e-4", "12345e-9"},
		{".12345e-5", ".12345e-5"},

		{".123456e-3", "123456e-9"},
		{".123456e-2", ".00123456"},
		{".1234567e-4", ".1234567e-4"},
		{".1234567e-3", ".0001234567"},

		{"12345678e-1", "1234567.8"},
		{"72.e-3", ".072"},
		{"7640e-2", "76.4"},
		{"10.e-3", ".01"},
		{".0319e3", "31.9"},
		{"39.7e-2", ".397"},
		{"39.7e-3", ".0397"},
		{".01e1", ".1"},
		{".001e1", ".01"},
		{"39.7e-5", "397e-6"},
	}
	for _, tt := range numberTests {
		t.Run(tt.number, func(t *testing.T) {
			number := Number([]byte(tt.number), -1)
			test.Minify(t, tt.number, nil, string(number), tt.expected)
		})
	}
}

func TestNumberTruncate(t *testing.T) {
	numberTests := []struct {
		number   string
		truncate int
		expected string
	}{
		{"0.1", 1, ".1"},
		{"0.0001", 1, "1e-4"},
		{"0.111", 1, ".1"},
		{"0.111", 0, "0"},
		{"0.075", 1, ".1"},
		{"0.025", 1, "0"},
		{"9.99", 1, "10"},
		{"8.88", 1, "8.9"},
		{"8.88", 0, "9"},
		{"8.00", 0, "8"},
		{".88", 0, "1"},
		{"1.234", 1, "1.2"},
		{"33.33", 0, "33"},
		{"29.666", 0, "30"},
		{"1.51", 1, "1.5"},
		{"1.01", 1, "1"},
	}
	for _, tt := range numberTests {
		t.Run(tt.number, func(t *testing.T) {
			number := Number([]byte(tt.number), tt.truncate)
			test.Minify(t, tt.number, nil, string(number), tt.expected, "truncate to", tt.truncate)
		})
	}
}

func TestDecimalRandom(t *testing.T) {
	N := int(1e4)
	if testing.Short() {
		N = 0
	}
	for i := 0; i < N; i++ {
		b := RandNumBytes(false)
		f, _ := strconv.ParseFloat(string(b), 64)

		b2 := make([]byte, len(b))
		copy(b2, b)
		b2 = Decimal(b2, -1)
		f2, _ := strconv.ParseFloat(string(b2), 64)
		if math.Abs(f-f2) > 1e-6 {
			fmt.Println("Bad:", f, "!=", f2, "in", string(b), "to", string(b2))
		}
	}
}

func TestNumberRandom(t *testing.T) {
	N := int(1e4)
	if testing.Short() {
		N = 0
	}
	for i := 0; i < N; i++ {
		b := RandNumBytes(true)
		f, _ := strconv.ParseFloat(string(b), 64)

		b2 := make([]byte, len(b))
		copy(b2, b)
		b2 = Number(b2, -1)
		f2, _ := strconv.ParseFloat(string(b2), 64)
		if math.Abs(f-f2) > 1e-6 {
			fmt.Println("Bad:", f, "!=", f2, "in", string(b), "to", string(b2))
		}
	}
}

////////////////

var n = 100
var numbers [][]byte

func TestMain(t *testing.T) {
	numbers = make([][]byte, 0, n)
	for j := 0; j < n; j++ {
		numbers = append(numbers, RandNumBytes(true))
	}
}

func RandNumBytes(withExp bool) []byte {
	var b []byte
	n := rand.Int() % 10
	for i := 0; i < n; i++ {
		b = append(b, byte(rand.Int()%10)+'0')
	}
	if rand.Int()%2 == 0 {
		b = append(b, '.')
		n = rand.Int() % 10
		for i := 0; i < n; i++ {
			b = append(b, byte(rand.Int()%10)+'0')
		}
	}
	if withExp && rand.Int()%2 == 0 {
		b = append(b, 'e')
		if rand.Int()%2 == 0 {
			b = append(b, '-')
		}
		n = 1 + rand.Int()%4
		for i := 0; i < n; i++ {
			b = append(b, byte(rand.Int()%10)+'0')
		}
	}
	return b
}

func BenchmarkNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			Number(numbers[j], -1)
		}
	}
}

func BenchmarkNumber2(b *testing.B) {
	num := []byte("1.2345e-6")
	for i := 0; i < b.N; i++ {
		Number(num, -1)
	}
}

package strconv // import "github.com/tdewolff/parse/strconv"

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/tdewolff/test"
)

func TestParseFloat(t *testing.T) {
	floatTests := []struct {
		f        string
		expected float64
	}{
		{"5", 5},
		{"5.1", 5.1},
		{"-5.1", -5.1},
		{"5.1e-2", 5.1e-2},
		{"5.1e+2", 5.1e+2},
		{"0.0e1", 0.0e1},
		{"18446744073709551620", 18446744073709551620.0},
		{"1e23", 1e23},
		// TODO: hard to test due to float imprecision
		// {"1.7976931348623e+308", 1.7976931348623e+308)
		// {"4.9406564584124e-308", 4.9406564584124e-308)
	}
	for _, tt := range floatTests {
		f, n := ParseFloat([]byte(tt.f))
		test.That(t, n == len(tt.f), "parsed", n, "characters instead for", tt.f)
		test.That(t, f == tt.expected, "return", tt.expected, "for", tt.f)
	}
}

func TestAppendFloat(t *testing.T) {
	floatTests := []struct {
		f        float64
		prec     int
		expected string
	}{
		{0, 6, "0"},
		{1, 6, "1"},
		{9, 6, "9"},
		{9.99999, 6, "9.99999"},
		{123, 6, "123"},
		{0.123456, 6, ".123456"},
		{0.066, 6, ".066"},
		{0.0066, 6, ".0066"},
		{12e2, 6, "1200"},
		{12e3, 6, "12e3"},
		{0.1, 6, ".1"},
		{0.001, 6, ".001"},
		{0.0001, 6, "1e-4"},
		{-1, 6, "-1"},
		{-123, 6, "-123"},
		{-123.456, 6, "-123.456"},
		{-12e3, 6, "-12e3"},
		{-0.1, 6, "-.1"},
		{-0.0001, 6, "-1e-4"},
		{0.000100009, 10, "100009e-9"},
		{0.0001000009, 10, "1.000009e-4"},
		{1e18, 0, "1e18"},
		//{1e19, 0, "1e19"},
		//{1e19, 18, "1e19"},
		{1e1, 0, "10"},
		{1e2, 1, "100"},
		{1e3, 2, "1e3"},
		{1e10, -1, "1e10"},
		{1e15, -1, "1e15"},
		{1e-5, 6, "1e-5"},
		{math.NaN(), 0, ""},
		{math.Inf(1), 0, ""},
		{math.Inf(-1), 0, ""},
		{0, 19, ""},
		{.000923361977200859392, -1, "9.23361977200859392e-4"},
	}
	for _, tt := range floatTests {
		f, _ := AppendFloat([]byte{}, tt.f, tt.prec)
		test.String(t, string(f), tt.expected, "for", tt.f)
	}

	b := make([]byte, 0, 22)
	AppendFloat(b, 12.34, -1)
	test.String(t, string(b[:5]), "12.34", "in buffer")
}

////////////////////////////////////////////////////////////////

func TestAppendFloatRandom(t *testing.T) {
	N := int(1e6)
	if testing.Short() {
		N = 0
	}
	r := rand.New(rand.NewSource(99))
	//prec := 10
	for i := 0; i < N; i++ {
		f := r.ExpFloat64()
		//f = math.Floor(f*float64(prec)) / float64(prec)

		b, _ := AppendFloat([]byte{}, f, -1)
		f2, _ := strconv.ParseFloat(string(b), 64)
		if math.Abs(f-f2) > 1e-6 {
			fmt.Println("Bad:", f, "!=", f2, "in", string(b))
		}
	}
}

func BenchmarkFloatToBytes1(b *testing.B) {
	r := []byte{} //make([]byte, 10)
	f := 123.456
	for i := 0; i < b.N; i++ {
		r = strconv.AppendFloat(r[:0], f, 'g', 6, 64)
	}
}

func BenchmarkFloatToBytes2(b *testing.B) {
	r := make([]byte, 10)
	f := 123.456
	for i := 0; i < b.N; i++ {
		r, _ = AppendFloat(r[:0], f, 6)
	}
}

func BenchmarkModf1(b *testing.B) {
	f := 123.456
	x := 0.0
	for i := 0; i < b.N; i++ {
		a, b := math.Modf(f)
		x += a + b
	}
}

func BenchmarkModf2(b *testing.B) {
	f := 123.456
	x := 0.0
	for i := 0; i < b.N; i++ {
		a := float64(int64(f))
		b := f - a
		x += a + b
	}
}

func BenchmarkPrintInt1(b *testing.B) {
	X := int64(123456789)
	n := LenInt(X)
	r := make([]byte, n)
	for i := 0; i < b.N; i++ {
		x := X
		j := n
		for x > 0 {
			j--
			r[j] = '0' + byte(x%10)
			x /= 10
		}
	}
}

func BenchmarkPrintInt2(b *testing.B) {
	X := int64(123456789)
	n := LenInt(X)
	r := make([]byte, n)
	for i := 0; i < b.N; i++ {
		x := X
		j := n
		for x > 0 {
			j--
			newX := x / 10
			r[j] = '0' + byte(x-10*newX)
			x = newX
		}
	}
}

var int64pow10 = []int64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
}

func BenchmarkPrintInt3(b *testing.B) {
	X := int64(123456789)
	n := LenInt(X)
	r := make([]byte, n)
	for i := 0; i < b.N; i++ {
		x := X
		j := 0
		for j < n {
			pow := int64pow10[n-j-1]
			tmp := x / pow
			r[j] = '0' + byte(tmp)
			j++
			x -= tmp * pow
		}
	}
}

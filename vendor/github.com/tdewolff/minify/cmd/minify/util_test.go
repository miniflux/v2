package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/tdewolff/test"
)

func testOpener(filename string) (io.ReadCloser, error) {
	if filename == "err" {
		return nil, test.ErrPlain
	} else if filename == "empty" {
		return ioutil.NopCloser(test.NewEmptyReader()), nil
	}
	return ioutil.NopCloser(bytes.NewReader([]byte(filename))), nil
}

func TestConcat(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)

	buf, err := ioutil.ReadAll(r)
	test.T(t, err, nil)
	test.Bytes(t, buf, []byte("testtest"))

	n, err := r.Read(buf)
	test.T(t, n, 0)
	test.T(t, err, io.EOF)
}

func TestConcatErr(t *testing.T) {
	r, err := NewConcatFileReader([]string{"err"}, testOpener)
	test.T(t, err, test.ErrPlain)

	r, err = NewConcatFileReader([]string{"test", "err"}, testOpener)
	test.T(t, err, nil)

	buf := make([]byte, 10)
	n, err := r.Read(buf)
	test.T(t, n, 4)
	test.T(t, err, nil)
	test.Bytes(t, buf[:n], []byte("test"))

	n, err = r.Read(buf)
	test.T(t, n, 0)
	test.T(t, err, test.ErrPlain)
}

func TestConcatSep(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("_"))

	buf := make([]byte, 10)
	n, err := r.Read(buf)
	test.T(t, n, 4)
	test.T(t, err, nil)
	test.Bytes(t, buf[:n], []byte("test"))

	n, err = r.Read(buf[n:])
	test.T(t, n, 5)
	test.T(t, err, nil)
	test.Bytes(t, buf[:4+n], []byte("test_test"))
}

func TestConcatSepShort1(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("_"))

	// insufficient room for separator
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	test.T(t, n, 4)
	test.T(t, err, nil)
	test.Bytes(t, buf, []byte("test"))

	n, err = r.Read(buf[4:])
	test.T(t, n, 0)
	test.T(t, err, nil)
}

func TestConcatSepShort2(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("_"))

	// insufficient room after separator
	buf := make([]byte, 5)
	_, _ = r.Read(buf)

	n, err := r.Read(buf[4:])
	test.T(t, n, 1)
	test.T(t, err, nil)
	test.Bytes(t, buf, []byte("test_"))
}

func TestConcatSepShort3(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("_"))

	// insufficient room after separator
	buf := make([]byte, 6)
	_, _ = r.Read(buf)

	n, err := r.Read(buf[4:])
	test.T(t, n, 2)
	test.T(t, err, nil)
	test.Bytes(t, buf, []byte("test_t"))
}

func TestConcatSepShort4(t *testing.T) {
	r, err := NewConcatFileReader([]string{"test", "test"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("xx"))

	// insufficient room after separator
	buf := make([]byte, 5)
	_, _ = r.Read(buf)

	n, err := r.Read(buf[4:])
	test.T(t, n, 1)
	test.T(t, err, nil)
	test.Bytes(t, buf, []byte("testx"))

	n, err = r.Read(buf[5:])
	test.T(t, n, 0)
	test.T(t, err, nil)

	buf2 := make([]byte, 5)
	n, err = r.Read(buf2)
	test.T(t, n, 5)
	test.T(t, err, nil)
	test.Bytes(t, buf2, []byte("xtest"))
}

func TestConcatSepEmpty(t *testing.T) {
	r, err := NewConcatFileReader([]string{"empty", "empty"}, testOpener)
	test.T(t, err, nil)
	r.SetSeparator([]byte("_"))

	// insufficient room after separator
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	test.T(t, n, 1)
	test.T(t, err, io.EOF)
	test.Bytes(t, buf, []byte("_"))
}

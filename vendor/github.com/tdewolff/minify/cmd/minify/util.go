package main

import "io"

type countingReader struct {
	io.Reader
	N int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.N += n
	return n, err
}

type countingWriter struct {
	io.Writer
	N int
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.N += n
	return n, err
}

type prependReader struct {
	io.ReadCloser
	prepend []byte
}

func NewPrependReader(r io.ReadCloser, prepend []byte) *prependReader {
	return &prependReader{r, prepend}
}

func (r *prependReader) Read(p []byte) (int, error) {
	if r.prepend != nil {
		n := copy(p, r.prepend)
		if n != len(r.prepend) {
			return n, io.ErrShortBuffer
		}
		r.prepend = nil
		return n, nil
	}
	return r.ReadCloser.Read(p)
}

package main

import (
	"io"
)

type countingReader struct {
	io.Reader
	N int
}

func NewCountingReader(r io.Reader) *countingReader {
	return &countingReader{r, 0}
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

func NewCountingWriter(w io.Writer) *countingWriter {
	return &countingWriter{w, 0}
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.N += n
	return n, err
}

type eofReader struct{}

func (r eofReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (r eofReader) Close() error {
	return nil
}

type concatFileReader struct {
	filenames []string
	opener    func(string) (io.ReadCloser, error)
	sep       []byte

	cur     io.ReadCloser
	sepLeft int
}

// NewConcatFileReader reads from a list of filenames, and lazily loads files as it needs it.
// It is a reader that reads a concatenation of those files separated by the separator.
// You must call Close to close the last file in the list.
func NewConcatFileReader(filenames []string, opener func(string) (io.ReadCloser, error)) (*concatFileReader, error) {
	var cur io.ReadCloser
	if len(filenames) > 0 {
		var filename string
		filename, filenames = filenames[0], filenames[1:]

		var err error
		if cur, err = opener(filename); err != nil {
			return nil, err
		}
	} else {
		cur = eofReader{}
	}
	return &concatFileReader{filenames, opener, nil, cur, 0}, nil
}

func (r *concatFileReader) SetSeparator(sep []byte) {
	r.sep = sep
}

func (r *concatFileReader) Read(p []byte) (int, error) {
	m := r.writeSep(p)
	n, err := r.cur.Read(p[m:])
	n += m

	// current reader is finished, load in the new reader
	if err == io.EOF && len(r.filenames) > 0 {
		if err := r.cur.Close(); err != nil {
			return n, err
		}

		var filename string
		filename, r.filenames = r.filenames[0], r.filenames[1:]
		if r.cur, err = r.opener(filename); err != nil {
			return n, err
		}
		r.sepLeft = len(r.sep)

		// if previous read returned (0, io.EOF), read from the new reader
		if n == 0 {
			return r.Read(p)
		} else {
			n += r.writeSep(p[n:])
		}
	}
	return n, err
}

func (r *concatFileReader) writeSep(p []byte) int {
	m := 0
	if r.sepLeft > 0 {
		m = copy(p, r.sep[len(r.sep)-r.sepLeft:])
		r.sepLeft -= m
	}
	return m
}

func (r *concatFileReader) Close() error {
	return r.cur.Close()
}

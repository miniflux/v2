package minify // import "github.com/tdewolff/minify"

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/tdewolff/test"
)

var errDummy = errors.New("dummy error")

// from os/exec/exec_test.go
func helperCommand(t *testing.T, s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

////////////////////////////////////////////////////////////////

var m *M

func init() {
	m = New()
	m.AddFunc("dummy/copy", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		io.Copy(w, r)
		return nil
	})
	m.AddFunc("dummy/nil", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return nil
	})
	m.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	m.AddFunc("dummy/charset", func(m *M, w io.Writer, r io.Reader, params map[string]string) error {
		w.Write([]byte(params["charset"]))
		return nil
	})
	m.AddFunc("dummy/params", func(m *M, w io.Writer, r io.Reader, params map[string]string) error {
		return m.Minify(params["type"]+"/"+params["sub"], w, r)
	})
	m.AddFunc("type/sub", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("type/sub"))
		return nil
	})
	m.AddFuncRegexp(regexp.MustCompile("^type/.+$"), func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("type/*"))
		return nil
	})
	m.AddFuncRegexp(regexp.MustCompile("^.+/.+$"), func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		w.Write([]byte("*/*"))
		return nil
	})
}

func TestMinify(t *testing.T) {
	test.T(t, m.Minify("?", nil, nil), ErrNotExist, "minifier doesn't exist")
	test.T(t, m.Minify("dummy/nil", nil, nil), nil)
	test.T(t, m.Minify("dummy/err", nil, nil), errDummy)

	b := []byte("test")
	out, err := m.Bytes("dummy/nil", b)
	test.T(t, err, nil)
	test.Bytes(t, out, []byte{}, "dummy/nil returns empty byte slice")
	out, err = m.Bytes("?", b)
	test.T(t, err, ErrNotExist, "minifier doesn't exist")
	test.Bytes(t, out, b, "return input when minifier doesn't exist")

	s := "test"
	out2, err := m.String("dummy/nil", s)
	test.T(t, err, nil)
	test.String(t, out2, "", "dummy/nil returns empty string")
	out2, err = m.String("?", s)
	test.T(t, err, ErrNotExist, "minifier doesn't exist")
	test.String(t, out2, s, "return input when minifier doesn't exist")
}

type DummyMinifier struct{}

func (d *DummyMinifier) Minify(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
	return errDummy
}

func TestAdd(t *testing.T) {
	mAdd := New()
	r := bytes.NewBufferString("test")
	w := &bytes.Buffer{}
	mAdd.Add("dummy/err", &DummyMinifier{})
	test.T(t, mAdd.Minify("dummy/err", nil, nil), errDummy)

	mAdd.AddRegexp(regexp.MustCompile("err1$"), &DummyMinifier{})
	test.T(t, mAdd.Minify("dummy/err1", nil, nil), errDummy)

	mAdd.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	test.T(t, mAdd.Minify("dummy/err", nil, nil), errDummy)

	mAdd.AddFuncRegexp(regexp.MustCompile("err2$"), func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	test.T(t, mAdd.Minify("dummy/err2", nil, nil), errDummy)

	mAdd.AddCmd("dummy/copy", helperCommand(t, "dummy/copy"))
	mAdd.AddCmd("dummy/err", helperCommand(t, "dummy/err"))
	mAdd.AddCmdRegexp(regexp.MustCompile("err6$"), helperCommand(t, "werr6"))
	test.T(t, mAdd.Minify("dummy/copy", w, r), nil)
	test.String(t, w.String(), "test", "dummy/copy command returns input")
	test.String(t, mAdd.Minify("dummy/err", w, r).Error(), "exit status 1", "command returns status 1 for dummy/err")
	test.String(t, mAdd.Minify("werr6", w, r).Error(), "exit status 2", "command returns status 2 when minifier doesn't exist")
	test.String(t, mAdd.Minify("stderr6", w, r).Error(), "exit status 2", "command returns status 2 when minifier doesn't exist")
}

func TestMatch(t *testing.T) {
	pattern, params, _ := m.Match("dummy/copy; a=b")
	test.String(t, pattern, "dummy/copy")
	test.String(t, params["a"], "b")

	pattern, _, _ = m.Match("type/foobar")
	test.String(t, pattern, "^type/.+$")

	_, _, minifier := m.Match("dummy/")
	test.That(t, minifier == nil)
}

func TestWildcard(t *testing.T) {
	mimetypeTests := []struct {
		mimetype string
		expected string
	}{
		{"type/sub", "type/sub"},
		{"type/*", "type/*"},
		{"*/*", "*/*"},
		{"type/sub2", "type/*"},
		{"type2/sub", "*/*"},
		{"dummy/charset;charset=UTF-8", "UTF-8"},
		{"dummy/charset; charset = UTF-8 ", "UTF-8"},
		{"dummy/params;type=type;sub=two2", "type/*"},
	}

	for _, tt := range mimetypeTests {
		r := bytes.NewBufferString("")
		w := &bytes.Buffer{}
		err := m.Minify(tt.mimetype, w, r)
		test.Error(t, err)
		test.Minify(t, tt.mimetype, nil, w.String(), tt.expected)
	}
}

func TestReader(t *testing.T) {
	m := New()
	m.AddFunc("dummy/dummy", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	m.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})

	w := &bytes.Buffer{}
	r := bytes.NewBufferString("test")
	mr := m.Reader("dummy/dummy", r)
	_, err := io.Copy(w, mr)
	test.Error(t, err)
	test.String(t, w.String(), "test", "equal input after dummy minify reader")

	mr = m.Reader("dummy/err", r)
	_, err = io.Copy(w, mr)
	test.T(t, err, errDummy)
}

func TestWriter(t *testing.T) {
	m := New()
	m.AddFunc("dummy/dummy", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})
	m.AddFunc("dummy/err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		return errDummy
	})
	m.AddFunc("dummy/late-err", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, _ = ioutil.ReadAll(r)
		return errDummy
	})

	w := &bytes.Buffer{}
	mw := m.Writer("dummy/dummy", w)
	_, _ = mw.Write([]byte("test"))
	test.Error(t, mw.Close())
	test.String(t, w.String(), "test", "equal input after dummy minify writer")

	w = &bytes.Buffer{}
	mw = m.Writer("dummy/err", w)
	_, _ = mw.Write([]byte("test"))
	test.T(t, mw.Close(), errDummy)
	test.String(t, w.String(), "test", "equal input after dummy minify writer")

	w = &bytes.Buffer{}
	mw = m.Writer("dummy/late-err", w)
	_, _ = mw.Write([]byte("test"))
	test.T(t, mw.Close(), errDummy)
	test.String(t, w.String(), "")
}

type responseWriter struct {
	writer io.Writer
	header http.Header
}

func (w *responseWriter) Header() http.Header {
	return w.header
}

func (w *responseWriter) WriteHeader(_ int) {}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func TestResponseWriter(t *testing.T) {
	m := New()
	m.AddFunc("text/html", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})

	b := &bytes.Buffer{}
	w := &responseWriter{b, http.Header{}}
	r := &http.Request{RequestURI: "/index.html"}
	mw := m.ResponseWriter(w, r)
	test.Error(t, mw.Close())
	_, _ = mw.Write([]byte("test"))
	test.Error(t, mw.Close())
	test.String(t, b.String(), "test", "equal input after dummy minify response writer")

	b = &bytes.Buffer{}
	w = &responseWriter{b, http.Header{}}
	r = &http.Request{RequestURI: "/index"}
	mw = m.ResponseWriter(w, r)
	mw.Header().Add("Content-Type", "text/html")
	_, _ = mw.Write([]byte("test"))
	test.Error(t, mw.Close())
	test.String(t, b.String(), "test", "equal input after dummy minify response writer")
}

func TestMiddleware(t *testing.T) {
	m := New()
	m.AddFunc("text/html", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		_, err := io.Copy(w, r)
		return err
	})

	b := &bytes.Buffer{}
	w := &responseWriter{b, http.Header{}}
	r := &http.Request{RequestURI: "/index.html"}
	m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	})).ServeHTTP(w, r)
	test.String(t, b.String(), "test", "equal input after dummy minify middleware")
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	switch args[0] {
	case "dummy/copy":
		io.Copy(os.Stdout, os.Stdin)
	case "dummy/err":
		os.Exit(1)
	default:
		os.Exit(2)
	}
	os.Exit(0)
}

////////////////////////////////////////////////////////////////

func ExampleM_Minify_custom() {
	m := New()
	m.AddFunc("text/plain", func(m *M, w io.Writer, r io.Reader, _ map[string]string) error {
		// remove all newlines and spaces
		rb := bufio.NewReader(r)
		for {
			line, err := rb.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}
			if _, errws := io.WriteString(w, strings.Replace(line, " ", "", -1)); errws != nil {
				return errws
			}
			if err == io.EOF {
				break
			}
		}
		return nil
	})

	in := "Because my coffee was too cold, I heated it in the microwave."
	out, err := m.String("text/plain", in)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output: Becausemycoffeewastoocold,Iheateditinthemicrowave.
}

func ExampleM_Reader() {
	b := bytes.NewReader([]byte("input"))

	m := New()
	// add minfiers

	r := m.Reader("mime/type", b)
	if _, err := io.Copy(os.Stdout, r); err != nil {
		if _, err := io.Copy(os.Stdout, b); err != nil {
			panic(err)
		}
	}
}

func ExampleM_Writer() {
	m := New()
	// add minfiers

	w := m.Writer("mime/type", os.Stdout)
	if _, err := w.Write([]byte("input")); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
}

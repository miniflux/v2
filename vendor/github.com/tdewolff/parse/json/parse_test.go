package json // import "github.com/tdewolff/parse/json"

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/parse"
	"github.com/tdewolff/test"
)

type GTs []GrammarType

func TestGrammars(t *testing.T) {
	var grammarTests = []struct {
		json     string
		expected []GrammarType
	}{
		{" \t\n\r", GTs{}}, // WhitespaceGrammar
		{"null", GTs{LiteralGrammar}},
		{"[]", GTs{StartArrayGrammar, EndArrayGrammar}},
		{"15.2", GTs{NumberGrammar}},
		{"0.4", GTs{NumberGrammar}},
		{"5e9", GTs{NumberGrammar}},
		{"-4E-3", GTs{NumberGrammar}},
		{"true", GTs{LiteralGrammar}},
		{"false", GTs{LiteralGrammar}},
		{"null", GTs{LiteralGrammar}},
		{`""`, GTs{StringGrammar}},
		{`"abc"`, GTs{StringGrammar}},
		{`"\""`, GTs{StringGrammar}},
		{`"\\"`, GTs{StringGrammar}},
		{"{}", GTs{StartObjectGrammar, EndObjectGrammar}},
		{`{"a": "b", "c": "d"}`, GTs{StartObjectGrammar, StringGrammar, StringGrammar, StringGrammar, StringGrammar, EndObjectGrammar}},
		{`{"a": [1, 2], "b": {"c": 3}}`, GTs{StartObjectGrammar, StringGrammar, StartArrayGrammar, NumberGrammar, NumberGrammar, EndArrayGrammar, StringGrammar, StartObjectGrammar, StringGrammar, NumberGrammar, EndObjectGrammar, EndObjectGrammar}},
		{"[null,]", GTs{StartArrayGrammar, LiteralGrammar, EndArrayGrammar}},
		// {"[\"x\\\x00y\", 0]", GTs{StartArrayGrammar, StringGrammar, NumberGrammar, EndArrayGrammar}},
	}
	for _, tt := range grammarTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(bytes.NewBufferString(tt.json))
			i := 0
			for {
				grammar, _ := p.Next()
				if grammar == ErrorGrammar {
					test.T(t, p.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if grammar == WhitespaceGrammar {
					continue
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected grammar types size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, grammar, tt.expected[i], "grammar types must match")
				}
				i++
			}
		})
	}

	test.T(t, WhitespaceGrammar.String(), "Whitespace")
	test.T(t, GrammarType(100).String(), "Invalid(100)")
	test.T(t, ValueState.String(), "Value")
	test.T(t, ObjectKeyState.String(), "ObjectKey")
	test.T(t, ObjectValueState.String(), "ObjectValue")
	test.T(t, ArrayState.String(), "Array")
	test.T(t, State(100).String(), "Invalid(100)")
}

func TestGrammarsError(t *testing.T) {
	var grammarErrorTests = []struct {
		json string
		col  int
	}{
		{"true, false", 5},
		{"[true false]", 7},
		{"]", 1},
		{"}", 1},
		{"{0: 1}", 2},
		{"{\"a\" 1}", 6},
		{"1.", 2},
		{"1e+", 2},
		{`{"":"`, 0},
		{"\"a\\", 0},
	}
	for _, tt := range grammarErrorTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(bytes.NewBufferString(tt.json))
			for {
				grammar, _ := p.Next()
				if grammar == ErrorGrammar {
					if tt.col == 0 {
						test.T(t, p.Err(), io.EOF)
					} else if perr, ok := p.Err().(*parse.Error); ok {
						test.T(t, perr.Col, tt.col)
					} else {
						test.Fail(t, "bad error:", p.Err())
					}
					break
				}
			}
		})
	}
}

func TestStates(t *testing.T) {
	var stateTests = []struct {
		json     string
		expected []State
	}{
		{"null", []State{ValueState}},
		{"[null]", []State{ArrayState, ArrayState, ValueState}},
		{"{\"\":null}", []State{ObjectKeyState, ObjectValueState, ObjectKeyState, ValueState}},
	}
	for _, tt := range stateTests {
		t.Run(tt.json, func(t *testing.T) {
			p := NewParser(bytes.NewBufferString(tt.json))
			i := 0
			for {
				grammar, _ := p.Next()
				state := p.State()
				if grammar == ErrorGrammar {
					test.T(t, p.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if grammar == WhitespaceGrammar {
					continue
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected states size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, state, tt.expected[i], "states must match")
				}
				i++
			}
		})
	}
}

////////////////////////////////////////////////////////////////

func ExampleNewParser() {
	p := NewParser(bytes.NewBufferString(`{"key": 5}`))
	out := ""
	for {
		state := p.State()
		gt, data := p.Next()
		if gt == ErrorGrammar {
			break
		}
		out += string(data)
		if state == ObjectKeyState && gt != EndObjectGrammar {
			out += ":"
		}
		// not handling comma insertion
	}
	fmt.Println(out)
	// Output: {"key":5}
}

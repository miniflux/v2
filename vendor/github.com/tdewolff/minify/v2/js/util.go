package js

import (
	"bytes"
	"encoding/hex"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2/js"
	"github.com/tdewolff/parse/v2/strconv"
)

var (
	spaceBytes                 = []byte(" ")
	newlineBytes               = []byte("\n")
	starBytes                  = []byte("*")
	colonBytes                 = []byte(":")
	semicolonBytes             = []byte(";")
	commaBytes                 = []byte(",")
	dotBytes                   = []byte(".")
	ellipsisBytes              = []byte("...")
	openBraceBytes             = []byte("{")
	closeBraceBytes            = []byte("}")
	openParenBytes             = []byte("(")
	closeParenBytes            = []byte(")")
	openBracketBytes           = []byte("[")
	closeBracketBytes          = []byte("]")
	openParenBracketBytes      = []byte("({")
	closeParenOpenBracketBytes = []byte("){")
	notBytes                   = []byte("!")
	questionBytes              = []byte("?")
	equalBytes                 = []byte("=")
	notNotBytes                = []byte("!!")
	andBytes                   = []byte("&&")
	orBytes                    = []byte("||")
	optChainBytes              = []byte("?.")
	nullishBytes               = []byte("??")
	arrowBytes                 = []byte("=>")
	zeroBytes                  = []byte("0")
	oneBytes                   = []byte("1")
	letBytes                   = []byte("let")
	getBytes                   = []byte("get")
	setBytes                   = []byte("set")
	asyncBytes                 = []byte("async")
	functionBytes              = []byte("function")
	staticBytes                = []byte("static")
	ifOpenBytes                = []byte("if(")
	elseBytes                  = []byte("else")
	withOpenBytes              = []byte("with(")
	doBytes                    = []byte("do")
	whileOpenBytes             = []byte("while(")
	forOpenBytes               = []byte("for(")
	forAwaitOpenBytes          = []byte("for await(")
	inBytes                    = []byte("in")
	ofBytes                    = []byte("of")
	switchOpenBytes            = []byte("switch(")
	throwBytes                 = []byte("throw")
	tryBytes                   = []byte("try")
	catchBytes                 = []byte("catch")
	finallyBytes               = []byte("finally")
	importBytes                = []byte("import")
	exportBytes                = []byte("export")
	fromBytes                  = []byte("from")
	returnBytes                = []byte("return")
	classBytes                 = []byte("class")
	asSpaceBytes               = []byte("as ")
	asyncSpaceBytes            = []byte("async ")
	spaceDefaultBytes          = []byte(" default")
	spaceExtendsBytes          = []byte(" extends")
	yieldBytes                 = []byte("yield")
	newBytes                   = []byte("new")
	openNewBytes               = []byte("(new")
	newTargetBytes             = []byte("new.target")
	importMetaBytes            = []byte("import.meta")
	nanBytes                   = []byte("NaN")
	undefinedBytes             = []byte("undefined")
	infinityBytes              = []byte("Infinity")
	voidZeroBytes              = []byte("void 0")
	groupedVoidZeroBytes       = []byte("(void 0)")
	oneDivZeroBytes            = []byte("1/0")
	groupedOneDivZeroBytes     = []byte("(1/0)")
	notZeroBytes               = []byte("!0")
	groupedNotZeroBytes        = []byte("(!0)")
	notOneBytes                = []byte("!1")
	groupedNotOneBytes         = []byte("(!1)")
	debuggerBytes              = []byte("debugger")
	regExpScriptBytes          = []byte("/script>")
)

// precedence maps for the precedence inside the operation
var unaryPrecMap = map[js.TokenType]js.OpPrec{
	js.PostIncrToken: js.OpLHS,
	js.PostDecrToken: js.OpLHS,
	js.PreIncrToken:  js.OpUnary,
	js.PreDecrToken:  js.OpUnary,
	js.NotToken:      js.OpUnary,
	js.BitNotToken:   js.OpUnary,
	js.TypeofToken:   js.OpUnary,
	js.VoidToken:     js.OpUnary,
	js.DeleteToken:   js.OpUnary,
	js.PosToken:      js.OpUnary,
	js.NegToken:      js.OpUnary,
	js.AwaitToken:    js.OpUnary,
}

var binaryLeftPrecMap = map[js.TokenType]js.OpPrec{
	js.EqToken:         js.OpLHS,
	js.MulEqToken:      js.OpLHS,
	js.DivEqToken:      js.OpLHS,
	js.ModEqToken:      js.OpLHS,
	js.ExpEqToken:      js.OpLHS,
	js.AddEqToken:      js.OpLHS,
	js.SubEqToken:      js.OpLHS,
	js.LtLtEqToken:     js.OpLHS,
	js.GtGtEqToken:     js.OpLHS,
	js.GtGtGtEqToken:   js.OpLHS,
	js.BitAndEqToken:   js.OpLHS,
	js.BitXorEqToken:   js.OpLHS,
	js.BitOrEqToken:    js.OpLHS,
	js.ExpToken:        js.OpUpdate,
	js.MulToken:        js.OpMul,
	js.DivToken:        js.OpMul,
	js.ModToken:        js.OpMul,
	js.AddToken:        js.OpAdd,
	js.SubToken:        js.OpAdd,
	js.LtLtToken:       js.OpShift,
	js.GtGtToken:       js.OpShift,
	js.GtGtGtToken:     js.OpShift,
	js.LtToken:         js.OpCompare,
	js.LtEqToken:       js.OpCompare,
	js.GtToken:         js.OpCompare,
	js.GtEqToken:       js.OpCompare,
	js.InToken:         js.OpCompare,
	js.InstanceofToken: js.OpCompare,
	js.EqEqToken:       js.OpEquals,
	js.NotEqToken:      js.OpEquals,
	js.EqEqEqToken:     js.OpEquals,
	js.NotEqEqToken:    js.OpEquals,
	js.BitAndToken:     js.OpBitAnd,
	js.BitXorToken:     js.OpBitXor,
	js.BitOrToken:      js.OpBitOr,
	js.AndToken:        js.OpAnd,
	js.OrToken:         js.OpOr,
	js.NullishToken:    js.OpBitOr, // or OpCoalesce
	js.CommaToken:      js.OpExpr,
}

var binaryRightPrecMap = map[js.TokenType]js.OpPrec{
	js.EqToken:         js.OpAssign,
	js.MulEqToken:      js.OpAssign,
	js.DivEqToken:      js.OpAssign,
	js.ModEqToken:      js.OpAssign,
	js.ExpEqToken:      js.OpAssign,
	js.AddEqToken:      js.OpAssign,
	js.SubEqToken:      js.OpAssign,
	js.LtLtEqToken:     js.OpAssign,
	js.GtGtEqToken:     js.OpAssign,
	js.GtGtGtEqToken:   js.OpAssign,
	js.BitAndEqToken:   js.OpAssign,
	js.BitXorEqToken:   js.OpAssign,
	js.BitOrEqToken:    js.OpAssign,
	js.ExpToken:        js.OpExp,
	js.MulToken:        js.OpExp,
	js.DivToken:        js.OpExp,
	js.ModToken:        js.OpExp,
	js.AddToken:        js.OpMul,
	js.SubToken:        js.OpMul,
	js.LtLtToken:       js.OpAdd,
	js.GtGtToken:       js.OpAdd,
	js.GtGtGtToken:     js.OpAdd,
	js.LtToken:         js.OpShift,
	js.LtEqToken:       js.OpShift,
	js.GtToken:         js.OpShift,
	js.GtEqToken:       js.OpShift,
	js.InToken:         js.OpShift,
	js.InstanceofToken: js.OpShift,
	js.EqEqToken:       js.OpCompare,
	js.NotEqToken:      js.OpCompare,
	js.EqEqEqToken:     js.OpCompare,
	js.NotEqEqToken:    js.OpCompare,
	js.BitAndToken:     js.OpEquals,
	js.BitXorToken:     js.OpBitAnd,
	js.BitOrToken:      js.OpBitXor,
	js.AndToken:        js.OpAnd,   // changes order in AST but not in execution
	js.OrToken:         js.OpOr,    // changes order in AST but not in execution
	js.NullishToken:    js.OpBitOr, // or OpCoalesce
	js.CommaToken:      js.OpAssign,
}

// precedence maps of the operation itself
var unaryOpPrecMap = map[js.TokenType]js.OpPrec{
	js.PostIncrToken: js.OpUpdate,
	js.PostDecrToken: js.OpUpdate,
	js.PreIncrToken:  js.OpUpdate,
	js.PreDecrToken:  js.OpUpdate,
	js.NotToken:      js.OpUnary,
	js.BitNotToken:   js.OpUnary,
	js.TypeofToken:   js.OpUnary,
	js.VoidToken:     js.OpUnary,
	js.DeleteToken:   js.OpUnary,
	js.PosToken:      js.OpUnary,
	js.NegToken:      js.OpUnary,
	js.AwaitToken:    js.OpUnary,
}

var binaryOpPrecMap = map[js.TokenType]js.OpPrec{
	js.EqToken:         js.OpAssign,
	js.MulEqToken:      js.OpAssign,
	js.DivEqToken:      js.OpAssign,
	js.ModEqToken:      js.OpAssign,
	js.ExpEqToken:      js.OpAssign,
	js.AddEqToken:      js.OpAssign,
	js.SubEqToken:      js.OpAssign,
	js.LtLtEqToken:     js.OpAssign,
	js.GtGtEqToken:     js.OpAssign,
	js.GtGtGtEqToken:   js.OpAssign,
	js.BitAndEqToken:   js.OpAssign,
	js.BitXorEqToken:   js.OpAssign,
	js.BitOrEqToken:    js.OpAssign,
	js.ExpToken:        js.OpExp,
	js.MulToken:        js.OpMul,
	js.DivToken:        js.OpMul,
	js.ModToken:        js.OpMul,
	js.AddToken:        js.OpAdd,
	js.SubToken:        js.OpAdd,
	js.LtLtToken:       js.OpShift,
	js.GtGtToken:       js.OpShift,
	js.GtGtGtToken:     js.OpShift,
	js.LtToken:         js.OpCompare,
	js.LtEqToken:       js.OpCompare,
	js.GtToken:         js.OpCompare,
	js.GtEqToken:       js.OpCompare,
	js.InToken:         js.OpCompare,
	js.InstanceofToken: js.OpCompare,
	js.EqEqToken:       js.OpEquals,
	js.NotEqToken:      js.OpEquals,
	js.EqEqEqToken:     js.OpEquals,
	js.NotEqEqToken:    js.OpEquals,
	js.BitAndToken:     js.OpBitAnd,
	js.BitXorToken:     js.OpBitXor,
	js.BitOrToken:      js.OpBitOr,
	js.AndToken:        js.OpAnd,
	js.OrToken:         js.OpOr,
	js.NullishToken:    js.OpCoalesce,
	js.CommaToken:      js.OpExpr,
}

func exprPrec(i js.IExpr) js.OpPrec {
	switch expr := i.(type) {
	case *js.Var, *js.LiteralExpr, *js.ArrayExpr, *js.ObjectExpr, *js.FuncDecl, *js.ClassDecl:
		return js.OpPrimary
	case *js.UnaryExpr:
		return unaryOpPrecMap[expr.Op]
	case *js.BinaryExpr:
		return binaryOpPrecMap[expr.Op]
	case *js.NewExpr:
		if expr.Args == nil {
			return js.OpNew
		}
		return js.OpMember
	case *js.TemplateExpr:
		if expr.Tag == nil {
			return js.OpPrimary
		}
		return expr.Prec
	case *js.DotExpr:
		return expr.Prec
	case *js.IndexExpr:
		return expr.Prec
	case *js.NewTargetExpr, *js.ImportMetaExpr:
		return js.OpMember
	case *js.OptChainExpr, *js.CallExpr:
		return js.OpCall
	case *js.CondExpr, *js.YieldExpr, *js.ArrowFunc:
		return js.OpAssign
	case *js.GroupExpr:
		return exprPrec(expr.X)
	}
	return js.OpExpr // does not happen
}

// TODO: use in more cases
func groupExpr(i js.IExpr, prec js.OpPrec) js.IExpr {
	precInside := exprPrec(i)
	if _, ok := i.(*js.GroupExpr); !ok && precInside < prec && (precInside != js.OpCoalesce || prec != js.OpBitOr) {
		return &js.GroupExpr{X: i}
	}
	return i
}

// TODO: use in more cases
func condExpr(cond, x, y js.IExpr) js.IExpr {
	return &js.CondExpr{
		Cond: groupExpr(cond, js.OpCoalesce),
		X:    groupExpr(x, js.OpAssign),
		Y:    groupExpr(y, js.OpAssign),
	}
}

func commaExpr(x, y js.IExpr) js.IExpr {
	comma, ok := x.(*js.CommaExpr)
	if !ok {
		comma = &js.CommaExpr{List: []js.IExpr{x}}
	}
	if comma2, ok := y.(*js.CommaExpr); ok {
		comma.List = append(comma.List, comma2.List...)
	} else {
		comma.List = append(comma.List, y)
	}
	return comma
}

func (m *jsMinifier) isEmptyStmt(stmt js.IStmt) bool {
	if stmt == nil {
		return true
	} else if _, ok := stmt.(*js.EmptyStmt); ok {
		return true
	} else if decl, ok := stmt.(*js.VarDecl); ok && decl.TokenType == js.ErrorToken {
		for _, item := range decl.List {
			if item.Default != nil {
				return false
			}
		}
		return true
	} else if block, ok := stmt.(*js.BlockStmt); ok {
		for _, item := range block.List {
			if ok := m.isEmptyStmt(item); !ok {
				return false
			}
		}
		return true
	}
	return false
}

func finalExpr(expr js.IExpr) js.IExpr {
	if group, ok := expr.(*js.GroupExpr); ok {
		expr = group.X
	}
	if comma, ok := expr.(*js.CommaExpr); ok {
		expr = comma.List[len(comma.List)-1]
	}
	if binary, ok := expr.(*js.BinaryExpr); ok && binary.Op == js.EqToken {
		expr = binary.X // return first
	}
	return expr
}

func isFlowStmt(stmt js.IStmt) bool {
	if _, ok := stmt.(*js.ReturnStmt); ok {
		return true
	} else if _, ok := stmt.(*js.ThrowStmt); ok {
		return true
	} else if _, ok := stmt.(*js.BranchStmt); ok {
		return true
	}
	return false
}

func lastStmt(stmt js.IStmt) js.IStmt {
	if block, ok := stmt.(*js.BlockStmt); ok && 0 < len(block.List) {
		return block.List[len(block.List)-1]
	}
	return stmt
}

func (m *jsMinifier) isTrue(i js.IExpr) bool {
	if lit, ok := i.(*js.LiteralExpr); ok && lit.TokenType == js.TrueToken {
		return true
	} else if unary, ok := i.(*js.UnaryExpr); ok && unary.Op == js.NotToken {
		ret, _ := m.isFalsy(unary.X)
		return ret
	}
	return false
}

func (m *jsMinifier) isFalse(i js.IExpr) bool {
	if lit, ok := i.(*js.LiteralExpr); ok {
		return lit.TokenType == js.FalseToken
	} else if unary, ok := i.(*js.UnaryExpr); ok && unary.Op == js.NotToken {
		ret, _ := m.isTruthy(unary.X)
		return ret
	}
	return false
}

func (m *jsMinifier) toNullishExpr(condExpr *js.CondExpr) (js.IExpr, js.IExpr, bool) {
	// convert conditional expression to nullish:  a!=null?a:b  =>  a??b
	if binaryExpr, ok := condExpr.Cond.(*js.BinaryExpr); ok && (binaryExpr.Op == js.EqEqToken || binaryExpr.Op == js.NotEqToken) {
		var left, right js.IExpr
		if binaryExpr.Op == js.EqEqToken {
			left = condExpr.Y
			right = condExpr.X
		} else {
			left = condExpr.X
			right = condExpr.Y
		}
		if lit, ok := binaryExpr.X.(*js.LiteralExpr); ((ok && lit.TokenType == js.NullToken) || m.isUndefined(binaryExpr.X)) && m.isEqualExpr(binaryExpr.Y, left) {
			return left, right, true
		} else if lit, ok := binaryExpr.Y.(*js.LiteralExpr); ((ok && lit.TokenType == js.NullToken) || m.isUndefined(binaryExpr.Y)) && m.isEqualExpr(binaryExpr.X, left) {
			return left, right, true
		}
	}
	return nil, nil, false
}

func (m *jsMinifier) isUndefined(i js.IExpr) bool {
	if v, ok := i.(*js.Var); ok {
		if bytes.Equal(v.Name(), undefinedBytes) { // TODO: only if not defined
			return true
		}
	} else if unary, ok := i.(*js.UnaryExpr); ok && unary.Op == js.VoidToken {
		return true
	}
	return false
}

// returns whether truthy and whether it could be coerced to a boolean (i.e. when returns (false,true) this means it is falsy)
func (m *jsMinifier) isTruthy(i js.IExpr) (bool, bool) {
	if falsy, ok := m.isFalsy(i); ok {
		return !falsy, true
	}
	return false, false
}

// returns whether falsy and whether it could be coerced to a boolean (i.e. when returns (false,true) this means it is truthy)
func (m *jsMinifier) isFalsy(i js.IExpr) (bool, bool) {
	negated := false
	group, isGroup := i.(*js.GroupExpr)
	unary, isUnary := i.(*js.UnaryExpr)
	for isGroup || isUnary && unary.Op == js.NotToken {
		if isGroup {
			i = group.X
		} else {
			i = unary.X
			negated = !negated
		}
		group, isGroup = i.(*js.GroupExpr)
		unary, isUnary = i.(*js.UnaryExpr)
	}
	if lit, ok := i.(*js.LiteralExpr); ok {
		tt := lit.TokenType
		d := lit.Data
		if tt == js.FalseToken || tt == js.NullToken || tt == js.StringToken && len(lit.Data) == 0 {
			return !negated, true // falsy
		} else if tt == js.TrueToken || tt == js.StringToken {
			return negated, true // truthy
		} else if tt == js.DecimalToken || tt == js.BinaryToken || tt == js.OctalToken || tt == js.HexadecimalToken || tt == js.BigIntToken {
			for _, c := range d {
				if c == 'e' || c == 'E' || c == 'n' {
					break
				} else if c != '0' && c != '.' && c != 'x' && c != 'X' && c != 'b' && c != 'B' && c != 'o' && c != 'O' {
					return negated, true // truthy
				}
			}
			return !negated, true // falsy
		}
	} else if m.isUndefined(i) {
		return !negated, true // falsy
	} else if v, ok := i.(*js.Var); ok && bytes.Equal(v.Name(), nanBytes) {
		return !negated, true // falsy
	}
	return false, false // unknown
}

func (m *jsMinifier) isEqualExpr(a, b js.IExpr) bool {
	if group, ok := a.(*js.GroupExpr); ok {
		a = group.X
	}
	if group, ok := b.(*js.GroupExpr); ok {
		b = group.X
	}
	if left, ok := a.(*js.Var); ok {
		if right, ok := b.(*js.Var); ok {
			return bytes.Equal(left.Name(), right.Name())
		}
	}
	// TODO: use reflect.DeepEqual?
	return false
}

func isBooleanExpr(expr js.IExpr) bool {
	if unaryExpr, ok := expr.(*js.UnaryExpr); ok {
		return unaryExpr.Op == js.NotToken
	} else if binaryExpr, ok := expr.(*js.BinaryExpr); ok {
		op := binaryOpPrecMap[binaryExpr.Op]
		return op == js.OpCompare || op == js.OpEquals
	} else if litExpr, ok := expr.(*js.LiteralExpr); ok {
		return litExpr.TokenType == js.TrueToken || litExpr.TokenType == js.FalseToken
	} else if groupExpr, ok := expr.(*js.GroupExpr); ok {
		return isBooleanExpr(groupExpr.X)
	}
	return false
}

func (m *jsMinifier) minifyBooleanExpr(expr js.IExpr, invert bool, prec js.OpPrec) {
	if invert {
		// unary !(boolean) has already been handled
		if binaryExpr, ok := expr.(*js.BinaryExpr); ok && binaryOpPrecMap[binaryExpr.Op] == js.OpEquals {
			if binaryExpr.Op == js.EqEqToken {
				binaryExpr.Op = js.NotEqToken
			} else if binaryExpr.Op == js.NotEqToken {
				binaryExpr.Op = js.EqEqToken
			} else if binaryExpr.Op == js.EqEqEqToken {
				binaryExpr.Op = js.NotEqEqToken
			} else if binaryExpr.Op == js.NotEqEqToken {
				binaryExpr.Op = js.EqEqEqToken
			}
			m.minifyExpr(expr, prec)
		} else {
			m.write(notBytes)
			m.minifyExpr(&js.GroupExpr{X: expr}, js.OpUnary)
		}
	} else if isBooleanExpr(expr) {
		m.minifyExpr(&js.GroupExpr{X: expr}, prec)
	} else {
		m.write(notNotBytes)
		m.minifyExpr(&js.GroupExpr{X: expr}, js.OpUnary)
	}
}

func endsInIf(istmt js.IStmt) bool {
	switch stmt := istmt.(type) {
	case *js.IfStmt:
		if stmt.Else == nil {
			return true
		}
		return endsInIf(stmt.Else)
	case *js.BlockStmt:
		if 0 < len(stmt.List) {
			return endsInIf(stmt.List[len(stmt.List)-1])
		}
	case *js.LabelledStmt:
		return endsInIf(stmt.Value)
	case *js.WithStmt:
		return endsInIf(stmt.Body)
	case *js.WhileStmt:
		return endsInIf(stmt.Body)
	case *js.ForStmt:
		return endsInIf(stmt.Body)
	case *js.ForInStmt:
		return endsInIf(stmt.Body)
	case *js.ForOfStmt:
		return endsInIf(stmt.Body)
	}
	return false
}

func isHexDigit(b byte) bool {
	return '0' <= b && b <= '9' || 'a' <= b && b <= 'f' || 'A' <= b && b <= 'F'
}

func minifyString(b []byte) []byte {
	if len(b) < 3 {
		return b
	}

	// switch quotes if more optimal
	singleQuotes := 0
	doubleQuotes := 0
	for i := 1; i < len(b)-1; i++ {
		if b[i] == '\'' {
			singleQuotes++
		} else if b[i] == '"' {
			doubleQuotes++
		}
	}
	quote := byte('"')
	if singleQuotes < doubleQuotes {
		quote = byte('\'')
	}
	b[0] = quote
	b[len(b)-1] = quote

	// strip unnecessary escapes
	j := 0
	start := 0
	for i := 1; i < len(b)-1; i++ {
		if c := b[i]; c == '\\' {
			c = b[i+1]
			if c == '0' && (i+2 == len(b)-1 || b[i+2] < '0' || '7' < b[i+2]) || c == '\\' || c == quote || c == 'n' || c == 'r' || c == 'u' {
				// keep escape sequence
				i++
				continue
			}
			n := 1
			if c == '\n' || c == '\r' || c == 0xE2 && i+3 < len(b)-1 && b[i+2] == 0x80 && (b[i+3] == 0xA8 || b[i+3] == 0xA9) {
				// line continuations
				if c == 0xE2 {
					n = 4
				} else if c == '\r' && i+2 < len(b)-1 && b[i+2] == '\n' {
					n = 3
				} else {
					n = 2
				}
			} else if c == 'x' {
				if i+3 < len(b)-1 && isHexDigit(b[i+2]) && b[i+2] < '8' && isHexDigit(b[i+3]) {
					// hexadecimal escapes
					_, _ = hex.Decode(b[i+3:i+4:i+4], b[i+2:i+4])
					n = 3
					if b[i+3] == 0 || b[i+3] == '\\' || b[i+3] == quote || b[i+3] == '\n' || b[i+3] == '\r' {
						if b[i+3] == 0 {
							b[i+3] = '0'
						} else if b[i+3] == '\n' {
							b[i+3] = 'n'
						} else if b[i+3] == '\r' {
							b[i+3] = 'r'
						}
						n--
						b[i+2] = '\\'
					}
				} else {
					i++
					continue
				}
			} else if '0' <= c && c <= '7' {
				// octal escapes (legacy), \0 already handled
				num := c - '0'
				if i+2 < len(b)-1 && '0' <= b[i+2] && b[i+2] <= '7' {
					num = num*8 + b[i+2] - '0'
					n++
					if num < 32 && i+3 < len(b)-1 && '0' <= b[i+3] && b[i+3] <= '7' {
						num = num*8 + b[i+3] - '0'
						n++
					}
				}
				b[i+n] = num
				if num == 0 || num == '\\' || num == quote || num == '\n' || num == '\r' {
					if num == 0 {
						b[i+n] = '0'
					} else if num == '\n' {
						b[i+n] = 'n'
					} else if num == '\r' {
						b[i+n] = 'r'
					}
					n--
					b[i+n] = '\\'
				}
			} else if c == 't' {
				b[i+1] = '\t'
			} else if c == 'f' {
				b[i+1] = '\f'
			} else if c == 'v' {
				b[i+1] = '\v'
			} else if c == 'b' {
				b[i+1] = '\b'
			}
			// remove unnecessary escape character, anything but 0x00, 0x0A, 0x0D, \, ' or "
			if start != 0 {
				j += copy(b[j:], b[start:i])
			} else {
				j = i
			}
			start = i + n
			i += n - 1
		} else if c == quote {
			// may not be escaped properly when changing quotes
			if j < start {
				// avoid append
				j += copy(b[j:], b[start:i])
				b[j] = '\\'
				j++
				start = i
			} else {
				b = append(append(b[:i], '\\'), b[i:]...)
				i++
				b[i] = quote // was overwritten above
			}
		} else if c == '<' && 9 <= len(b)-1-i {
			if b[i+1] == '\\' && 10 <= len(b)-1-i && bytes.Equal(b[i+2:i+10], []byte("/script>")) {
				i += 9
			} else if bytes.Equal(b[i+1:i+9], []byte("/script>")) {
				i++
				if j < start {
					// avoid append
					j += copy(b[j:], b[start:i])
					b[j] = '\\'
					j++
					start = i
				} else {
					b = append(append(b[:i], '\\'), b[i:]...)
					i++
					b[i] = '/' // was overwritten above
				}
			}
		}
	}
	if start != 0 {
		j += copy(b[j:], b[start:])
		return b[:j]
	}
	return b
}

var regexpEscapeTable = [256]bool{
	// ASCII
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, true, false, false, false, // $
	true, true, true, true, false, false, true, true, // (, ), *, +, ., /
	true, true, true, true, true, true, true, true, // 0, 1, 2, 3, 4, 5, 6, 7
	true, true, false, false, false, false, false, true, // 8, 9, ?

	false, false, true, false, true, false, false, false, // B, D
	false, false, false, false, false, false, false, false,
	true, false, false, true, false, false, false, true, // P, S, W
	false, false, false, true, true, true, true, false, // [, \, ], ^

	false, false, true, true, true, false, true, false, // b, c, d, f
	false, false, false, true, false, false, true, false, // k, n
	true, false, true, true, true, true, true, true, // p, r, s, t, u, v, w
	true, false, false, true, true, true, false, false, // x, {, |, }

	// non-ASCII
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
}

var regexpClassEscapeTable = [256]bool{
	// ASCII
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	true, true, true, true, true, true, true, true, // 0, 1, 2, 3, 4, 5, 6, 7
	true, true, false, false, false, false, false, false, // 8, 9

	false, false, false, false, true, false, false, false, // D
	false, false, false, false, false, false, false, false,
	true, false, false, true, false, false, false, true, // P, S, W
	false, false, false, false, true, true, false, false, // \, ]

	false, false, true, true, true, false, true, false, // b, c, d, f
	false, false, false, false, false, false, true, false, // n
	true, false, true, true, true, true, true, true, // p, r, s, t, u, v, w
	true, false, false, false, false, false, false, false, // x

	// non-ASCII
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,

	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
	false, false, false, false, false, false, false, false,
}

func minifyRegExp(b []byte) []byte {
	inClass := false
	afterDash := 0
	iClass := 0
	for i := 1; i < len(b)-1; i++ {
		if inClass {
			afterDash++
		}
		if b[i] == '\\' {
			c := b[i+1]
			escape := true
			if inClass {
				escape = regexpClassEscapeTable[c] || c == '-' && 2 < afterDash && i+2 < len(b) && b[i+2] != ']' || c == '^' && i == iClass+1
			} else {
				escape = regexpEscapeTable[c]
			}
			if !escape {
				b = append(b[:i], b[i+1:]...)
				if inClass && 2 < afterDash && c == '-' {
					afterDash = 0
				} else if inClass && c == '^' {
					afterDash = 1
				}
			} else {
				i++
			}
		} else if b[i] == '[' {
			if b[i+1] == '^' {
				i++
			}
			afterDash = 1
			inClass = true
			iClass = i
		} else if inClass && b[i] == ']' {
			inClass = false
		} else if b[i] == '/' {
			break
		} else if inClass && 2 < afterDash && b[i] == '-' {
			afterDash = 0
		}
	}
	return b
}

func binaryNumber(b []byte, prec int) []byte {
	if len(b) <= 2 || 65 < len(b) {
		return b
	}
	var n int64
	for _, c := range b[2:] {
		n *= 2
		n += int64(c - '0')
	}
	i := strconv.LenInt(n) - 1
	b = b[:i+1]
	for 0 <= i {
		b[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return minify.Number(b, prec)
}

func octalNumber(b []byte, prec int) []byte {
	if len(b) <= 2 || 23 < len(b) {
		return b
	}
	var n int64
	for _, c := range b[2:] {
		n *= 8
		n += int64(c - '0')
	}
	i := strconv.LenInt(n) - 1
	b = b[:i+1]
	for 0 <= i {
		b[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return minify.Number(b, prec)
}

func hexadecimalNumber(b []byte, prec int) []byte {
	if len(b) <= 2 || 12 < len(b) || len(b) == 12 && ('D' < b[2] && b[2] <= 'F' || 'd' < b[2]) {
		return b
	}
	var n int64
	for _, c := range b[2:] {
		n *= 16
		if c <= '9' {
			n += int64(c - '0')
		} else if c <= 'F' {
			n += 10 + int64(c-'A')
		} else {
			n += 10 + int64(c-'a')
		}
	}
	i := strconv.LenInt(n) - 1
	b = b[:i+1]
	for 0 <= i {
		b[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	return minify.Number(b, prec)
}

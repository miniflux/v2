// Package js minifies ECMAScript 2021 following the language specification at https://tc39.es/ecma262/.
package js

import (
	"bytes"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

type blockType int

const (
	defaultBlock blockType = iota
	functionBlock
	iterationBlock
)

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JS minifier.
type Minifier struct {
	Precision    int // number of significant digits
	KeepVarNames bool
}

// Minify minifies JS data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies JS data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	z := parse.NewInput(r)
	ast, err := js.Parse(z)
	if err != nil {
		return err
	}

	// license comments
	for _, comment := range ast.Comments {
		if 3 < len(comment) && comment[2] == '!' {
			w.Write(comment)
			if comment[1] == '/' {
				w.Write(newlineBytes)
			}
		} else if 2 < len(comment) && comment[0] == '#' && comment[1] == '!' {
			w.Write(comment)
		}
	}

	m := &jsMinifier{
		o:       o,
		w:       w,
		renamer: newRenamer(ast, ast.Undeclared, !o.KeepVarNames),
	}
	m.hoistVars(&ast.BlockStmt)
	ast.List = m.optimizeStmtList(ast.List, functionBlock)
	for _, item := range ast.List {
		m.writeSemicolon()
		m.minifyStmt(item)
	}

	if _, err := w.Write(nil); err != nil {
		return err
	}
	return nil
}

type expectExpr int

const (
	expectAny      expectExpr = iota
	expectExprStmt            // in statement
	expectExprBody            // in arrow function body
)

type jsMinifier struct {
	o *Minifier
	w io.Writer

	prev           []byte
	needsSemicolon bool       // write a semicolon if required
	needsSpace     bool       // write a space if next token is an identifier
	expectExpr     expectExpr // avoid ambiguous syntax such as an expression starting with function
	groupedStmt    bool       // avoid ambiguous syntax by grouping the expression statement
	inFor          bool
	spaceBefore    byte
	varsHoisted    *js.VarDecl // set when variables are hoisted to this declaration

	renamer *renamer
}

func (m *jsMinifier) write(b []byte) {
	// 0 < len(b)
	if m.needsSpace && js.IsIdentifierContinue(b) || m.spaceBefore == b[0] {
		m.w.Write(spaceBytes)
	}
	m.w.Write(b)
	m.prev = b
	m.needsSpace = false
	m.expectExpr = expectAny
	m.spaceBefore = 0
}

func (m *jsMinifier) writeSpaceAfterIdent() {
	// space after identifier and after regular expression (to prevent confusion with its tag)
	if js.IsIdentifierEnd(m.prev) || 1 < len(m.prev) && m.prev[0] == '/' {
		m.w.Write(spaceBytes)
	}
}

func (m *jsMinifier) writeSpaceBeforeIdent() {
	m.needsSpace = true
}

func (m *jsMinifier) writeSpaceBefore(c byte) {
	m.spaceBefore = c
}

func (m *jsMinifier) requireSemicolon() {
	m.needsSemicolon = true
}

func (m *jsMinifier) writeSemicolon() {
	if m.needsSemicolon {
		m.w.Write(semicolonBytes)
		m.needsSemicolon = false
		m.needsSpace = false
	}
}

func (m *jsMinifier) minifyStmt(i js.IStmt) {
	switch stmt := i.(type) {
	case *js.ExprStmt:
		m.expectExpr = expectExprStmt
		m.minifyExpr(stmt.Value, js.OpExpr)
		if m.groupedStmt {
			m.write(closeParenBytes)
			m.groupedStmt = false
		}
		m.requireSemicolon()
	case *js.VarDecl:
		m.minifyVarDecl(stmt, false)
		m.requireSemicolon()
	case *js.IfStmt:
		hasIf := !m.isEmptyStmt(stmt.Body)
		hasElse := !m.isEmptyStmt(stmt.Else)

		m.write(ifOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)

		if !hasIf && hasElse {
			m.requireSemicolon()
		} else if hasIf {
			if hasElse && endsInIf(stmt.Body) {
				// prevent: if(a){if(b)c}else d;  =>  if(a)if(b)c;else d;
				m.write(openBraceBytes)
				m.minifyStmt(stmt.Body)
				m.write(closeBraceBytes)
				m.needsSemicolon = false
			} else {
				m.minifyStmt(stmt.Body)
			}
		}
		if hasElse {
			m.writeSemicolon()
			m.write(elseBytes)
			m.writeSpaceBeforeIdent()
			m.minifyStmt(stmt.Else)
		}
	case *js.BlockStmt:
		m.renamer.renameScope(stmt.Scope)
		m.minifyBlockStmt(stmt)
	case *js.ReturnStmt:
		m.write(returnBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.LabelledStmt:
		m.write(stmt.Label)
		m.write(colonBytes)
		m.minifyStmt(stmt.Value)
	case *js.BranchStmt:
		m.write(stmt.Type.Bytes())
		if stmt.Label != nil {
			m.write(spaceBytes)
			m.write(stmt.Label)
		}
		m.requireSemicolon()
	case *js.WithStmt:
		m.write(withOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, defaultBlock)
	case *js.DoWhileStmt:
		m.write(doBytes)
		m.writeSpaceBeforeIdent()
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
		m.writeSemicolon()
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
	case *js.WhileStmt:
		m.write(whileOpenBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyStmtOrBlock(stmt.Body, iterationBlock)
	case *js.ForStmt:
		stmt.Body.List = m.optimizeStmtList(stmt.Body.List, iterationBlock)
		m.renamer.renameScope(stmt.Body.Scope)
		m.write(forOpenBytes)
		m.inFor = true
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(decl, false)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.inFor = false
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Cond, js.OpExpr)
		m.write(semicolonBytes)
		m.minifyExpr(stmt.Post, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyBlockAsStmt(stmt.Body, iterationBlock)
	case *js.ForInStmt:
		stmt.Body.List = m.optimizeStmtList(stmt.Body.List, iterationBlock)
		m.renamer.renameScope(stmt.Body.Scope)
		m.write(forOpenBytes)
		m.inFor = true
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(decl, false)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.inFor = false
		m.writeSpaceAfterIdent()
		m.write(inBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.write(closeParenBytes)
		m.minifyBlockAsStmt(stmt.Body, iterationBlock)
	case *js.ForOfStmt:
		stmt.Body.List = m.optimizeStmtList(stmt.Body.List, iterationBlock)
		m.renamer.renameScope(stmt.Body.Scope)
		if stmt.Await {
			m.write(forAwaitOpenBytes)
		} else {
			m.write(forOpenBytes)
		}
		m.inFor = true
		if decl, ok := stmt.Init.(*js.VarDecl); ok {
			m.minifyVarDecl(decl, false)
		} else {
			m.minifyExpr(stmt.Init, js.OpLHS)
		}
		m.inFor = false
		m.writeSpaceAfterIdent()
		m.write(ofBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpAssign)
		m.write(closeParenBytes)
		m.minifyBlockAsStmt(stmt.Body, iterationBlock)
	case *js.SwitchStmt:
		m.write(switchOpenBytes)
		m.minifyExpr(stmt.Init, js.OpExpr)
		m.write(closeParenOpenBracketBytes)
		m.needsSemicolon = false
		for _, clause := range stmt.List {
			clause.List = m.optimizeStmtList(clause.List, defaultBlock)
		}
		m.renamer.renameScope(stmt.Scope)
		for _, clause := range stmt.List {
			m.writeSemicolon()
			m.write(clause.TokenType.Bytes())
			if clause.Cond != nil {
				m.writeSpaceBeforeIdent()
				m.minifyExpr(clause.Cond, js.OpExpr)
			}
			m.write(colonBytes)
			for _, item := range clause.List {
				m.writeSemicolon()
				m.minifyStmt(item)
			}
		}
		m.write(closeBraceBytes)
		m.needsSemicolon = false
	case *js.ThrowStmt:
		m.write(throwBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(stmt.Value, js.OpExpr)
		m.requireSemicolon()
	case *js.TryStmt:
		m.write(tryBytes)
		stmt.Body.List = m.optimizeStmtList(stmt.Body.List, defaultBlock)
		m.renamer.renameScope(stmt.Body.Scope)
		m.minifyBlockStmt(stmt.Body)
		if stmt.Catch != nil {
			m.write(catchBytes)
			stmt.Catch.List = m.optimizeStmtList(stmt.Catch.List, defaultBlock)
			m.renamer.renameScope(stmt.Catch.Scope)
			if stmt.Binding != nil {
				m.write(openParenBytes)
				m.minifyBinding(stmt.Binding)
				m.write(closeParenBytes)
			}
			m.minifyBlockStmt(stmt.Catch)
		}
		if stmt.Finally != nil {
			m.write(finallyBytes)
			stmt.Finally.List = m.optimizeStmtList(stmt.Finally.List, defaultBlock)
			m.renamer.renameScope(stmt.Finally.Scope)
			m.minifyBlockStmt(stmt.Finally)
		}
	case *js.FuncDecl:
		m.minifyFuncDecl(stmt, false)
	case *js.ClassDecl:
		m.minifyClassDecl(stmt)
	case *js.DebuggerStmt:
		m.write(debuggerBytes)
		m.requireSemicolon()
	case *js.EmptyStmt:
	case *js.ImportStmt:
		m.write(importBytes)
		if stmt.Default != nil {
			m.write(spaceBytes)
			m.write(stmt.Default)
			if len(stmt.List) != 0 {
				m.write(commaBytes)
			} else if stmt.Default != nil || len(stmt.List) != 0 {
				m.write(spaceBytes)
			}
		}
		if len(stmt.List) == 1 && len(stmt.List[0].Name) == 1 && stmt.List[0].Name[0] == '*' {
			m.writeSpaceBeforeIdent()
			m.minifyAlias(stmt.List[0])
			if stmt.Default != nil || len(stmt.List) != 0 {
				m.write(spaceBytes)
			}
		} else if 0 < len(stmt.List) {
			m.write(openBraceBytes)
			for i, item := range stmt.List {
				if i != 0 {
					m.write(commaBytes)
				}
				m.minifyAlias(item)
			}
			m.write(closeBraceBytes)
		}
		if stmt.Default != nil || len(stmt.List) != 0 {
			m.write(fromBytes)
		}
		m.write(stmt.Module)
		m.requireSemicolon()
	case *js.ExportStmt:
		m.write(exportBytes)
		if stmt.Decl != nil {
			if stmt.Default {
				m.write(spaceDefaultBytes)
				m.writeSpaceBeforeIdent()
				m.minifyExpr(stmt.Decl, js.OpAssign)
				_, isHoistable := stmt.Decl.(*js.FuncDecl)
				_, isClass := stmt.Decl.(*js.ClassDecl)
				if !isHoistable && !isClass {
					m.requireSemicolon()
				}
			} else {
				m.writeSpaceBeforeIdent()
				m.minifyStmt(stmt.Decl.(js.IStmt)) // can only be variable, function, or class decl
			}
		} else {
			if len(stmt.List) == 1 && (len(stmt.List[0].Name) == 1 && stmt.List[0].Name[0] == '*' || stmt.List[0].Name == nil && len(stmt.List[0].Binding) == 1 && stmt.List[0].Binding[0] == '*') {
				m.writeSpaceBeforeIdent()
				m.minifyAlias(stmt.List[0])
				if stmt.Module != nil && stmt.List[0].Name != nil {
					m.write(spaceBytes)
				}
			} else if 0 < len(stmt.List) {
				m.write(openBraceBytes)
				for i, item := range stmt.List {
					if i != 0 {
						m.write(commaBytes)
					}
					m.minifyAlias(item)
				}
				m.write(closeBraceBytes)
			}
			if stmt.Module != nil {
				m.write(fromBytes)
				m.write(stmt.Module)
			}
			m.requireSemicolon()
		}
	case *js.DirectivePrologueStmt:
		m.write(stmt.Value)
		m.requireSemicolon()
	}
}

func (m *jsMinifier) minifyBlockStmt(stmt *js.BlockStmt) {
	m.write(openBraceBytes)
	m.needsSemicolon = false
	for _, item := range stmt.List {
		m.writeSemicolon()
		m.minifyStmt(item)
	}

	m.write(closeBraceBytes)
	m.needsSemicolon = false
}

func (m *jsMinifier) minifyBlockAsStmt(blockStmt *js.BlockStmt, blockType blockType) {
	// minify block when statement is expected, i.e. semicolon if empty or remove braces for single statement
	// assume we already renamed the scope
	hasLexicalVars := false
	for _, v := range blockStmt.Scope.Declared[blockStmt.Scope.NumForInit:] {
		if v.Decl == js.LexicalDecl {
			hasLexicalVars = true
			break
		}
	}
	if 1 < len(blockStmt.List) || hasLexicalVars {
		m.minifyBlockStmt(blockStmt)
	} else if len(blockStmt.List) == 1 {
		m.minifyStmt(blockStmt.List[0])
	} else {
		m.write(semicolonBytes)
		m.needsSemicolon = false
	}
}

func (m *jsMinifier) minifyStmtOrBlock(i js.IStmt, blockType blockType) {
	// minify stmt or a block
	if blockStmt, ok := i.(*js.BlockStmt); ok {
		blockStmt.List = m.optimizeStmtList(blockStmt.List, blockType)
		m.renamer.renameScope(blockStmt.Scope)
		m.minifyBlockAsStmt(blockStmt, blockType)
	} else {
		// optimizeStmtList can in some cases expand one stmt to two shorter stmts
		list := m.optimizeStmtList([]js.IStmt{i}, blockType)
		if len(list) == 1 {
			m.minifyStmt(list[0])
		} else if len(list) == 0 {
			m.write(semicolonBytes)
			m.needsSemicolon = false
		} else {
			m.minifyBlockStmt(&js.BlockStmt{List: list, Scope: js.Scope{}})
		}
	}
}

func (m *jsMinifier) minifyAlias(alias js.Alias) {
	if alias.Name != nil {
		m.write(alias.Name)
		if !bytes.Equal(alias.Name, starBytes) {
			m.write(spaceBytes)
		}
		m.write(asSpaceBytes)
	}
	if alias.Binding != nil {
		m.write(alias.Binding)
	}
}

func (m *jsMinifier) minifyParams(params js.Params) {
	m.write(openParenBytes)
	for i, item := range params.List {
		if i != 0 {
			m.write(commaBytes)
		}
		m.minifyBindingElement(item)
	}
	if params.Rest != nil {
		if len(params.List) != 0 {
			m.write(commaBytes)
		}
		m.write(ellipsisBytes)
		m.minifyBinding(params.Rest)
	}
	m.write(closeParenBytes)
}

func (m *jsMinifier) minifyArguments(args js.Args) {
	m.write(openParenBytes)
	for i, item := range args.List {
		if i != 0 {
			m.write(commaBytes)
		}
		if item.Rest {
			m.write(ellipsisBytes)
		}
		m.minifyExpr(item.Value, js.OpAssign)
	}
	m.write(closeParenBytes)
}

func (m *jsMinifier) minifyVarDecl(decl *js.VarDecl, onlyDefines bool) {
	if decl.TokenType == js.VarToken && m.varsHoisted != nil && decl != m.varsHoisted {
		// remove 'var' when hoisting variables
		first := true
		for _, item := range decl.List {
			if item.Default != nil || !onlyDefines {
				if !first {
					m.write(commaBytes)
				}
				m.minifyBindingElement(item)
				first = false
			}
		}
	} else {
		m.write(decl.TokenType.Bytes())
		m.writeSpaceBeforeIdent()
		for i, item := range decl.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyBindingElement(item)
		}
	}
}

func (m *jsMinifier) minifyFuncDecl(decl *js.FuncDecl, inExpr bool) {
	parentRename := m.renamer.rename
	m.renamer.rename = !decl.Body.Scope.HasWith && !m.o.KeepVarNames
	parentVarsHoisted := m.hoistVars(&decl.Body)

	decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)

	if decl.Async {
		m.write(asyncSpaceBytes)
	}
	m.write(functionBytes)
	if decl.Generator {
		m.write(starBytes)
	}
	if inExpr {
		m.renamer.renameScope(decl.Body.Scope)
	}
	if decl.Name != nil && (!inExpr || 1 < decl.Name.Uses) {
		if !decl.Generator {
			m.write(spaceBytes)
		}
		m.write(decl.Name.Data)
	}
	if !inExpr {
		m.renamer.renameScope(decl.Body.Scope)
	}
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(&decl.Body)

	m.varsHoisted = parentVarsHoisted
	m.renamer.rename = parentRename
}

func (m *jsMinifier) minifyMethodDecl(decl *js.MethodDecl) {
	parentRename := m.renamer.rename
	m.renamer.rename = !decl.Body.Scope.HasWith && !m.o.KeepVarNames
	parentVarsHoisted := m.hoistVars(&decl.Body)

	decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)

	if decl.Static {
		m.write(staticBytes)
		m.writeSpaceBeforeIdent()
	}
	if decl.Async {
		m.write(asyncBytes)
		if decl.Generator {
			m.write(starBytes)
		} else {
			m.writeSpaceBeforeIdent()
		}
	} else if decl.Generator {
		m.write(starBytes)
	} else if decl.Get {
		m.write(getBytes)
		m.writeSpaceBeforeIdent()
	} else if decl.Set {
		m.write(setBytes)
		m.writeSpaceBeforeIdent()
	}
	m.minifyPropertyName(decl.Name)
	m.renamer.renameScope(decl.Body.Scope)
	m.minifyParams(decl.Params)
	m.minifyBlockStmt(&decl.Body)

	m.varsHoisted = parentVarsHoisted
	m.renamer.rename = parentRename
}

func (m *jsMinifier) minifyArrowFunc(decl *js.ArrowFunc) {
	parentRename := m.renamer.rename
	m.renamer.rename = !decl.Body.Scope.HasWith && !m.o.KeepVarNames
	parentVarsHoisted := m.hoistVars(&decl.Body)

	decl.Body.List = m.optimizeStmtList(decl.Body.List, functionBlock)

	m.renamer.renameScope(decl.Body.Scope)
	if decl.Async {
		m.write(asyncBytes)
	}
	removeParens := false
	if decl.Params.Rest == nil && len(decl.Params.List) == 1 && decl.Params.List[0].Default == nil {
		if decl.Params.List[0].Binding == nil {
			removeParens = true
		} else if _, ok := decl.Params.List[0].Binding.(*js.Var); ok {
			removeParens = true
		}
	}
	if removeParens {
		if decl.Async && decl.Params.List[0].Binding != nil {
			// add space after async in: async a => ...
			m.write(spaceBytes)
		}
		m.minifyBindingElement(decl.Params.List[0])
	} else {
		parentInFor := m.inFor
		m.inFor = false
		m.minifyParams(decl.Params)
		m.inFor = parentInFor
	}
	m.write(arrowBytes)
	removeBraces := false
	if 0 < len(decl.Body.List) {
		returnStmt, isReturn := decl.Body.List[len(decl.Body.List)-1].(*js.ReturnStmt)
		if isReturn && returnStmt.Value != nil {
			// merge expression statements to final return statement, remove function body braces
			var list []js.IExpr
			removeBraces = true
			for _, item := range decl.Body.List[:len(decl.Body.List)-1] {
				if expr, isExpr := item.(*js.ExprStmt); isExpr {
					list = append(list, expr.Value)
				} else {
					removeBraces = false
					break
				}
			}
			if removeBraces {
				list = append(list, returnStmt.Value)
				expr := list[0]
				for _, right := range list[1:] {
					expr = &js.BinaryExpr{Op: js.CommaToken, X: expr, Y: right}
				}
				if 0 < len(list) {
					expr = &js.GroupExpr{X: expr}
				}
				m.expectExpr = expectExprBody
				m.minifyExpr(expr, js.OpAssign)
				if m.groupedStmt {
					m.write(closeParenBytes)
					m.groupedStmt = false
				}
			}
		} else if isReturn && returnStmt.Value == nil {
			// remove empty return
			decl.Body.List = decl.Body.List[:len(decl.Body.List)-1]
		}
	}
	if !removeBraces {
		m.minifyBlockStmt(&decl.Body)
	}

	m.varsHoisted = parentVarsHoisted
	m.renamer.rename = parentRename
}

func (m *jsMinifier) minifyClassDecl(decl *js.ClassDecl) {
	m.write(classBytes)
	if decl.Name != nil {
		m.write(spaceBytes)
		m.write(decl.Name.Data)
	}
	if decl.Extends != nil {
		m.write(spaceExtendsBytes)
		m.writeSpaceBeforeIdent()
		m.minifyExpr(decl.Extends, js.OpLHS)
	}
	m.write(openBraceBytes)
	m.needsSemicolon = false
	for _, item := range decl.Definitions {
		m.writeSemicolon()
		m.minifyPropertyName(item.Name)
		if item.Init != nil {
			m.write(equalBytes)
			m.minifyExpr(item.Init, js.OpAssign)
		}
		m.requireSemicolon()
	}
	if 0 < len(decl.Methods) {
		m.writeSemicolon()
		for _, item := range decl.Methods {
			m.minifyMethodDecl(item)
		}
	}
	m.write(closeBraceBytes)
	m.needsSemicolon = false
}

func (m *jsMinifier) minifyPropertyName(name js.PropertyName) {
	if name.IsComputed() {
		m.write(openBracketBytes)
		m.minifyExpr(name.Computed, js.OpAssign)
		m.write(closeBracketBytes)
	} else {
		m.write(name.Literal.Data)
	}
}

func (m *jsMinifier) minifyProperty(property js.Property) {
	// property.Name is always set in ObjectLiteral
	if property.Spread {
		m.write(ellipsisBytes)
	} else if v, ok := property.Value.(*js.Var); property.Name != nil && (!ok || !property.Name.IsIdent(v.Name())) {
		// add 'old-name:' before BindingName as the latter will be renamed
		m.minifyPropertyName(*property.Name)
		m.write(colonBytes)
	}
	m.minifyExpr(property.Value, js.OpAssign)
	if property.Init != nil {
		m.write(equalBytes)
		m.minifyExpr(property.Init, js.OpAssign)
	}
}

func (m *jsMinifier) minifyBindingElement(element js.BindingElement) {
	if element.Binding != nil {
		parentInFor := m.inFor
		m.inFor = false
		m.minifyBinding(element.Binding)
		m.inFor = parentInFor
		if element.Default != nil {
			m.write(equalBytes)
			m.minifyExpr(element.Default, js.OpAssign)
		}
	}
}

func (m *jsMinifier) minifyBinding(ibinding js.IBinding) {
	switch binding := ibinding.(type) {
	case *js.Var:
		m.write(binding.Data)
	case *js.BindingArray:
		m.write(openBracketBytes)
		for i, item := range binding.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyBindingElement(item)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write(commaBytes)
			}
			m.write(ellipsisBytes)
			m.minifyBinding(binding.Rest)
		}
		m.write(closeBracketBytes)
	case *js.BindingObject:
		m.write(openBraceBytes)
		for i, item := range binding.List {
			if i != 0 {
				m.write(commaBytes)
			}
			// item.Key is always set
			if item.Key.IsComputed() {
				m.minifyPropertyName(*item.Key)
				m.write(colonBytes)
			} else if v, ok := item.Value.Binding.(*js.Var); !ok || !item.Key.IsIdent(v.Data) {
				// add 'old-name:' before BindingName as the latter will be renamed
				m.minifyPropertyName(*item.Key)
				m.write(colonBytes)
			}
			m.minifyBindingElement(item.Value)
		}
		if binding.Rest != nil {
			if 0 < len(binding.List) {
				m.write(commaBytes)
			}
			m.write(ellipsisBytes)
			m.write(binding.Rest.Data)
		}
		m.write(closeBraceBytes)
	}
}

func (m *jsMinifier) minifyBinaryExpr(expr *js.BinaryExpr) bool {
	if lit, ok := expr.Y.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken && expr.Op == js.AddToken {
		// merge strings that are added together
		n := len(lit.Data) - 2
		strings := []*js.LiteralExpr{lit}

		left := expr
		for {
			if 50 < len(strings) {
				return false // limit recursion
			} else if lit, ok := left.X.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken {
				n += len(lit.Data) - 2
				strings = append(strings, lit)
				break
			} else if left, ok = left.X.(*js.BinaryExpr); ok && left.Op == js.AddToken {
				if lit, ok := left.Y.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken {
					n += len(lit.Data) - 2
					strings = append(strings, lit)
					continue
				}
			}
			return false
		}

		b := make([]byte, 0, n+2)
		b = append(b, strings[len(strings)-1].Data[:len(strings[len(strings)-1].Data)-1]...)
		for i := len(strings) - 2; 0 < i; i-- {
			b = append(b, strings[i].Data[1:len(strings[i].Data)-1]...)
		}
		b = append(b, strings[0].Data[1:]...)
		b[len(b)-1] = b[0]

		// unescaped quotes will be repaired in minifyString
		m.write(minifyString(b))
		return true
	}
	return false
}

func (m *jsMinifier) minifyExpr(i js.IExpr, prec js.OpPrec) {
	switch expr := i.(type) {
	case *js.Var:
		for expr.Link != nil {
			expr = expr.Link
		}
		data := expr.Data
		if bytes.Equal(data, undefinedBytes) { // TODO: only if not defined
			if js.OpUnary < prec {
				m.write(groupedVoidZeroBytes)
			} else {
				m.write(voidZeroBytes)
			}
		} else if bytes.Equal(data, infinityBytes) { // TODO: only if not defined
			if js.OpMul < prec {
				m.write(groupedOneDivZeroBytes)
			} else {
				m.write(oneDivZeroBytes)
			}
		} else {
			m.write(data)
		}
	case *js.LiteralExpr:
		if expr.TokenType == js.DecimalToken {
			m.write(minify.Number(expr.Data, m.o.Precision))
		} else if expr.TokenType == js.BinaryToken {
			m.write(binaryNumber(expr.Data, m.o.Precision))
		} else if expr.TokenType == js.OctalToken {
			m.write(octalNumber(expr.Data, m.o.Precision))
		} else if expr.TokenType == js.HexadecimalToken {
			m.write(hexadecimalNumber(expr.Data, m.o.Precision))
		} else if expr.TokenType == js.TrueToken {
			if js.OpUnary < prec {
				m.write(groupedNotZeroBytes)
			} else {
				m.write(notZeroBytes)
			}
		} else if expr.TokenType == js.FalseToken {
			if js.OpUnary < prec {
				m.write(groupedNotOneBytes)
			} else {
				m.write(notOneBytes)
			}
		} else if expr.TokenType == js.StringToken {
			m.write(minifyString(expr.Data))
		} else if expr.TokenType == js.RegExpToken {
			// </script>/ => < /script>/
			if 0 < len(m.prev) && m.prev[len(m.prev)-1] == '<' && bytes.HasPrefix(expr.Data, regExpScriptBytes) {
				m.write(spaceBytes)
			}
			m.write(expr.Data)
		} else {
			m.write(expr.Data)
		}
	case *js.BinaryExpr:
		if m.minifyBinaryExpr(expr) {
			break
		}

		precLeft := binaryLeftPrecMap[expr.Op]
		// convert (a,b)&&c into a,b&&c but not a=(b,c)&&d into a=(b,c&&d)
		if prec <= js.OpExpr {
			if group, ok := expr.X.(*js.GroupExpr); ok {
				if binary, ok := group.X.(*js.BinaryExpr); ok && binary.Op == js.CommaToken && js.OpAnd <= exprPrec(binary.Y) {
					expr.X = group.X
					precLeft = js.OpExpr
				}
			}
		}
		if expr.Op == js.InstanceofToken || expr.Op == js.InToken {
			group := expr.Op == js.InToken && m.inFor
			if group {
				m.write(openParenBytes)
			}
			m.minifyExpr(expr.X, precLeft)
			m.writeSpaceAfterIdent()
			m.write(expr.Op.Bytes())
			m.writeSpaceBeforeIdent()
			m.minifyExpr(expr.Y, binaryRightPrecMap[expr.Op])
			if group {
				m.write(closeParenBytes)
			}
		} else {
			m.minifyExpr(expr.X, precLeft)
			// 0 < len(m.prev) always
			if expr.Op == js.GtToken && m.prev[len(m.prev)-1] == '-' {
				m.write(spaceBytes)
			} else if expr.Op == js.EqEqEqToken || expr.Op == js.NotEqEqToken {
				if left, ok := expr.X.(*js.UnaryExpr); ok && left.Op == js.TypeofToken {
					if right, ok := expr.Y.(*js.LiteralExpr); ok && right.TokenType == js.StringToken {
						if expr.Op == js.EqEqEqToken {
							expr.Op = js.EqEqToken
						} else {
							expr.Op = js.NotEqToken
						}
					}
				} else if right, ok := expr.Y.(*js.UnaryExpr); ok && right.Op == js.TypeofToken {
					if left, ok := expr.X.(*js.LiteralExpr); ok && left.TokenType == js.StringToken {
						if expr.Op == js.EqEqEqToken {
							expr.Op = js.EqEqToken
						} else {
							expr.Op = js.NotEqToken
						}
					}
				}
			}
			m.write(expr.Op.Bytes())
			if expr.Op == js.AddToken {
				// +++  =>  + ++
				m.writeSpaceBefore('+')
			} else if expr.Op == js.SubToken {
				// ---  =>  - --
				m.writeSpaceBefore('-')
			} else if expr.Op == js.DivToken {
				// //  =>  / /
				m.writeSpaceBefore('/')
			}
			m.minifyExpr(expr.Y, binaryRightPrecMap[expr.Op])
		}
	case *js.UnaryExpr:
		if expr.Op == js.PostIncrToken || expr.Op == js.PostDecrToken {
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
			m.write(expr.Op.Bytes())
		} else {
			isLtNot := expr.Op == js.NotToken && 0 < len(m.prev) && m.prev[len(m.prev)-1] == '<'
			m.write(expr.Op.Bytes())
			if expr.Op == js.DeleteToken || expr.Op == js.VoidToken || expr.Op == js.TypeofToken || expr.Op == js.AwaitToken {
				m.writeSpaceBeforeIdent()
			} else if expr.Op == js.PosToken {
				// +++  =>  + ++
				m.writeSpaceBefore('+')
			} else if expr.Op == js.NegToken || isLtNot {
				// ---  =>  - --
				// <!--  =>  <! --
				m.writeSpaceBefore('-')
			} else if expr.Op == js.NotToken {
				if lit, ok := expr.X.(*js.LiteralExpr); ok && (lit.TokenType == js.StringToken || lit.TokenType == js.RegExpToken) {
					// !"string"  =>  !1
					m.write(oneBytes)
					break
				} else if ok && lit.TokenType == js.DecimalToken {
					// !123  =>  !1 (except for !0)
					if num := minify.Number(lit.Data, m.o.Precision); len(num) == 1 && num[0] == '0' {
						m.write(zeroBytes)
					} else {
						m.write(oneBytes)
					}
					break
				}
			}
			m.minifyExpr(expr.X, unaryPrecMap[expr.Op])
		}
	case *js.DotExpr:
		if group, ok := expr.X.(*js.GroupExpr); ok {
			if lit, ok := group.X.(*js.LiteralExpr); ok && lit.TokenType == js.DecimalToken {
				num := minify.Number(lit.Data, m.o.Precision)
				isInt := true
				for _, c := range num {
					if c == '.' || c == 'e' || c == 'E' {
						isInt = false
						break
					}
				}
				if isInt {
					m.write(num)
					m.write(dotBytes)
				} else {
					m.write(num)
				}
				m.write(dotBytes)
				m.write(expr.Y.Data)
				break
			}
		}
		if prec < js.OpMember {
			m.minifyExpr(expr.X, js.OpCall)
		} else {
			m.minifyExpr(expr.X, js.OpMember)
		}
		// 0 < len(m.prev) always
		if last := m.prev[len(m.prev)-1]; '0' <= last && last <= '9' {
			isInteger := true
			for _, c := range m.prev[:len(m.prev)-1] {
				if c < '0' || '9' < c {
					isInteger = false
					break
				}
			}
			if isInteger {
				// prevent previous integer
				m.write(dotBytes)
			}
		}
		m.write(dotBytes)
		m.write(expr.Y.Data)
	case *js.GroupExpr:
		precInside := exprPrec(expr.X)
		if prec <= precInside || precInside == js.OpCoalesce && prec == js.OpBitOr {
			m.minifyExpr(expr.X, prec)
		} else {
			parentInFor := m.inFor
			m.inFor = false
			m.write(openParenBytes)
			m.minifyExpr(expr.X, js.OpExpr)
			m.write(closeParenBytes)
			m.inFor = parentInFor
		}
	case *js.ArrayExpr:
		parentInFor := m.inFor
		m.inFor = false
		m.write(openBracketBytes)
		for i, item := range expr.List {
			if i != 0 {
				m.write(commaBytes)
			}
			if item.Spread {
				m.write(ellipsisBytes)
			}
			m.minifyExpr(item.Value, js.OpAssign)
		}
		if 0 < len(expr.List) && expr.List[len(expr.List)-1].Value == nil {
			m.write(commaBytes)
		}
		m.write(closeBracketBytes)
		m.inFor = parentInFor
	case *js.ObjectExpr:
		parentInFor := m.inFor
		m.inFor = false
		if m.expectExpr != expectAny {
			m.write(openParenBracketBytes)
			m.groupedStmt = true
		} else {
			m.write(openBraceBytes)
		}
		for i, item := range expr.List {
			if i != 0 {
				m.write(commaBytes)
			}
			m.minifyProperty(item)
		}
		m.write(closeBraceBytes)
		m.inFor = parentInFor
	case *js.TemplateExpr:
		if expr.Tag != nil {
			if prec < js.OpMember {
				m.minifyExpr(expr.Tag, js.OpCall)
			} else {
				m.minifyExpr(expr.Tag, js.OpMember)
			}
		}
		parentInFor := m.inFor
		m.inFor = false
		for _, item := range expr.List {
			m.write(item.Value)
			m.minifyExpr(item.Expr, js.OpExpr)
		}
		m.write(expr.Tail)
		m.inFor = parentInFor
	case *js.NewExpr:
		if expr.Args == nil && js.OpLHS < prec && prec != js.OpNew {
			m.write(openNewBytes)
			m.writeSpaceBeforeIdent()
			m.minifyExpr(expr.X, js.OpNew)
			m.write(closeParenBytes)
		} else {
			m.write(newBytes)
			m.writeSpaceBeforeIdent()
			if expr.Args != nil {
				m.minifyExpr(expr.X, js.OpMember)
				m.minifyArguments(*expr.Args)
			} else {
				m.minifyExpr(expr.X, js.OpNew)
			}
		}
	case *js.NewTargetExpr:
		m.write(newTargetBytes)
		m.writeSpaceBeforeIdent()
	case *js.ImportMetaExpr:
		if m.expectExpr == expectExprStmt {
			m.write(openParenBytes)
			m.groupedStmt = true
		}
		m.write(importMetaBytes)
		m.writeSpaceBeforeIdent()
	case *js.YieldExpr:
		m.write(yieldBytes)
		m.writeSpaceBeforeIdent()
		if expr.X != nil {
			if expr.Generator {
				m.write(starBytes)
				m.minifyExpr(expr.X, js.OpAssign)
			} else if v, ok := expr.X.(*js.Var); !ok || !bytes.Equal(v.Name(), undefinedBytes) { // TODO: only if not defined
				m.minifyExpr(expr.X, js.OpAssign)
			}
		}
	case *js.CallExpr:
		m.minifyExpr(expr.X, js.OpCall)
		parentInFor := m.inFor
		m.inFor = false
		m.minifyArguments(expr.Args)
		m.inFor = parentInFor
	case *js.IndexExpr:
		if m.expectExpr == expectExprStmt {
			if v, ok := expr.X.(*js.Var); ok && bytes.Equal(v.Name(), letBytes) {
				m.write(notBytes)
			}
		}
		if prec < js.OpMember {
			m.minifyExpr(expr.X, js.OpCall)
		} else {
			m.minifyExpr(expr.X, js.OpMember)
		}
		if lit, ok := expr.Y.(*js.LiteralExpr); ok && lit.TokenType == js.StringToken && 2 < len(lit.Data) {
			if isIdent := js.AsIdentifierName(lit.Data[1 : len(lit.Data)-1]); isIdent {
				m.write(dotBytes)
				m.write(lit.Data[1 : len(lit.Data)-1])
				break
			} else if isNum := js.AsDecimalLiteral(lit.Data[1 : len(lit.Data)-1]); isNum {
				m.write(openBracketBytes)
				m.write(minify.Number(lit.Data[1:len(lit.Data)-1], 0))
				m.write(closeBracketBytes)
				break
			}
		}
		parentInFor := m.inFor
		m.inFor = false
		m.write(openBracketBytes)
		m.minifyExpr(expr.Y, js.OpExpr)
		m.write(closeBracketBytes)
		m.inFor = parentInFor
	case *js.CondExpr:
		// remove double negative !! in condition, or switch cases for single negative !
		if unary1, ok := expr.Cond.(*js.UnaryExpr); ok && unary1.Op == js.NotToken {
			if unary2, ok := unary1.X.(*js.UnaryExpr); ok && unary2.Op == js.NotToken {
				if isBooleanExpr(unary2.X) {
					expr.Cond = unary2.X
				}
			} else {
				expr.Cond = unary1.X
				expr.X, expr.Y = expr.Y, expr.X
			}
		}

		finalCond := finalExpr(expr.Cond)
		if truthy, ok := m.isTruthy(expr.Cond); truthy && ok {
			// if condition is truthy
			m.minifyExpr(expr.X, prec)
		} else if !truthy && ok {
			// if condition is falsy
			m.minifyExpr(expr.Y, prec)
		} else if m.isEqualExpr(finalCond, expr.X) && (exprPrec(finalCond) < js.OpAssign || binaryLeftPrecMap[js.OrToken] <= exprPrec(finalCond)) && (exprPrec(expr.Y) < js.OpAssign || binaryRightPrecMap[js.OrToken] <= exprPrec(expr.Y)) {
			// if condition is equal to true body
			// for higher prec we need to add group parenthesis, and for lower prec we have parenthesis anyways. This only is shorter if len(expr.X) >= 3. isEqualExpr only checks for literal variables, which is a name will be minified to a one or two character name.
			m.minifyExpr(&js.BinaryExpr{js.OrToken, groupExpr(expr.Cond, binaryLeftPrecMap[js.OrToken]), expr.Y}, prec)
		} else if m.isEqualExpr(finalCond, expr.Y) && (exprPrec(finalCond) < js.OpAssign || binaryLeftPrecMap[js.AndToken] <= exprPrec(finalCond)) && (exprPrec(expr.X) < js.OpAssign || binaryRightPrecMap[js.AndToken] <= exprPrec(expr.X)) {
			// if condition is equal to false body
			// for higher prec we need to add group parenthesis, and for lower prec we have parenthesis anyways. This only is shorter if len(expr.X) >= 3. isEqualExpr only checks for literal variables, which is a name will be minified to a one or two character name.
			m.minifyExpr(&js.BinaryExpr{js.AndToken, groupExpr(expr.Cond, binaryLeftPrecMap[js.AndToken]), expr.X}, prec)
		} else if m.isEqualExpr(expr.X, expr.Y) {
			// if true and false bodies are equal
			if prec <= js.OpExpr {
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write(commaBytes)
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
			} else {
				m.write(openParenBytes)
				m.minifyExpr(expr.Cond, binaryLeftPrecMap[js.CommaToken])
				m.write(commaBytes)
				m.minifyExpr(expr.X, binaryRightPrecMap[js.CommaToken])
				m.write(closeParenBytes)
			}
			//} else if left, right, ok := m.toNullishExpr(expr); ok {
			//	// no need to check whether left/right need to add groups, as the space saving is always more
			//	m.minifyExpr(groupExpr(left, binaryLeftPrecMap[js.NullishToken]), binaryLeftPrecMap[js.NullishToken])
			//	m.write(nullishBytes)
			//	m.minifyExpr(groupExpr(right, binaryRightPrecMap[js.NullishToken]), binaryRightPrecMap[js.NullishToken])
		} else {
			// shorten when true and false bodies are true and false
			trueX, falseX := m.isTrue(expr.X), m.isFalse(expr.X)
			trueY, falseY := m.isTrue(expr.Y), m.isFalse(expr.Y)
			if trueX && falseY || falseX && trueY {
				m.minifyBooleanExpr(expr.Cond, falseX, prec)
			} else if trueX || trueY {
				// trueX != trueY
				m.minifyBooleanExpr(expr.Cond, trueY, binaryLeftPrecMap[js.OrToken])
				m.write(orBytes)
				if trueY {
					m.minifyExpr(&js.GroupExpr{X: expr.X}, binaryRightPrecMap[js.OrToken])
				} else {
					m.minifyExpr(&js.GroupExpr{X: expr.Y}, binaryRightPrecMap[js.OrToken])
				}
			} else if falseX || falseY {
				// falseX != falseY
				m.minifyBooleanExpr(expr.Cond, falseX, binaryLeftPrecMap[js.AndToken])
				m.write(andBytes)
				if falseX {
					m.minifyExpr(&js.GroupExpr{X: expr.Y}, binaryRightPrecMap[js.AndToken])
				} else {
					m.minifyExpr(&js.GroupExpr{X: expr.X}, binaryRightPrecMap[js.AndToken])
				}
			} else if condExpr, ok := expr.X.(*js.CondExpr); ok && m.isEqualExpr(expr.Y, condExpr.Y) {
				// nested conditional expression with same false bodies
				m.minifyExpr(&js.GroupExpr{X: expr.Cond}, binaryLeftPrecMap[js.AndToken])
				m.write(andBytes)
				m.minifyExpr(&js.GroupExpr{X: condExpr.Cond}, binaryRightPrecMap[js.AndToken])
				m.write(questionBytes)
				m.minifyExpr(condExpr.X, js.OpAssign)
				m.write(colonBytes)
				m.minifyExpr(expr.Y, js.OpAssign)
			} else {
				// regular conditional expression
				// convert  (a,b)?c:d  =>  a,b?c:d
				if prec <= js.OpExpr {
					if group, ok := expr.Cond.(*js.GroupExpr); ok {
						if binary, ok := group.X.(*js.BinaryExpr); ok && binary.Op == js.CommaToken && js.OpCoalesce <= exprPrec(binary.Y) {
							expr.Cond = group.X
						}
					}
				}
				m.minifyExpr(expr.Cond, js.OpCoalesce)
				m.write(questionBytes)
				m.minifyExpr(expr.X, js.OpAssign)
				m.write(colonBytes)
				m.minifyExpr(expr.Y, js.OpAssign)
			}
		}
	case *js.OptChainExpr:
		m.minifyExpr(expr.X, js.OpCall)
		m.write(optChainBytes)
		if callExpr, ok := expr.Y.(*js.CallExpr); ok {
			m.minifyArguments(callExpr.Args)
		} else if indexExpr, ok := expr.Y.(*js.IndexExpr); ok {
			m.write(openBracketBytes)
			m.minifyExpr(indexExpr.Y, js.OpExpr)
			m.write(closeBracketBytes)
		} else {
			m.minifyExpr(expr.Y, js.OpPrimary) // TemplateExpr or LiteralExpr
		}
	case *js.VarDecl:
		m.minifyVarDecl(expr, true) // happens in for statement or when vars were hoisted
	case *js.FuncDecl:
		grouped := m.expectExpr == expectExprStmt && prec != js.OpExpr
		if grouped {
			m.write(openParenBytes)
		} else if m.expectExpr == expectExprStmt {
			m.write(notBytes)
		}
		parentInFor := m.inFor
		m.inFor = false
		m.minifyFuncDecl(expr, true)
		m.inFor = parentInFor
		if grouped {
			m.write(closeParenBytes)
		}
	case *js.ArrowFunc:
		m.minifyArrowFunc(expr)
	case *js.MethodDecl:
		m.minifyMethodDecl(expr) // only happens in object literal
	case *js.ClassDecl:
		if m.expectExpr == expectExprStmt {
			m.write(notBytes)
		}
		parentInFor := m.inFor
		m.inFor = false
		m.minifyClassDecl(expr)
		m.inFor = parentInFor
	}
}

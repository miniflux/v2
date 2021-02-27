package js

import (
	"bytes"
	"sort"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
)

type renamer struct {
	ast      *js.AST
	reserved map[string]struct{}
	rename   bool
}

func newRenamer(ast *js.AST, undeclared js.VarArray, rename bool) *renamer {
	reserved := make(map[string]struct{}, len(js.Keywords))
	for name := range js.Keywords {
		reserved[name] = struct{}{}
	}
	return &renamer{
		ast:      ast,
		reserved: reserved,
		rename:   rename,
	}
}

func (r *renamer) renameScope(scope js.Scope) {
	if !r.rename {
		return
	}

	rename := []byte("`") // so that the next is 'a'
	sort.Sort(js.VarsByUses(scope.Declared))
	for _, v := range scope.Declared {
		rename = r.next(rename)
		for r.isReserved(rename, scope.Undeclared) {
			rename = r.next(rename)
		}
		v.Data = parse.Copy(rename)
	}
}

func (r *renamer) isReserved(name []byte, undeclared js.VarArray) bool {
	if 1 < len(name) { // there are no keywords or known globals that are one character long
		if _, ok := r.reserved[string(name)]; ok {
			return true
		}
	}
	for _, v := range undeclared {
		for v.Link != nil {
			v = v.Link
		}
		if bytes.Equal(v.Data, name) {
			return true
		}
	}
	return false
}

func (r *renamer) next(name []byte) []byte {
	// Generate new names for variables where the last character is (a-zA-Z$_) and others are (a-zA-Z).
	// Thus we can have 54 one-character names and 52*54=2808 two-character names for every branch leaf.
	// That is sufficient for virtually all input.
	if name[len(name)-1] == 'z' {
		name[len(name)-1] = 'A'
	} else if name[len(name)-1] == 'Z' {
		name[len(name)-1] = '_'
	} else if name[len(name)-1] == '_' {
		name[len(name)-1] = '$'
	} else if name[len(name)-1] == '$' {
		i := len(name) - 2
		for ; 0 <= i; i-- {
			if name[i] == 'Z' {
				continue // happens after 52*54=2808 variables
			} else if name[i] == 'z' {
				name[i] = 'A' // happens after 26*54=1404 variables
			} else {
				name[i]++
				break
			}
		}
		for j := i + 1; j < len(name); j++ {
			name[j] = 'a'
		}
		if i < 0 {
			name = append(name, 'a')
		}
	} else {
		name[len(name)-1]++
	}
	return name
}

////////////////////////////////////////////////////////////////

func bindingRefs(ibinding js.IBinding) (refs []*js.Var) {
	switch binding := ibinding.(type) {
	case *js.Var:
		refs = append(refs, binding)
	case *js.BindingArray:
		for _, item := range binding.List {
			if item.Binding != nil {
				refs = append(refs, bindingRefs(item.Binding)...)
			}
		}
		if binding.Rest != nil {
			refs = append(refs, bindingRefs(binding.Rest)...)
		}
	case *js.BindingObject:
		for _, item := range binding.List {
			if item.Value.Binding != nil {
				refs = append(refs, bindingRefs(item.Value.Binding)...)
			}
		}
		if binding.Rest != nil {
			refs = append(refs, binding.Rest)
		}
	}
	return
}

func appendBindingVars(vars *[]*js.Var, binding js.IBinding) {
	if v, ok := binding.(*js.Var); ok {
		*vars = append(*vars, v)
	} else if array, ok := binding.(*js.BindingArray); ok {
		for _, item := range array.List {
			appendBindingVars(vars, item.Binding)
		}
		if array.Rest != nil {
			appendBindingVars(vars, array.Rest)
		}
	} else if object, ok := binding.(*js.BindingObject); ok {
		for _, item := range object.List {
			appendBindingVars(vars, item.Value.Binding)
		}
		if object.Rest != nil {
			*vars = append(*vars, object.Rest)
		}
	}
}

func addDefinition(decl *js.VarDecl, iDefines int, binding js.IBinding, value js.IExpr) bool {
	// see if not already defined in variable declaration list
	if vdef, ok := binding.(*js.Var); ok {
		for i, item := range decl.List[iDefines:] {
			if v, ok := item.Binding.(*js.Var); ok && v == vdef {
				decl.List[iDefines+i].Default = value
				if 0 < i {
					decl.List[iDefines], decl.List[iDefines+i] = decl.List[iDefines+i], decl.List[iDefines]
				}
				return true
			}
		}
	} else {
		vars := []*js.Var{}
		appendBindingVars(&vars, binding)
		if len(vars) == 0 {
			return false
		}
		locs := make([]int, len(vars))
		for i, vdef := range vars {
			locs[i] = -1
			for loc, item := range decl.List[iDefines:] {
				if v, ok := item.Binding.(*js.Var); ok && v == vdef {
					locs[i] = loc
					break
				}
			}
			if locs[i] == -1 {
				return false // cannot (probably) happen if we hoist variables
			}
		}
		sort.Ints(locs)
		if iDefines != locs[0] {
			decl.List[iDefines], decl.List[iDefines+locs[0]] = decl.List[iDefines+locs[0]], decl.List[iDefines]
		}
		decl.List[iDefines].Binding = binding
		decl.List[iDefines].Default = value
		for i := len(locs) - 1; 1 <= i; i-- {
			if locs[i] != locs[i-1] { // ignore duplicates, otherwise remove items from hoisted var declaration
				decl.List = append(decl.List[:locs[i]], decl.List[locs[i]+1:]...)
			}
		}
		return true
	}
	return false
}

func (m *jsMinifier) hoistVars(body *js.BlockStmt) *js.VarDecl {
	// Hoist all variable declarations in the current module/function scope to the top.
	// If the first statement is a var declaration, expand it. Otherwise prepend a new var declaration.
	// Except for the first var declaration, all others are converted to expressions. This is possible because an ArrayBindingPattern and ObjectBindingPattern can be converted to an ArrayLiteral or ObjectLiteral respectively, as they are supersets of the BindingPatterns.
	parentVarsHoisted := m.varsHoisted
	m.varsHoisted = nil
	if 1 < body.Scope.NumVarDecls {
		iDefines := 0 // position past last variable definition in declaration
		mergeStatements := true

		// ignore "use strict"
		declStart := 0
		for {
			if _, ok := body.List[declStart].(*js.DirectivePrologueStmt); ok {
				declStart++
			} else {
				break
			}
		}

		var decl *js.VarDecl
		if varDecl, ok := body.List[declStart].(*js.VarDecl); ok && varDecl.TokenType == js.VarToken {
			decl = varDecl
		} else if forStmt, ok := body.List[declStart].(*js.ForStmt); ok {
			// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
			if forStmt.Init == nil {
				decl = &js.VarDecl{TokenType: js.VarToken, List: nil}
				forStmt.Init = decl
			} else if varDecl, ok := forStmt.Init.(*js.VarDecl); ok && varDecl.TokenType == js.VarToken {
				decl = varDecl
			}
			mergeStatements = false
		} else if whileStmt, ok := body.List[declStart].(*js.WhileStmt); ok {
			// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
			decl = &js.VarDecl{TokenType: js.VarToken, List: nil}
			var forBody js.BlockStmt
			if blockStmt, ok := whileStmt.Body.(*js.BlockStmt); ok {
				forBody = *blockStmt
			} else {
				forBody.List = []js.IStmt{whileStmt.Body}
			}
			body.List[declStart] = &js.ForStmt{Init: decl, Cond: whileStmt.Cond, Post: nil, Body: forBody}
			mergeStatements = false
		}
		if decl != nil {
			// original declarations
			vs := []*js.Var{}
			for i, item := range decl.List {
				if item.Default != nil {
					iDefines = i + 1
				}
				vs = append(vs, bindingRefs(item.Binding)...)
			}

			// hoist other variable declarations in this function scope but don't initialize yet
		DeclaredLoop:
			for _, v := range body.Scope.Declared {
				if v.Decl == js.VariableDecl {
					for _, vdef := range vs {
						if v == vdef {
							continue DeclaredLoop
						}
					}
					//v.Uses++ // might be inaccurate as we remove non-defining variable declarations later on
					decl.List = append(decl.List, js.BindingElement{Binding: v, Default: nil})
				}
			}
		} else {
			decl = &js.VarDecl{TokenType: js.VarToken, List: nil}
			for _, v := range body.Scope.Declared {
				if v.Decl == js.VariableDecl {
					v.Uses++ // might be inaccurate as we remove non-defining variable declarations later on
					decl.List = append(decl.List, js.BindingElement{Binding: v, Default: nil})
				}
			}
			body.List = append(body.List[:declStart], append([]js.IStmt{decl}, body.List[declStart:]...)...)
		}

		if mergeStatements {
			// pull in assignments to variables into the declaration, e.g. var a;a=5  =>  var a=5
			// sort in order of definitions
			nMerged := 0
			declEnd := declStart + 1
		FindDefinitionsLoop:
			for k, item := range body.List[declEnd:] {
				if exprStmt, ok := item.(*js.ExprStmt); ok {
					if binaryExpr, ok := exprStmt.Value.(*js.BinaryExpr); ok && binaryExpr.Op == js.EqToken {
						if v, ok := binaryExpr.X.(*js.Var); ok && v.Decl == js.VariableDecl {
							if addDefinition(decl, iDefines, v, binaryExpr.Y) {
								iDefines++
								nMerged++
								continue
							}
						}
					}
				} else if varDecl, ok := item.(*js.VarDecl); ok && varDecl.TokenType == js.VarToken {
					for j := 0; j < len(varDecl.List); j++ {
						item := varDecl.List[j]
						if item.Default != nil {
							if addDefinition(decl, iDefines, item.Binding, item.Default) {
								iDefines++
								varDecl.List = append(varDecl.List[:j], varDecl.List[j+1:]...)
								j--
							} else {
								break FindDefinitionsLoop
							}
						}
						// declaration has no definition, that's fine as it's already merged previously
					}
					body.List[declEnd+k] = varDecl // update varDecl.List
					nMerged++
					continue // all variable declarations were matched, keep looking
				}
				break // not ExprStmt nor VarDecl
			}
			if 0 < nMerged {
				body.List = append(body.List[:declEnd], body.List[declEnd+nMerged:]...)
			}
		}
		m.varsHoisted = decl
	}
	return parentVarsHoisted
}

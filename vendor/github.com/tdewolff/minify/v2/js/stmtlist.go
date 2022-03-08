package js

import (
	"github.com/tdewolff/parse/v2/js"
)

func (m *jsMinifier) optimizeStmt(i js.IStmt) js.IStmt {
	// convert if/else into expression statement, and optimize blocks
	if ifStmt, ok := i.(*js.IfStmt); ok {
		hasIf := !m.isEmptyStmt(ifStmt.Body)
		hasElse := !m.isEmptyStmt(ifStmt.Else)
		if unaryExpr, ok := ifStmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken && hasElse {
			ifStmt.Cond = unaryExpr.X
			ifStmt.Body, ifStmt.Else = ifStmt.Else, ifStmt.Body
			hasIf, hasElse = hasElse, hasIf
		}
		if !hasIf && !hasElse {
			return &js.ExprStmt{Value: ifStmt.Cond}
		} else if hasIf && !hasElse {
			ifStmt.Body = m.optimizeStmt(ifStmt.Body)
			if X, isExprBody := ifStmt.Body.(*js.ExprStmt); isExprBody {
				if unaryExpr, ok := ifStmt.Cond.(*js.UnaryExpr); ok && unaryExpr.Op == js.NotToken {
					left := groupExpr(unaryExpr.X, binaryLeftPrecMap[js.OrToken])
					right := groupExpr(X.Value, binaryRightPrecMap[js.OrToken])
					return &js.ExprStmt{Value: &js.BinaryExpr{Op: js.OrToken, X: left, Y: right}}
				}
				left := groupExpr(ifStmt.Cond, binaryLeftPrecMap[js.AndToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.AndToken])
				return &js.ExprStmt{Value: &js.BinaryExpr{Op: js.AndToken, X: left, Y: right}}
			}
		} else if !hasIf && hasElse {
			ifStmt.Else = m.optimizeStmt(ifStmt.Else)
			if X, isExprElse := ifStmt.Else.(*js.ExprStmt); isExprElse {
				left := groupExpr(ifStmt.Cond, binaryLeftPrecMap[js.OrToken])
				right := groupExpr(X.Value, binaryRightPrecMap[js.OrToken])
				return &js.ExprStmt{Value: &js.BinaryExpr{Op: js.OrToken, X: left, Y: right}}
			}
		} else if hasIf && hasElse {
			ifStmt.Body = m.optimizeStmt(ifStmt.Body)
			ifStmt.Else = m.optimizeStmt(ifStmt.Else)
			XExpr, isExprBody := ifStmt.Body.(*js.ExprStmt)
			YExpr, isExprElse := ifStmt.Else.(*js.ExprStmt)
			if isExprBody && isExprElse {
				return &js.ExprStmt{Value: condExpr(ifStmt.Cond, XExpr.Value, YExpr.Value)}
			}
			XReturn, isReturnBody := ifStmt.Body.(*js.ReturnStmt)
			YReturn, isReturnElse := ifStmt.Else.(*js.ReturnStmt)
			if isReturnBody && isReturnElse {
				if XReturn.Value == nil && YReturn.Value == nil {
					return &js.ReturnStmt{Value: commaExpr(ifStmt.Cond, &js.UnaryExpr{
						Op: js.VoidToken,
						X:  &js.LiteralExpr{TokenType: js.NumericToken, Data: zeroBytes},
					})}
				} else if XReturn.Value != nil && YReturn.Value != nil {
					return &js.ReturnStmt{Value: condExpr(ifStmt.Cond, XReturn.Value, YReturn.Value)}
				}
				return ifStmt
			}
			XThrow, isThrowBody := ifStmt.Body.(*js.ThrowStmt)
			YThrow, isThrowElse := ifStmt.Else.(*js.ThrowStmt)
			if isThrowBody && isThrowElse {
				return &js.ThrowStmt{Value: condExpr(ifStmt.Cond, XThrow.Value, YThrow.Value)}
			}
		}
	} else if decl, ok := i.(*js.VarDecl); ok {
		if decl.TokenType == js.ErrorToken {
			// convert hoisted var declaration to expression or empty (if there are no defines) statement
			for _, item := range decl.List {
				if item.Default != nil {
					return &js.ExprStmt{Value: decl}
				}
			}
			return &js.EmptyStmt{}
		}
		// TODO: remove unused declarations
		//for i := 0; i < len(decl.List); i++ {
		//	if v, ok := decl.List[i].Binding.(*js.Var); ok && v.Uses < 2 {
		//		decl.List = append(decl.List[:i], decl.List[i+1:]...)
		//		i--
		//	}
		//}
		//if len(decl.List) == 0 {
		//	return &js.EmptyStmt{}
		//}
		return decl
	} else if blockStmt, ok := i.(*js.BlockStmt); ok {
		// merge body and remove braces if it is not a lexical declaration
		blockStmt.List = m.optimizeStmtList(blockStmt.List, defaultBlock)
		if len(blockStmt.List) == 1 {
			varDecl, isVarDecl := blockStmt.List[0].(*js.VarDecl)
			_, isClassDecl := blockStmt.List[0].(*js.ClassDecl)
			if !isClassDecl && (!isVarDecl || varDecl.TokenType == js.VarToken) {
				return m.optimizeStmt(blockStmt.List[0])
			}
			return &js.EmptyStmt{}
		} else if len(blockStmt.List) == 0 {
			return &js.EmptyStmt{}
		}
		return blockStmt
	}
	return i
}

func (m *jsMinifier) optimizeStmtList(list []js.IStmt, blockType blockType) []js.IStmt {
	// merge expression statements as well as if/else statements followed by flow control statements
	if len(list) == 0 {
		return list
	}
	j := 0                           // write index
	for i := 0; i < len(list); i++ { // read index
		if ifStmt, ok := list[i].(*js.IfStmt); ok && !m.isEmptyStmt(ifStmt.Else) && isFlowStmt(lastStmt(ifStmt.Body)) {
			// if body ends in flow statement (return, throw, break, continue), so we can remove the else statement and put its body in the current scope
			if blockStmt, ok := ifStmt.Else.(*js.BlockStmt); ok {
				blockStmt.Scope.Unscope()
				list = append(list[:i+1], append(blockStmt.List, list[i+1:]...)...)
			} else {
				list = append(list[:i+1], append([]js.IStmt{ifStmt.Else}, list[i+1:]...)...)
			}
			ifStmt.Else = nil
		}

		list[i] = m.optimizeStmt(list[i])

		if _, ok := list[i].(*js.EmptyStmt); ok {
			k := i + 1
			for ; k < len(list); k++ {
				if _, ok := list[k].(*js.EmptyStmt); !ok {
					break
				}
			}
			list = append(list[:i], list[k:]...)
			i--
			continue
		}

		if 0 < i {
			// merge expression statements with expression, return, and throw statements
			if left, ok := list[i-1].(*js.ExprStmt); ok {
				if right, ok := list[i].(*js.ExprStmt); ok {
					right.Value = commaExpr(left.Value, right.Value)
					j--
				} else if returnStmt, ok := list[i].(*js.ReturnStmt); ok && returnStmt.Value != nil {
					returnStmt.Value = commaExpr(left.Value, returnStmt.Value)
					j--
				} else if throwStmt, ok := list[i].(*js.ThrowStmt); ok {
					throwStmt.Value = commaExpr(left.Value, throwStmt.Value)
					j--
				} else if forStmt, ok := list[i].(*js.ForStmt); ok {
					if varDecl, ok := forStmt.Init.(*js.VarDecl); ok && len(varDecl.List) == 0 || forStmt.Init == nil {
						// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
						forStmt.Init = left.Value
						j--
					}
				} else if whileStmt, ok := list[i].(*js.WhileStmt); ok {
					// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
					var body *js.BlockStmt
					if blockStmt, ok := whileStmt.Body.(*js.BlockStmt); ok {
						body = blockStmt
					} else {
						body = &js.BlockStmt{}
						body.List = []js.IStmt{whileStmt.Body}
					}
					list[i] = &js.ForStmt{Init: left.Value, Cond: whileStmt.Cond, Post: nil, Body: body}
					j--
				} else if switchStmt, ok := list[i].(*js.SwitchStmt); ok {
					switchStmt.Init = commaExpr(left.Value, switchStmt.Init)
					j--
				} else if withStmt, ok := list[i].(*js.WithStmt); ok {
					withStmt.Cond = commaExpr(left.Value, withStmt.Cond)
					j--
				} else if ifStmt, ok := list[i].(*js.IfStmt); ok {
					ifStmt.Cond = commaExpr(left.Value, ifStmt.Cond)
					j--
				} else if varDecl, ok := list[i].(*js.VarDecl); ok {
					if merge := mergeVarDeclExprStmt(varDecl, left, true); merge {
						j--
					}
				}
			} else if left, ok := list[i-1].(*js.VarDecl); ok {
				if right, ok := list[i].(*js.VarDecl); ok && left.TokenType == right.TokenType {
					// merge const and let declarations
					right.List = append(left.List, right.List...)
					j--
				} else if left.TokenType == js.VarToken {
					if exprStmt, ok := list[i].(*js.ExprStmt); ok {
						// pull in assignments to variables into the declaration, e.g. var a;a=5  =>  var a=5
						if merge := mergeVarDeclExprStmt(left, exprStmt, false); merge {
							list[i] = list[i-1]
							j--
						}
					} else if forStmt, ok := list[i].(*js.ForStmt); ok {
						// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
						if forStmt.Init == nil {
							forStmt.Init = left
							j--
						} else if decl, ok := forStmt.Init.(*js.VarDecl); ok && decl.TokenType == js.ErrorToken && !hasDefines(decl) {
							forStmt.Init = left
							j--
						} else if ok && (decl.TokenType == js.VarToken || decl.TokenType == js.ErrorToken) {
							// this is the second VarDecl, so we are hoisting var declarations, which means the forInit variables are already in 'left'
							merge := mergeVarDecls(left, decl)
							if merge {
								decl.TokenType = js.VarToken
								forStmt.Init = left
								j--
							}
						}
					} else if whileStmt, ok := list[i].(*js.WhileStmt); ok {
						// TODO: only merge statements that don't have 'in' or 'of' keywords (slow to check?)
						var body *js.BlockStmt
						if blockStmt, ok := whileStmt.Body.(*js.BlockStmt); ok {
							body = blockStmt
						} else {
							body = &js.BlockStmt{}
							body.List = []js.IStmt{whileStmt.Body}
						}
						list[i] = &js.ForStmt{Init: left, Cond: whileStmt.Cond, Post: nil, Body: body}
						j--
					}
				}
			}
		}
		list[j] = list[i]

		// merge if/else with return/throw when followed by return/throw
	MergeIfReturnThrow:
		if 0 < j {
			// separate from expression merging in case of:  if(a)return b;b=c;return d
			if ifStmt, ok := list[j-1].(*js.IfStmt); ok && m.isEmptyStmt(ifStmt.Body) != m.isEmptyStmt(ifStmt.Else) {
				// either the if body is empty or the else body is empty. In case where both bodies have return/throw, we already rewrote that if statement to an return/throw statement
				if returnStmt, ok := list[j].(*js.ReturnStmt); ok {
					if returnStmt.Value == nil {
						if left, ok := ifStmt.Body.(*js.ReturnStmt); ok && left.Value == nil {
							list[j-1] = &js.ExprStmt{Value: ifStmt.Cond}
						} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value == nil {
							list[j-1] = &js.ExprStmt{Value: ifStmt.Cond}
						}
					} else {
						if left, ok := ifStmt.Body.(*js.ReturnStmt); ok && left.Value != nil {
							returnStmt.Value = condExpr(ifStmt.Cond, left.Value, returnStmt.Value)
							list[j-1] = returnStmt
							j--
							goto MergeIfReturnThrow
						} else if left, ok := ifStmt.Else.(*js.ReturnStmt); ok && left.Value != nil {
							returnStmt.Value = condExpr(ifStmt.Cond, returnStmt.Value, left.Value)
							list[j-1] = returnStmt
							j--
							goto MergeIfReturnThrow
						}
					}
				} else if throwStmt, ok := list[j].(*js.ThrowStmt); ok {
					if left, ok := ifStmt.Body.(*js.ThrowStmt); ok {
						throwStmt.Value = condExpr(ifStmt.Cond, left.Value, throwStmt.Value)
						list[j-1] = throwStmt
						j--
						goto MergeIfReturnThrow
					} else if left, ok := ifStmt.Else.(*js.ThrowStmt); ok {
						throwStmt.Value = condExpr(ifStmt.Cond, throwStmt.Value, left.Value)
						list[j-1] = throwStmt
						j--
						goto MergeIfReturnThrow
					}
				}
			}
		}
		j++
	}

	// remove superfluous return or continue
	if 0 < j {
		if blockType == functionBlock {
			if returnStmt, ok := list[j-1].(*js.ReturnStmt); ok {
				if returnStmt.Value == nil || m.isUndefined(returnStmt.Value) {
					j--
				} else if commaExpr, ok := returnStmt.Value.(*js.CommaExpr); ok && m.isUndefined(commaExpr.List[len(commaExpr.List)-1]) {
					// rewrite function f(){return a,void 0} => function f(){a}
					if len(commaExpr.List) == 2 {
						list[j-1] = &js.ExprStmt{Value: commaExpr.List[0]}
					} else {
						commaExpr.List = commaExpr.List[:len(commaExpr.List)-1]
					}
				}
			}
		} else if blockType == iterationBlock {
			if branchStmt, ok := list[j-1].(*js.BranchStmt); ok && branchStmt.Type == js.ContinueToken && branchStmt.Label == nil {
				j--
			}
		}
	}
	return list[:j]
}

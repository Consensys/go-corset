// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package format

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/expr"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/lval"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/stmt"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/variable"
	"github.com/consensys/go-corset/pkg/zkc/compiler/parser"
	zkc_util "github.com/consensys/go-corset/pkg/zkc/util"
)

type commentEntry struct {
	pos  int
	text string
}

type formatter struct {
	w       *bufio.Writer
	indent  int
	cmts    []commentEntry
	nextCmt int
}

// Format formats a parsed source file, writing the canonical result to w.
// Comments are re-inserted before declarations that appear in the source map
// (constants, functions, and type aliases). Comments before memory declarations
// are accumulated and flushed before the next source-mapped declaration.
func Format(w io.Writer, file parser.UnlinkedSourceFile, src source.File) error {
	f := &formatter{
		w:    bufio.NewWriter(w),
		cmts: collectComments(src),
	}
	// Emit include directives.
	for _, inc := range file.Includes {
		f.writeln("include \"" + *inc + "\"")
	}

	if len(file.Includes) > 0 && len(file.Components) > 0 {
		f.writeln("")
	}

	// Emit declarations.
	for i, comp := range file.Components {
		if i != 0 {
			f.writeln("")
		}

		if file.SourceMap.Has(comp) {
			span := file.SourceMap.Get(comp)
			f.flushComments(span.Start())
		}

		f.formatDecl(comp)
	}

	// Flush any trailing comments (and comments before memory declarations).
	f.flushComments(math.MaxInt)

	return f.w.Flush()
}

func collectComments(src source.File) []commentEntry {
	tokens, errs := parser.Lex(src, true)
	if len(errs) > 0 {
		return nil
	}

	var cmts []commentEntry

	contents := src.Contents()

	for _, tok := range tokens {
		if tok.Kind == parser.COMMENT {
			text := string(contents[tok.Span.Start():tok.Span.End()])
			cmts = append(cmts, commentEntry{tok.Span.Start(), text})
		}
	}

	return cmts
}

func (f *formatter) flushComments(upTo int) {
	for f.nextCmt < len(f.cmts) && f.cmts[f.nextCmt].pos < upTo {
		f.writeln(f.cmts[f.nextCmt].text)
		f.nextCmt++
	}
}

//nolint:errcheck
func (f *formatter) write(s string) {
	f.w.WriteString(s)
}

//nolint:errcheck
func (f *formatter) writeln(s string) {
	f.w.WriteString(s)
	f.w.WriteString("\n")
}

//nolint:errcheck
func (f *formatter) writeIndent() {
	f.w.WriteString(strings.Repeat("    ", f.indent))
}

// ============================================================================
// Declaration formatters
// ============================================================================

func (f *formatter) formatDecl(d decl.Unresolved) {
	for _, ann := range d.Annotations() {
		f.writeln("@" + ann)
	}

	switch d := d.(type) {
	case *decl.Constant[symbol.Unresolved]:
		f.formatConstant(d)
	case *decl.Function[symbol.Unresolved]:
		f.formatFunction(d)
	case *decl.Memory[symbol.Unresolved]:
		f.formatMemory(d)
	case *decl.TypeAlias[symbol.Unresolved]:
		f.formatTypeAlias(d)
	default:
		panic(fmt.Sprintf("unknown declaration type %T", d))
	}
}

func (f *formatter) formatConstant(c *decl.Constant[symbol.Unresolved]) {
	var b strings.Builder

	b.WriteString("const ")
	b.WriteString(c.Name())
	b.WriteString(":")
	b.WriteString(formatType(c.DataType))
	b.WriteString(" = ")
	b.WriteString(formatExpr(c.ConstExpr, nil))
	f.writeln(b.String())
}

func (f *formatter) formatFunction(fn *decl.Function[symbol.Unresolved]) {
	var b strings.Builder

	b.WriteString("fn ")
	b.WriteString(fn.Name())
	// Memory effects: <eff1,eff2,...>
	if len(fn.Effects) > 0 {
		b.WriteString("<")

		for i, eff := range fn.Effects {
			if i != 0 {
				b.WriteString(",")
			}

			b.WriteString((*eff).String())
		}

		b.WriteString(">")
	}
	// Parameter list.
	b.WriteString("(")
	b.WriteString(formatVarList(fn.Inputs()))
	b.WriteString(")")
	// Optional return list.
	if len(fn.Outputs()) > 0 {
		b.WriteString(" -> (")
		b.WriteString(formatVarList(fn.Outputs()))
		b.WriteString(")")
	}

	b.WriteString(" {")
	f.writeln(b.String())
	f.indent++

	for _, s := range fn.Code {
		f.formatStmt(s, fn)
	}

	f.indent--
	f.writeln("}")
}

func (f *formatter) formatMemory(m *decl.Memory[symbol.Unresolved]) {
	var b strings.Builder

	// Optional pub modifier.
	switch m.Kind {
	case decl.PUBLIC_READ_ONLY_MEMORY, decl.PUBLIC_WRITE_ONCE_MEMORY, decl.PUBLIC_STATIC_MEMORY:
		b.WriteString("pub ")
	}
	// Memory kind keyword.
	switch m.Kind {
	case decl.PUBLIC_READ_ONLY_MEMORY, decl.PRIVATE_READ_ONLY_MEMORY:
		b.WriteString("input")
	case decl.PUBLIC_WRITE_ONCE_MEMORY, decl.PRIVATE_WRITE_ONCE_MEMORY:
		b.WriteString("output")
	case decl.PUBLIC_STATIC_MEMORY, decl.PRIVATE_STATIC_MEMORY:
		b.WriteString("static")
	case decl.RANDOM_ACCESS_MEMORY:
		b.WriteString("memory")
	}

	b.WriteString(" ")
	b.WriteString(m.Name())
	b.WriteString("(")
	b.WriteString(formatVarList(m.Address))
	b.WriteString(")")

	if len(m.Data) > 0 {
		b.WriteString(" -> (")
		b.WriteString(formatVarList(m.Data))
		b.WriteString(")")
	}

	if m.IsStatic() {
		b.WriteString(" {")
		f.writeln(b.String())
		f.indent++

		for i, e := range m.Contents {
			f.writeIndent()
			f.write(formatExpr(e, nil))

			if i < len(m.Contents)-1 {
				f.writeln(",")
			} else {
				f.writeln("")
			}
		}

		f.indent--
		f.writeln("}")
	} else {
		f.writeln(b.String())
	}
}

func (f *formatter) formatTypeAlias(ta *decl.TypeAlias[symbol.Unresolved]) {
	f.writeln("type " + ta.Name() + " = " + formatType(ta.DataType))
}

// ============================================================================
// Statement formatters
// ============================================================================

func (f *formatter) formatStmt(s stmt.Unresolved, fn *decl.Function[symbol.Unresolved]) {
	switch s := s.(type) {
	case *stmt.VarDecl[symbol.Unresolved]:
		f.writeIndent()
		f.write("var ")

		for i, id := range s.Variables {
			if i != 0 {
				f.write(", ")
			}

			v := fn.Variable(id)
			f.write(v.Name + ":" + formatType(v.DataType))
		}

		if s.Init.HasValue() {
			f.write(" = ")
			f.write(formatExpr(s.Init.Unwrap(), fn))
		}

		f.writeln("")
	case *stmt.Assign[symbol.Unresolved]:
		f.writeIndent()
		f.writeln(formatAssign(s, fn))
	case *stmt.IfElse[symbol.Unresolved]:
		f.formatIfElse(s, fn, false)
	case *stmt.While[symbol.Unresolved]:
		f.writeIndent()
		f.write("while ")
		f.write(formatExpr(s.Cond, fn))
		f.writeln(" {")
		f.indent++

		for _, b := range s.Body {
			f.formatStmt(b, fn)
		}

		f.indent--
		f.writeIndent()
		f.writeln("}")
	case *stmt.For[symbol.Unresolved]:
		f.writeIndent()
		f.write("for ")
		f.write(formatForInit(s.Init, fn))
		f.write("; ")
		f.write(formatExpr(s.Cond, fn))
		f.write("; ")
		f.write(formatStmtInline(s.Post, fn))
		f.writeln(" {")
		f.indent++

		for _, b := range s.Body {
			f.formatStmt(b, fn)
		}

		f.indent--
		f.writeIndent()
		f.writeln("}")
	case *stmt.Return[symbol.Unresolved]:
		f.writeIndent()
		f.writeln("return")
	case *stmt.Fail[symbol.Unresolved]:
		f.writeIndent()
		f.writeln("fail")
	case *stmt.Break[symbol.Unresolved]:
		f.writeIndent()
		f.writeln("break")
	case *stmt.Continue[symbol.Unresolved]:
		f.writeIndent()
		f.writeln("continue")
	case *stmt.Printf[symbol.Unresolved]:
		f.writeIndent()
		f.writeln(formatPrintf(s, fn))
	default:
		panic(fmt.Sprintf("unknown statement type %T", s))
	}
}

func (f *formatter) formatIfElse(
	s *stmt.IfElse[symbol.Unresolved],
	fn *decl.Function[symbol.Unresolved],
	isElseIf bool,
) {
	if !isElseIf {
		f.writeIndent()
	}

	f.write("if ")
	f.write(formatExpr(s.Cond, fn))
	f.writeln(" {")
	f.indent++

	for _, b := range s.TrueBranch {
		f.formatStmt(b, fn)
	}

	f.indent--

	if len(s.FalseBranch) == 0 {
		f.writeIndent()
		f.writeln("}")

		return
	}

	// Check for else-if: a single IfElse in the false branch.
	if len(s.FalseBranch) == 1 {
		if elseIf, ok := s.FalseBranch[0].(*stmt.IfElse[symbol.Unresolved]); ok {
			f.writeIndent()
			f.write("} else ")
			f.formatIfElse(elseIf, fn, true)

			return
		}
	}

	f.writeIndent()
	f.writeln("} else {")
	f.indent++

	for _, b := range s.FalseBranch {
		f.formatStmt(b, fn)
	}

	f.indent--
	f.writeIndent()
	f.writeln("}")
}

// formatForInit formats a for-loop init statement.
// A VarDecl init omits the "var" keyword (e.g. "i:u8 = 0" not "var i:u8 = 0").
func formatForInit(s stmt.Unresolved, fn *decl.Function[symbol.Unresolved]) string {
	switch s := s.(type) {
	case *stmt.VarDecl[symbol.Unresolved]:
		v := fn.Variable(s.Variables[0])
		init := ""

		if s.Init.HasValue() {
			init = " = " + formatExpr(s.Init.Unwrap(), fn)
		}

		return v.Name + ":" + formatType(v.DataType) + init
	case *stmt.Assign[symbol.Unresolved]:
		return formatAssign(s, fn)
	default:
		panic(fmt.Sprintf("unexpected for-init type %T", s))
	}
}

// formatStmtInline formats a statement as an inline string (for for-loop post).
func formatStmtInline(s stmt.Unresolved, fn *decl.Function[symbol.Unresolved]) string {
	switch s := s.(type) {
	case *stmt.Assign[symbol.Unresolved]:
		return formatAssign(s, fn)
	default:
		panic(fmt.Sprintf("unexpected inline statement type %T", s))
	}
}

// formatAssign formats an Assign statement.
// Targets are stored LSB-first (reversed from source); we reverse back.
func formatAssign(s *stmt.Assign[symbol.Unresolved], fn *decl.Function[symbol.Unresolved]) string {
	var b strings.Builder

	if len(s.Targets) > 0 {
		for i := len(s.Targets) - 1; i >= 0; i-- {
			if i != len(s.Targets)-1 {
				b.WriteString(", ")
			}

			b.WriteString(formatLval(s.Targets[i], fn))
		}

		b.WriteString(" = ")
	}

	b.WriteString(formatExpr(s.Source, fn))

	return b.String()
}

func formatPrintf(s *stmt.Printf[symbol.Unresolved], fn *decl.Function[symbol.Unresolved]) string {
	var b strings.Builder

	b.WriteString("printf \"")

	for _, chunk := range s.Chunks {
		b.WriteString(zkc_util.EscapeFormattedText(chunk.Text))

		if chunk.Format.HasFormat() {
			b.WriteString(chunk.Format.String())
		}
	}

	b.WriteString("\"")

	for _, arg := range s.Arguments {
		b.WriteString(", ")
		b.WriteString(formatExpr(arg, fn))
	}

	return b.String()
}

// ============================================================================
// Expression formatter
// ============================================================================

// formatExpr converts an expression to its canonical string form.
// fn may be nil when formatting constant/static-initialiser expressions.
func formatExpr(e expr.Unresolved, fn *decl.Function[symbol.Unresolved]) string {
	switch e := e.(type) {
	case *expr.Cast[symbol.Unresolved]:
		inner := formatExprParens(e.Expr, fn)

		var env data.Environment[symbol.Unresolved]

		return inner + " as " + e.CastType.String(env)
	case *expr.Const[symbol.Unresolved]:
		return e.String(nil)
	case *expr.LocalAccess[symbol.Unresolved]:
		return fn.Variable(e.Variable).Name
	case *expr.ExternAccess[symbol.Unresolved]:
		return formatExternAccess(e, fn)
	case *expr.Add[symbol.Unresolved]:
		return joinExprs(e.Exprs, "+", fn)
	case *expr.Sub[symbol.Unresolved]:
		return joinExprs(e.Exprs, "-", fn)
	case *expr.Mul[symbol.Unresolved]:
		return joinExprs(e.Exprs, "*", fn)
	case *expr.Div[symbol.Unresolved]:
		return joinExprs(e.Exprs, "/", fn)
	case *expr.Rem[symbol.Unresolved]:
		return joinExprs(e.Exprs, "%", fn)
	case *expr.BitwiseAnd[symbol.Unresolved]:
		return joinExprs(e.Exprs, "&", fn)
	case *expr.BitwiseOr[symbol.Unresolved]:
		return joinExprs(e.Exprs, "|", fn)
	case *expr.Xor[symbol.Unresolved]:
		return joinExprs(e.Exprs, "^", fn)
	case *expr.Shl[symbol.Unresolved]:
		return joinExprs(e.Exprs, "<<", fn)
	case *expr.Shr[symbol.Unresolved]:
		return joinExprs(e.Exprs, ">>", fn)
	case *expr.Concat[symbol.Unresolved]:
		return joinExprs(e.Exprs, "::", fn)
	case *expr.BitwiseNot[symbol.Unresolved]:
		return "~" + formatExprParens(e.Expr, fn)
	case *expr.LogicalNot[symbol.Unresolved]:
		return "!" + formatExprParens(e.Expr, fn)
	case *expr.LogicalAnd[symbol.Unresolved]:
		return joinExprs(e.Exprs, "&&", fn)
	case *expr.LogicalOr[symbol.Unresolved]:
		return joinExprs(e.Exprs, "||", fn)
	case *expr.Cmp[symbol.Unresolved]:
		l := formatExprParens(e.Left, fn)
		r := formatExprParens(e.Right, fn)

		return l + " " + cmpOpStr(e.Operator) + " " + r
	case *expr.Ternary[symbol.Unresolved]:
		return formatExpr(e.Cond, fn) + " ? " + formatExpr(e.IfTrue, fn) + " : " + formatExpr(e.IfFalse, fn)
	default:
		panic(fmt.Sprintf("unknown expression type %T", e))
	}
}

// formatExprParens wraps the expression in parens if it is a compound expression.
func formatExprParens(e expr.Unresolved, fn *decl.Function[symbol.Unresolved]) string {
	if exprNeedsBraces(e) {
		return "(" + formatExpr(e, fn) + ")"
	}

	return formatExpr(e, fn)
}

func formatExternAccess(e *expr.ExternAccess[symbol.Unresolved], fn *decl.Function[symbol.Unresolved]) string {
	name := e.Name.String()

	if e.Name.IsFunction() {
		var b strings.Builder

		b.WriteString(name)
		b.WriteString("(")

		for i, arg := range e.Args {
			if i != 0 {
				b.WriteString(", ")
			}

			b.WriteString(formatExpr(arg, fn))
		}

		b.WriteString(")")

		return b.String()
	}

	if e.Name.IsMemory() {
		var b strings.Builder

		b.WriteString(name)
		b.WriteString("[")

		for i, arg := range e.Args {
			if i != 0 {
				b.WriteString(", ")
			}

			b.WriteString(formatExpr(arg, fn))
		}

		b.WriteString("]")

		return b.String()
	}

	return name
}

func joinExprs(exprs []expr.Unresolved, op string, fn *decl.Function[symbol.Unresolved]) string {
	var b strings.Builder

	for i, e := range exprs {
		if i != 0 {
			b.WriteString(" ")
			b.WriteString(op)
			b.WriteString(" ")
		}

		if exprNeedsBraces(e) {
			b.WriteString("(")
			b.WriteString(formatExpr(e, fn))
			b.WriteString(")")
		} else {
			b.WriteString(formatExpr(e, fn))
		}
	}

	return b.String()
}

func exprNeedsBraces(e expr.Unresolved) bool {
	switch e.(type) {
	case *expr.Cast[symbol.Unresolved]:
		return false
	case *expr.Const[symbol.Unresolved]:
		return false
	case *expr.LocalAccess[symbol.Unresolved]:
		return false
	case *expr.ExternAccess[symbol.Unresolved]:
		return false
	default:
		return true
	}
}

func cmpOpStr(op expr.CmpOp) string {
	switch op {
	case expr.EQ:
		return "=="
	case expr.NEQ:
		return "!="
	case expr.LT:
		return "<"
	case expr.LTEQ:
		return "<="
	case expr.GT:
		return ">"
	case expr.GTEQ:
		return ">="
	default:
		panic("unknown cmp op")
	}
}

// ============================================================================
// LVal formatter
// ============================================================================

func formatLval(lv lval.Unresolved, fn *decl.Function[symbol.Unresolved]) string {
	switch lv := lv.(type) {
	case *lval.Variable[symbol.Unresolved]:
		var b strings.Builder

		for i, id := range lv.Ids {
			if i != 0 {
				b.WriteString("::")
			}

			b.WriteString(fn.Variable(id).Name)
		}

		return b.String()
	case *lval.MemAccess[symbol.Unresolved]:
		var b strings.Builder

		b.WriteString(lv.Name.String())
		b.WriteString("[")

		for i, arg := range lv.Args {
			if i != 0 {
				b.WriteString(", ")
			}

			b.WriteString(formatExpr(arg, fn))
		}

		b.WriteString("]")

		return b.String()
	default:
		panic(fmt.Sprintf("unknown lval type %T", lv))
	}
}

// ============================================================================
// Type / variable list helpers
// ============================================================================

func formatType(t data.Type[symbol.Unresolved]) string {
	var env data.Environment[symbol.Unresolved]
	return t.String(env)
}

func formatVarList(vars []variable.Descriptor[symbol.Unresolved]) string {
	var b strings.Builder

	for i, v := range vars {
		if i != 0 {
			b.WriteString(", ")
		}

		b.WriteString(v.Name)
		b.WriteString(":")
		b.WriteString(formatType(v.DataType))
	}

	return b.String()
}

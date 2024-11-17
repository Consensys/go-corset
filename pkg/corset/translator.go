package corset

import (
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// TranslateCircuit translates the components of a Corset circuit and add them
// to the schema.  By the time we get to this point, all malformed source files
// should have been rejected already and the translation should go through
// easily.  Thus, whilst syntax errors can be returned here, this should never
// happen.  The mechanism is supported, however, to simplify development of new
// features, etc.
func TranslateCircuit(env *Environment, srcmap *sexp.SourceMaps[Node], circuit *Circuit) (*hir.Schema, []SyntaxError) {
	t := translator{env, srcmap, hir.EmptySchema()}
	// Allocate all modules into schema
	t.translateModules(circuit)
	// Translate root declarations
	errors := t.translateDeclarations("", circuit.Declarations)
	// Translate nested declarations
	for _, m := range circuit.Modules {
		errs := t.translateDeclarations(m.Name, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return t.schema, errors
}

// Translator packages up information necessary for translating a circuit into
// the schema form required for the HIR level.
type translator struct {
	// Environment determines module and column indices, as needed for
	// translating the various constructs found in a circuit.
	env *Environment
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *sexp.SourceMaps[Node]
	// Represents the schema being constructed by this translator.
	schema *hir.Schema
}

func (t *translator) translateModules(circuit *Circuit) {
	// Add root module
	t.schema.AddModule("")
	// Add nested modules
	for _, m := range circuit.Modules {
		mid := t.schema.AddModule(m.Name)
		aid := t.env.Module(m.Name)
		// Sanity check everything lines up.
		if aid != mid {
			panic(fmt.Sprintf("Invalid module identifier: %d vs %d", mid, aid))
		}
	}
}

// Translate all Corset declarations in a given module, adding them to the
// schema.  By the time we get to this point, all malformed source files should
// have been rejected already and the translation should go through easily.
// Thus, whilst syntax errors can be returned here, this should never happen.
// The mechanism is supported, however, to simplify development of new features,
// etc.
func (t *translator) translateDeclarations(module string, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	// Construct context for enclosing module
	context := t.env.Module(module)
	//
	for _, d := range decls {
		errs := t.translateDeclaration(d, context)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate a Corset declaration and add it to the schema.  By the time we get
// to this point, all malformed source files should have been rejected already
// and the translation should go through easily.  Thus, whilst syntax errors can
// be returned here, this should never happen.  The mechanism is supported,
// however, to simplify development of new features, etc.
func (t *translator) translateDeclaration(decl Declaration, module uint) []SyntaxError {
	var errors []SyntaxError
	//
	if d, ok := decl.(*DefColumns); ok {
		t.translateDefColumns(d, module)
	} else if d, ok := decl.(*DefConstraint); ok {
		errors = t.translateDefConstraint(d, module)
	} else if d, ok := decl.(*DefInRange); ok {
		errors = t.translateDefInRange(d, module)
	} else if d, ok := decl.(*DefProperty); ok {
		errors = t.translateDefProperty(d, module)
	} else {
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Translate a "defcolumns" declaration.
func (t *translator) translateDefColumns(decl *DefColumns, module uint) {
	// Add each column to schema
	for _, c := range decl.Columns {
		// FIXME: support user-defined length multiplier
		context := tr.NewContext(module, 1)
		cid := t.schema.AddDataColumn(context, c.Name, c.DataType)
		// Sanity check column identifier
		if id := t.env.Column(module, c.Name); id != cid {
			panic(fmt.Sprintf("invalid column identifier: %d vs %d", cid, id))
		}
	}
}

// Translate a "defconstraint" declaration.
func (t *translator) translateDefConstraint(decl *DefConstraint, module uint) []SyntaxError {
	// Translate constraint body
	constraint, errors := t.translateExpressionInModule(decl.Constraint, module)
	// Translate (optional) guard
	guard, guard_errors := t.translateOptionalExpressionInModule(decl.Guard, module)
	// Combine errors
	errors = append(errors, guard_errors...)
	// Apply guard
	if guard != nil {
		constraint = &hir.Mul{Args: []hir.Expr{guard, constraint}}
	}
	//
	if len(errors) == 0 {
		context := tr.NewContext(module, 1)
		// Add translated constraint
		t.schema.AddVanishingConstraint(decl.Handle, context, decl.Domain, constraint)
	}
	// Done
	return errors
}

// Translate a "definrange" declaration.
func (t *translator) translateDefInRange(decl *DefInRange, module uint) []SyntaxError {
	// Translate constraint body
	expr, errors := t.translateExpressionInModule(decl.Expr, module)
	//
	if len(errors) == 0 {
		context := tr.NewContext(module, 1)
		// Add translated constraint
		t.schema.AddRangeConstraint("", context, expr, decl.Bound)
	}
	// Done
	return errors
}

// Translate a "defproperty" declaration.
func (t *translator) translateDefProperty(decl *DefProperty, module uint) []SyntaxError {
	// Translate constraint body
	assertion, errors := t.translateExpressionInModule(decl.Assertion, module)
	//
	if len(errors) == 0 {
		context := tr.NewContext(module, 1)
		// Add translated constraint
		t.schema.AddPropertyAssertion(decl.Handle, context, assertion)
	}
	// Done
	return errors
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateOptionalExpressionInModule(expr Expr, module uint) (hir.Expr, []SyntaxError) {
	if expr != nil {
		return t.translateExpressionInModule(expr, module)
	}

	return nil, nil
}

// Translate a sequence of zero or more expressions enclosed in a given module.
func (t *translator) translateExpressionsInModule(exprs []Expr, module uint) ([]hir.Expr, []SyntaxError) {
	errors := []SyntaxError{}
	hirExprs := make([]hir.Expr, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			hirExprs[i], errs = t.translateExpressionInModule(e, module)
			errors = append(errors, errs...)
		}
	}
	// Done
	return hirExprs, errors
}

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for column access, function
// invocations, etc).
func (t *translator) translateExpressionInModule(expr Expr, module uint) (hir.Expr, []SyntaxError) {
	if e, ok := expr.(*Constant); ok {
		return &hir.Constant{Val: e.Val}, nil
	} else if v, ok := expr.(*Add); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module)
		return &hir.Add{Args: args}, errs
	} else if v, ok := expr.(*Exp); ok {
		arg, errs := t.translateExpressionInModule(v.Arg, module)
		return &hir.Exp{Arg: arg, Pow: v.Pow}, errs
	} else if v, ok := expr.(*IfZero); ok {
		args, errs := t.translateExpressionsInModule([]Expr{v.Condition, v.TrueBranch, v.FalseBranch}, module)
		return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}, errs
	} else if v, ok := expr.(*List); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module)
		return &hir.List{Args: args}, errs
	} else if v, ok := expr.(*Mul); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module)
		return &hir.Mul{Args: args}, errs
	} else if v, ok := expr.(*Normalise); ok {
		arg, errs := t.translateExpressionInModule(v.Arg, module)
		return &hir.Normalise{Arg: arg}, errs
	} else if v, ok := expr.(*Sub); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module)
		return &hir.Sub{Args: args}, errs
	} else if e, ok := expr.(*VariableAccess); ok {
		cid := t.env.Column(module, e.Name)
		return &hir.ColumnAccess{Column: cid, Shift: e.Shift}, nil
	} else {
		return nil, t.srcmap.SyntaxErrors(expr, "unknown expression")
	}
}

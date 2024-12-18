package corset

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
)

// TranslateCircuit translates the components of a Corset circuit and add them
// to the schema.  By the time we get to this point, all malformed source files
// should have been rejected already and the translation should go through
// easily.  Thus, whilst syntax errors can be returned here, this should never
// happen.  The mechanism is supported, however, to simplify development of new
// features, etc.
func TranslateCircuit(env Environment, debug bool, srcmap *sexp.SourceMaps[Node],
	circuit *Circuit) (*hir.Schema, []SyntaxError) {
	//
	t := translator{env, debug, srcmap, hir.EmptySchema()}
	// Allocate all modules into schema
	t.translateModules(circuit)
	// Translate input columns
	if errs := t.translateInputColumns(circuit); len(errs) > 0 {
		return nil, errs
	}
	// Translate everything else
	if errs := t.translateOtherDeclarations(circuit); len(errs) > 0 {
		return nil, errs
	}
	// Done
	return t.schema, nil
}

// Translator packages up information necessary for translating a circuit into
// the schema form required for the HIR level.
type translator struct {
	// Environment is needed for determining the identifiers for modules and
	// columns.
	env Environment
	// Debug enables the use of debug constraints.
	debug bool
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
		aid := t.env.Module(m.Name).mid
		// Sanity check everything lines up.
		if aid != mid {
			panic(fmt.Sprintf("Invalid module identifier: %d vs %d", mid, aid))
		}
	}
}

// Translate all input column declarations in the entire circuit.
func (t *translator) translateInputColumns(circuit *Circuit) []SyntaxError {
	errors := t.translateInputColumnsInModule("", circuit.Declarations)
	// Translate each module
	for _, m := range circuit.Modules {
		errs := t.translateInputColumnsInModule(m.Name, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate all input column declarations occurring in a given module within the circuit.
func (t *translator) translateInputColumnsInModule(module string, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		if dcols, ok := d.(*DefColumns); ok {
			errs := t.translateDefColumns(dcols, module)
			errors = append(errors, errs...)
		}
	}
	// Done
	return errors
}

// Translate a "defcolumns" declaration.
func (t *translator) translateDefColumns(decl *DefColumns, module string) []SyntaxError {
	var errors []SyntaxError
	// Add each column to schema
	for _, c := range decl.Columns {
		errs := t.translateDefColumn(c, module)
		errors = append(errors, errs...)
	}
	//
	return errors
}

// Translate a "defcolumns" declaration.
func (t *translator) translateDefColumn(decl *DefColumn, module string) []SyntaxError {
	// Extract base columnd ID
	columnId := t.env.Column(module, decl.Name()).cid
	//
	if arr_t, ok := decl.DataType().(*ArrayType); ok {
		var errors []SyntaxError
		// Handle array types
		for i := uint(1); i <= arr_t.size; i++ {
			name := fmt.Sprintf("%s_%d", decl.name, i)
			errs := t.translateRawColumn(decl, module, name, arr_t.element.AsUnderlying(), columnId)
			errors = append(errors, errs...)
			columnId++
		}
		//
		return errors
	} else {
		return t.translateRawColumn(decl, module, decl.name, decl.DataType().AsUnderlying(), columnId)
	}
}

func (t *translator) translateRawColumn(decl *DefColumn, module string, name string,
	datatype sc.Type, columnId uint) []SyntaxError {
	//
	context := t.env.ContextFrom(module, decl.LengthMultiplier())
	cid := t.schema.AddDataColumn(context, name, datatype)
	// Prove type (if requested)
	if decl.MustProve() {
		bound := datatype.AsUint().Bound()
		t.schema.AddRangeConstraint(name, context, &hir.ColumnAccess{Column: cid, Shift: 0}, bound)
	}
	// Sanity check column identifier
	if columnId != cid {
		msg := fmt.Sprintf("invalid column identifier (expected %d got %d)", columnId, cid)
		err := t.srcmap.SyntaxError(decl, msg)
		//
		return []SyntaxError{*err}
	}
	//
	return nil
}

// Translate all assignment or constraint declarations in the circuit.
func (t *translator) translateOtherDeclarations(circuit *Circuit) []SyntaxError {
	errors := t.translateOtherDeclarationsInModule("", circuit.Declarations)
	// Translate each module
	for _, m := range circuit.Modules {
		errs := t.translateOtherDeclarationsInModule(m.Name, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate all assignment or constraint declarations in a given module within
// the circuit.
func (t *translator) translateOtherDeclarationsInModule(module string, decls []Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		errs := t.translateDeclaration(d, module)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate an assignment or constraint declarartion which occurs within a
// given module.
func (t *translator) translateDeclaration(decl Declaration, module string) []SyntaxError {
	var errors []SyntaxError
	//
	if _, ok := decl.(*DefAliases); ok {
		// Not an assignment or a constraint, hence ignore.
	} else if _, ok := decl.(*DefColumns); ok {
		// Not an assignment or a constraint, hence ignore.
	} else if _, ok := decl.(*DefConst); ok {
		// For now, constants are always compiled out when going down to HIR.
	} else if d, ok := decl.(*DefConstraint); ok {
		errors = t.translateDefConstraint(d, module)
	} else if _, ok := decl.(*DefFun); ok {
		// For now, functions are always compiled out when going down to HIR.
		// In the future, this might change if we add support for macros to HIR.
	} else if d, ok := decl.(*DefInRange); ok {
		errors = t.translateDefInRange(d, module)
	} else if d, Ok := decl.(*DefInterleaved); Ok {
		errors = t.translateDefInterleaved(d, module)
	} else if d, ok := decl.(*DefLookup); ok {
		errors = t.translateDefLookup(d, module)
	} else if d, Ok := decl.(*DefPermutation); Ok {
		errors = t.translateDefPermutation(d, module)
	} else if d, ok := decl.(*DefProperty); ok {
		errors = t.translateDefProperty(d, module)
	} else {
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Translate a "defconstraint" declaration.
func (t *translator) translateDefConstraint(decl *DefConstraint, module string) []SyntaxError {
	// Translate constraint body
	constraint, errors := t.translateExpressionInModule(decl.Constraint, module, 0)
	// Translate (optional) guard
	guard, guard_errors := t.translateOptionalExpressionInModule(decl.Guard, module, 0)
	// Combine errors
	errors = append(errors, guard_errors...)
	// Apply guard
	if constraint == nil {
		// NOTE: in this case, the constraint itself has been translated as nil.
		// This means there is no constraint (e.g. its a debug constraint, but
		// debug mode is not enabled).
		return errors
	} else if guard != nil {
		constraint = &hir.Mul{Args: []hir.Expr{guard, constraint}}
	}
	//
	if len(errors) == 0 {
		context := constraint.Context(t.schema)
		//
		if context.IsVoid() {
			// Constraint is a constant (for some reason).
			if constraint.Multiplicity() != 0 {
				return t.srcmap.SyntaxErrors(decl, "constraint is a constant")
			}
		} else if context.Module() != t.env.Module(module).mid {
			return t.srcmap.SyntaxErrors(decl, "invalid context inferred")
		} else {
			// Add translated constraint
			t.schema.AddVanishingConstraint(decl.Handle, context, decl.Domain, constraint)
		}
	}
	// Done
	return errors
}

// Translate a "deflookup" declaration.
//
//nolint:staticcheck
func (t *translator) translateDefLookup(decl *DefLookup, module string) []SyntaxError {
	// Translate source expressions
	sources, src_errs := t.translateUnitExpressionsInModule(decl.Sources, module, 0)
	targets, tgt_errs := t.translateUnitExpressionsInModule(decl.Targets, module, 0)
	// Combine errors
	errors := append(src_errs, tgt_errs...)
	//
	if len(errors) == 0 {
		src_context := t.env.ToContext(ContextOfExpressions(decl.Sources))
		target_context := t.env.ToContext(ContextOfExpressions(decl.Targets))
		// Add translated constraint
		t.schema.AddLookupConstraint(decl.Handle, src_context, target_context, sources, targets)
	}
	// Done
	return errors
}

// Translate a "definrange" declaration.
func (t *translator) translateDefInRange(decl *DefInRange, module string) []SyntaxError {
	// Translate constraint body
	expr, errors := t.translateExpressionInModule(decl.Expr, module, 0)
	//
	if len(errors) == 0 {
		context := t.env.ContextFrom(module, 1)
		// Add translated constraint
		t.schema.AddRangeConstraint("", context, expr, decl.Bound)
	}
	// Done
	return errors
}

// Translate a "definterleaved" declaration.
func (t *translator) translateDefInterleaved(decl *DefInterleaved, module string) []SyntaxError {
	var errors []SyntaxError
	//
	sources := make([]uint, len(decl.Sources))
	// Lookup target column info
	info := t.env.Column(module, decl.Target.Name())
	// Determine source column identifiers
	for i, source := range decl.Sources {
		sources[i] = t.env.Column(module, source.Name()).ColumnId()
	}
	// Construct context for this assignment
	context := t.env.ContextFrom(module, info.multiplier)
	// Extract underlying datatype
	datatype := info.dataType.AsUnderlying()
	// Register assignment
	cid := t.schema.AddAssignment(assignment.NewInterleaving(context, decl.Target.Name(), sources, datatype))
	// Sanity check column identifiers align.
	if cid != info.ColumnId() {
		errors = append(errors, *t.srcmap.SyntaxError(decl, "invalid column identifier"))
	}
	// Done
	return errors
}

// Translate a "defpermutation" declaration.
func (t *translator) translateDefPermutation(decl *DefPermutation, module string) []SyntaxError {
	var (
		errors   []SyntaxError
		context  tr.Context
		firstCid uint
	)
	//
	targets := make([]sc.Column, len(decl.Sources))
	signs := make([]bool, len(decl.Sources))
	sources := make([]uint, len(decl.Sources))
	//
	for i := 0; i < len(decl.Sources); i++ {
		target := t.env.Column(module, decl.Targets[i].Name())
		context = t.env.ContextFrom(module, target.multiplier)
		// Extract underlying datatype
		datatype := target.dataType.AsUnderlying()
		// Construct columns
		targets[i] = sc.NewColumn(context, decl.Targets[i].Name(), datatype)
		sources[i] = t.env.Column(module, decl.Sources[i].Name()).ColumnId()
		signs[i] = decl.Signs[i]
		// Record first CID
		if i == 0 {
			firstCid = target.ColumnId()
		}
	}
	// Add the assignment and check the first identifier.
	cid := t.schema.AddAssignment(assignment.NewSortedPermutation(context, targets, signs, sources))
	// Sanity check column identifiers align.
	if cid != firstCid {
		errors = append(errors, *t.srcmap.SyntaxError(decl, "invalid column identifier"))
	}
	// Done
	return errors
}

// Translate a "defproperty" declaration.
func (t *translator) translateDefProperty(decl *DefProperty, module string) []SyntaxError {
	// Translate constraint body
	assertion, errors := t.translateExpressionInModule(decl.Assertion, module, 0)
	//
	if len(errors) == 0 {
		context := t.env.ContextFrom(module, 1)
		// Add translated constraint
		t.schema.AddPropertyAssertion(decl.Handle, context, assertion)
	}
	// Done
	return errors
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateOptionalExpressionInModule(expr Expr, module string,
	shift int) (hir.Expr, []SyntaxError) {
	//
	if expr != nil {
		return t.translateExpressionInModule(expr, module, shift)
	}

	return nil, nil
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateUnitExpressionsInModule(exprs []Expr, module string,
	shift int) ([]hir.UnitExpr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]hir.UnitExpr, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			expr, errs := t.translateExpressionInModule(e, module, shift)
			errors = append(errors, errs...)
			hirExprs[i] = hir.NewUnitExpr(expr)
		}
	}
	// Done
	return hirExprs, errors
}

// Translate a sequence of zero or more expressions enclosed in a given module.
// All expressions are expected to be non-voidable (see below for more on
// voidability).
func (t *translator) translateExpressionsInModule(exprs []Expr, module string,
	shift int) ([]hir.Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]hir.Expr, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			hirExprs[i], errs = t.translateExpressionInModule(e, module, shift)
			errors = append(errors, errs...)
			// Check for non-voidability
			if hirExprs[i] == nil {
				errors = append(errors, *t.srcmap.SyntaxError(e, "void expression not permitted here"))
			}
		}
	}
	//
	return hirExprs, errors
}

// Translate a sequence of zero or more expressions enclosed in a given module.
// A key aspect of this function is that it additionally accounts for "voidable"
// expressions.  That is, essentially, to account for debug constraints which
// only exist in debug mode.  Hence, when debug mode is not enabled, then a
// debug constraint is "void".
func (t *translator) translateVoidableExpressionsInModule(exprs []Expr, module string,
	shift int) ([]hir.Expr, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]hir.Expr, len(exprs))
	nils := 0
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			hirExprs[i], errs = t.translateExpressionInModule(e, module, shift)
			errors = append(errors, errs...)
			// Update dirty flag
			if hirExprs[i] == nil {
				nils++
			}
		}
	}
	// Nil check.
	if nils == 0 {
		// Done
		return hirExprs, errors
	}
	// Stip nils. Recall that nils can arise legitimately when we have debug
	// constraints, but debug mode is not enabled.  In such case, we want to
	// strip them out.  Since this is a rare occurrence, we try to keep the happy
	// path efficient.
	nHirExprs := make([]hir.Expr, len(exprs)-nils)
	i := 0
	// Strip out nils
	for _, e := range hirExprs {
		if e != nil {
			nHirExprs[i] = e
			i++
		}
	}
	//
	return nHirExprs, errors
}

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for column access, function
// invocations, etc).
func (t *translator) translateExpressionInModule(expr Expr, module string, shift int) (hir.Expr, []SyntaxError) {
	if e, ok := expr.(*ArrayAccess); ok {
		return t.translateArrayAccessInModule(e, shift)
	} else if v, ok := expr.(*Add); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module, shift)
		return &hir.Add{Args: args}, errs
	} else if e, ok := expr.(*Constant); ok {
		var val fr.Element
		// Initialise field from bigint
		val.SetBigInt(&e.Val)
		//
		return &hir.Constant{Val: val}, nil
	} else if e, ok := expr.(*Debug); ok {
		if t.debug {
			return t.translateExpressionInModule(e.Arg, module, shift)
		}
		// When debug is not enabled, simply substitute for 0.
		return nil, nil
	} else if e, ok := expr.(*Exp); ok {
		return t.translateExpInModule(e, module, shift)
	} else if e, ok := expr.(*For); ok {
		return t.translateForInModule(e, module, shift)
	} else if v, ok := expr.(*If); ok {
		args, errs := t.translateExpressionsInModule([]Expr{v.Condition, v.TrueBranch, v.FalseBranch}, module, shift)
		// Construct appropriate if form
		if v.IsIfZero() {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}, errs
		} else if v.IsIfNotZero() {
			// In this case, switch the ordering.
			return &hir.IfZero{Condition: args[0], TrueBranch: args[2], FalseBranch: args[1]}, errs
		}
		// Should be unreachable
		return nil, t.srcmap.SyntaxErrors(expr, "unresolved conditional encountered during translation")
	} else if e, ok := expr.(*Invoke); ok {
		return t.translateInvokeInModule(e, module, shift)
	} else if v, ok := expr.(*List); ok {
		args, errs := t.translateVoidableExpressionsInModule(v.Args, module, shift)
		return &hir.List{Args: args}, errs
	} else if v, ok := expr.(*Mul); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module, shift)
		return &hir.Mul{Args: args}, errs
	} else if v, ok := expr.(*Normalise); ok {
		arg, errs := t.translateExpressionInModule(v.Arg, module, shift)
		return &hir.Normalise{Arg: arg}, errs
	} else if v, ok := expr.(*Reduce); ok {
		return t.translateReduceInModule(v, module, shift)
	} else if v, ok := expr.(*Sub); ok {
		args, errs := t.translateExpressionsInModule(v.Args, module, shift)
		return &hir.Sub{Args: args}, errs
	} else if e, ok := expr.(*Shift); ok {
		return t.translateShiftInModule(e, module, shift)
	} else if e, ok := expr.(*VariableAccess); ok {
		return t.translateVariableAccessInModule(e, module, shift)
	} else {
		return nil, t.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
}

func (t *translator) translateArrayAccessInModule(expr *ArrayAccess, shift int) (hir.Expr, []SyntaxError) {
	var errors []SyntaxError
	// Lookup the column
	binding, ok := expr.Binding().(*ColumnBinding)
	// Did we find it?
	if !ok {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "invalid array index encountered during translation"))
	}
	// Array index should be statically known
	index := expr.arg.AsConstant()
	//
	if index == nil {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "expected constant array index"))
	} else if i := uint(index.Uint64()); i == 0 || (binding != nil && i > binding.dataType.Width()) {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "array index out-of-bounds"))
	}
	// Error check
	if len(errors) > 0 {
		return nil, errors
	}
	// Lookup underlying column info
	info := t.env.Column(binding.module, expr.Name())
	// Update column id (remember indices start from 1)
	columnId := info.cid + uint(index.Uint64()) - 1
	// Done
	return &hir.ColumnAccess{Column: columnId, Shift: shift}, nil
}

func (t *translator) translateExpInModule(expr *Exp, module string, shift int) (hir.Expr, []SyntaxError) {
	arg, errs := t.translateExpressionInModule(expr.Arg, module, shift)
	pow := expr.Pow.AsConstant()
	// Identity constant for pow
	if pow == nil {
		errs = append(errs, *t.srcmap.SyntaxError(expr.Pow, "expected constant power"))
	} else if !pow.IsUint64() {
		errs = append(errs, *t.srcmap.SyntaxError(expr.Pow, "constant power too large"))
	}
	// Sanity check errors
	if len(errs) == 0 {
		return &hir.Exp{Arg: arg, Pow: pow.Uint64()}, errs
	}
	//
	return nil, errs
}

func (t *translator) translateForInModule(expr *For, module string, shift int) (hir.Expr, []SyntaxError) {
	var (
		errors  []SyntaxError
		mapping map[uint]Expr = make(map[uint]Expr)
	)
	// Determine range for index variable
	n := expr.End - expr.Start + 1
	args := make([]hir.Expr, n)
	// Expand body n times
	for i := uint(0); i < n; i++ {
		var errs []SyntaxError
		// Substitute through for i
		mapping[expr.Binding.index] = &Constant{*big.NewInt(int64(i + expr.Start))}
		ith := expr.Body.Substitute(mapping)
		// Translate subsituted expression
		args[i], errs = t.translateExpressionInModule(ith, module, shift)
		errors = append(errors, errs...)
	}
	// Error check
	if len(errors) != 0 {
		return nil, errors
	}
	// Done
	return &hir.List{Args: args}, nil
}

func (t *translator) translateInvokeInModule(expr *Invoke, module string, shift int) (hir.Expr, []SyntaxError) {
	if expr.signature != nil {
		body := expr.signature.Apply(expr.Args())
		return t.translateExpressionInModule(body, module, shift)
	}
	//
	return nil, t.srcmap.SyntaxErrors(expr, "unbound function")
}

func (t *translator) translateReduceInModule(expr *Reduce, module string, shift int) (hir.Expr, []SyntaxError) {
	if list, ok := expr.arg.(*List); !ok {
		return nil, t.srcmap.SyntaxErrors(expr.arg, "expected list")
	} else if sig := expr.signature; sig == nil {
		return nil, t.srcmap.SyntaxErrors(expr.arg, "unbound function")
	} else {
		reduction := list.Args[0]
		// Build reduction
		for i := 1; i < len(list.Args); i++ {
			reduction = sig.Apply([]Expr{reduction, list.Args[i]})
		}
		// Translate reduction
		return t.translateExpressionInModule(reduction, module, shift)
	}
}

func (t *translator) translateShiftInModule(expr *Shift, module string, shift int) (hir.Expr, []SyntaxError) {
	constant := expr.Shift.AsConstant()
	// Determine the shift constant
	if constant == nil {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "expected constant shift")
	} else if !constant.IsInt64() {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "constant shift too large")
	}
	// Now translate target expression with updated shift.
	return t.translateExpressionInModule(expr.Arg, module, shift+int(constant.Int64()))
}

func (t *translator) translateVariableAccessInModule(expr *VariableAccess, module string,
	shift int) (hir.Expr, []SyntaxError) {
	if binding, ok := expr.Binding().(*ColumnBinding); ok {
		// Lookup column binding
		cinfo := t.env.Column(binding.module, expr.Name())
		// Done
		return &hir.ColumnAccess{Column: cinfo.ColumnId(), Shift: shift}, nil
	} else if binding, ok := expr.Binding().(*ConstantBinding); ok {
		// Just fill in the constant.
		return t.translateExpressionInModule(binding.value, module, shift)
	}
	// error
	return nil, t.srcmap.SyntaxErrors(expr, "unbound variable")
}

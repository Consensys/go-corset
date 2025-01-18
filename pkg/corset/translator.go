package corset

import (
	"fmt"
	"math"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/hir"
	sc "github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/assignment"
	"github.com/consensys/go-corset/pkg/sexp"
	tr "github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
)

// TranslateCircuit translates the components of a Corset circuit and add them
// to the schema.  By the time we get to this point, all malformed source files
// should have been rejected already and the translation should go through
// easily.  Thus, whilst syntax errors can be returned here, this should never
// happen.  The mechanism is supported, however, to simplify development of new
// features, etc.
func TranslateCircuit(env Environment, srcmap *sexp.SourceMaps[Node],
	circuit *Circuit) (*hir.Schema, []SyntaxError) {
	//
	t := translator{env, srcmap, hir.EmptySchema()}
	// Allocate all modules into schema
	t.translateModules(circuit)
	// Translate input columns
	t.translateInputColumns(circuit)
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
		info := t.env.Module(m.Name)
		// Sanity check everything lines up.
		if info.Id != mid {
			panic(fmt.Sprintf("Invalid module identifier: %d vs %d", mid, info.Id))
		}
	}
}

// Translate all input column declarations in the entire circuit.
func (t *translator) translateInputColumns(circuit *Circuit) {
	t.translateInputColumnsInModule("")
	// Translate each module
	for _, m := range circuit.Modules {
		t.translateInputColumnsInModule(m.Name)
	}
}

// Translate all input column declarations occurring in a given module within
// the circuit.  Observe that this does not attempt to translate column
// declarations directly, since register allocation has broken the link between
// source-level columns and registers (i.e. HIR-level columns).  Instead, we
// must rely on information provided by the environment.
//
// Furthermore, we only allocate input columns here.  This is actually safe
// since (at this time) input columns are the only ones subject to register
// allocation.  In the future, this might change and a more involved strategy
// will be required (e.g. adding another level of indirection between the
// register indices generated from register allocation and those column
// identifiers used at the HIR level; or, requiring the column identifier be
// specified to HIR at the point of allocation).
func (t *translator) translateInputColumnsInModule(module string) {
	// Process each register in turn.
	for _, regIndex := range t.env.RegistersOf(module) {
		regInfo := t.env.Register(regIndex)
		// Sanity Check
		if !regInfo.IsActive() {
			panic("inactive register encountered")
		} else if regInfo.IsInput() {
			// Declare column at HIR level.
			cid := t.schema.AddDataColumn(regInfo.Context, regInfo.Name(), regInfo.DataType)
			// Prove underlying types (as necessary)
			t.translateTypeConstraints(regIndex)
			// Sanity check
			if cid != regIndex {
				// Should be unreachable
				panic(fmt.Sprintf("inconsistent register index (%d versus %d)", cid, regIndex))
			}
		}
	}
}

// Translate any type constraints applicable for the given register.  Type
// constraints are determined by the source-level columns and, hence, there are
// several cases to consider:
//
// (1) none of the source-level columns allocated to this register was marked
// provable. Therefore, no need to do anything.
//
// (2) all source-level columns allocated to this register which are marked
// provable have the same type which, furthermore, is the largest type of any
// column allocated to this register.  In this case, we can use a single
// (global) constraint for the entire column.
//
// (3) source-level columns allocated to this register which are marked provable
// have the same type, but this is not the largest of any allocated to this
// register.  In fact, only binary@prove is supported here and we can assume
// each column is allocated to a different perspective.
//
// Any other cases are considered to be erroneous register allocations, and will
// lead to a panic.
func (t *translator) translateTypeConstraints(regIndex uint) {
	regInfo := t.env.Register(regIndex)
	// NOTE: at the moment, a more restrictive set of requirements is assumed
	// (compared with the description above).  Specifically we assume that, for
	// any register needing a type constraint, the type of all source columns is
	// the same.  The intention is that, in the future, the logic here (and
	// potentially elsewhere) will be updated to support the weaker requirements
	// above.  In particular, to make use of the weaker requirements requires a
	// more powerful register allocator than currently implemented.
	required := false
	// Check for provability
	for _, col := range regInfo.Sources {
		if col.mustProve {
			required = true
			break
		}
	}
	// Apply provability (if it is required)
	if required {
		reg_width := regInfo.DataType.AsUint().BitWidth()
		// For now, enforce all source columns have matching bitwidth.
		for _, col := range regInfo.Sources {
			// Determine bitwidth
			col_width := col.datatype.AsUint().BitWidth()
			// Sanity check (for now)
			if col_width != reg_width {
				// See above discussion of why this strong restriction is
				// currently in place.
				panic("cannot (currently) prove type of mixed-width register")
			}
		}
		// Add appropriate type constraint
		bound := regInfo.DataType.AsUint().Bound()
		t.schema.AddRangeConstraint(regInfo.Name(), regInfo.Context, &hir.ColumnAccess{Column: regIndex, Shift: 0}, bound)
	}
}

// Translate all assignment or constraint declarations in the circuit.
func (t *translator) translateOtherDeclarations(circuit *Circuit) []SyntaxError {
	rootPath := util.NewAbsolutePath()
	errors := t.translateOtherDeclarationsInModule(rootPath, circuit.Declarations)
	// Translate each module
	for _, m := range circuit.Modules {
		modPath := rootPath.Extend(m.Name)
		errs := t.translateOtherDeclarationsInModule(*modPath, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate all assignment or constraint declarations in a given module within
// the circuit.
func (t *translator) translateOtherDeclarationsInModule(module util.Path, decls []Declaration) []SyntaxError {
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
func (t *translator) translateDeclaration(decl Declaration, module util.Path) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *DefAliases:
		// Not an assignment or a constraint, hence ignore.
	case *DefComputed:
		errors = t.translateDefComputed(d, module)
	case *DefColumns:
		// Not an assignment or a constraint, hence ignore.
	case *DefConst:
		// For now, constants are always compiled out when going down to HIR.
	case *DefConstraint:
		errors = t.translateDefConstraint(d, module)
	case *DefFun:
		// For now, functions are always compiled out when going down to HIR.
		// In the future, this might change if we add support for macros to HIR.
	case *DefInRange:
		errors = t.translateDefInRange(d, module)
	case *DefInterleaved:
		errors = t.translateDefInterleaved(d, module)
	case *DefLookup:
		errors = t.translateDefLookup(d, module)
	case *DefPermutation:
		errors = t.translateDefPermutation(d, module)
	case *DefPerspective:
		// As for defcolumns, nothing generated here.
	case *DefProperty:
		errors = t.translateDefProperty(d, module)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Translate a "defcomputed" declaration.
func (t *translator) translateDefComputed(decl *DefComputed, module util.Path) []SyntaxError {
	var (
		errors   []SyntaxError
		context  tr.Context = tr.VoidContext[uint]()
		firstCid uint
	)
	//
	targets := make([]sc.Column, len(decl.Targets))
	sources := make([]uint, len(decl.Sources))
	// Identify source columns
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i].Binding().(*ColumnBinding)
		sources[i] = t.env.RegisterOf(&ith.path)
	}
	// Identify target columns
	for i := 0; i < len(decl.Targets); i++ {
		targetPath := module.Extend(decl.Targets[i].Name())
		targetId := t.env.RegisterOf(targetPath)
		target := t.env.Register(targetId)
		// Construct columns
		targets[i] = sc.NewColumn(target.Context, target.Name(), target.DataType)
		// Record first CID
		if i == 0 {
			firstCid = targetId
		}
		// Join contexts
		context = context.Join(target.Context)
	}
	// Extract the binding
	binding := decl.Function.Binding().(*NativeDefinition)
	// Add the assignment and check the first identifier.
	cid := t.schema.AddAssignment(assignment.NewComputation(context, binding.name, targets, sources))
	// Sanity check column identifiers align.
	if cid != firstCid {
		err := fmt.Sprintf("inconsistent (computed) column identifier (%d v %d)", cid, firstCid)
		errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	}
	// Done
	return errors
}

// Translate a "defconstraint" declaration.
func (t *translator) translateDefConstraint(decl *DefConstraint, module util.Path) []SyntaxError {
	// Translate constraint body
	constraint, errors := t.translateExpressionInModule(decl.Constraint, module, 0)
	// Translate (optional) guard
	guard, guard_errors := t.translateOptionalExpressionInModule(decl.Guard, module, 0)
	// Translate (optional) perspective selector
	selector, selector_errors := t.translateSelectorInModule(decl.Perspective, module)
	// Combine errors
	errors = append(errors, guard_errors...)
	errors = append(errors, selector_errors...)
	// Apply guard
	if constraint == nil {
		// NOTE: in this case, the constraint itself has been translated as nil.
		// This means there is no constraint (e.g. its a debug constraint, but
		// debug mode is not enabled).
		return errors
	}
	// Apply guard (if applicable)
	if guard != nil {
		constraint = &hir.Mul{Args: []hir.Expr{guard, constraint}}
	}
	// Apply perspective selector (if applicable)
	if selector != nil {
		constraint = &hir.Mul{Args: []hir.Expr{selector, constraint}}
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
		} else {
			// Add translated constraint
			t.schema.AddVanishingConstraint(decl.Handle, context, decl.Domain, constraint)
		}
	}
	// Done
	return errors
}

// Translate the selector for the perspective of a defconstraint.  Observe that
// a defconstraint may not be part of a perspective and, hence, would have no
// selector.
func (t *translator) translateSelectorInModule(perspective *PerspectiveName,
	module util.Path) (hir.Expr, []SyntaxError) {
	//
	if perspective != nil {
		return t.translateExpressionInModule(perspective.binding.selector, module, 0)
	}
	//
	return nil, nil
}

// Translate a "deflookup" declaration.
//
//nolint:staticcheck
func (t *translator) translateDefLookup(decl *DefLookup, module util.Path) []SyntaxError {
	// Translate source expressions
	sources, src_errs := t.translateUnitExpressionsInModule(decl.Sources, module, 0)
	targets, tgt_errs := t.translateUnitExpressionsInModule(decl.Targets, module, 0)
	// Combine errors
	errors := append(src_errs, tgt_errs...)
	//
	if len(errors) == 0 {
		src_context := t.env.ContextOf(ContextOfExpressions(decl.Sources))
		target_context := t.env.ContextOf(ContextOfExpressions(decl.Targets))
		// Add translated constraint
		t.schema.AddLookupConstraint(decl.Handle, src_context, target_context, sources, targets)
	}
	// Done
	return errors
}

// Translate a "definrange" declaration.
func (t *translator) translateDefInRange(decl *DefInRange, module util.Path) []SyntaxError {
	// Translate constraint body
	expr, errors := t.translateExpressionInModule(decl.Expr, module, 0)
	//
	if len(errors) == 0 {
		context := expr.Context(t.schema)
		// Add translated constraint
		t.schema.AddRangeConstraint("", context, expr, decl.Bound)
	}
	// Done
	return errors
}

// Translate a "definterleaved" declaration.
func (t *translator) translateDefInterleaved(decl *DefInterleaved, module util.Path) []SyntaxError {
	var errors []SyntaxError
	//
	sources := make([]uint, len(decl.Sources))
	// Lookup target column info
	targetPath := module.Extend(decl.Target.Name())
	targetId := t.env.RegisterOf(targetPath)
	target := t.env.Register(targetId)
	// Determine source column identifiers
	for i, source := range decl.Sources {
		var errs []SyntaxError
		sources[i], errs = t.registerOfColumnAccess(source)
		errors = append(errors, errs...)
	}
	// Register assignment
	cid := t.schema.AddAssignment(assignment.NewInterleaving(target.Context, target.Name(), sources, target.DataType))
	// Sanity check column identifiers align.
	if cid != targetId {
		err := fmt.Sprintf("inconsitent (interleaved) column identifier (%d v %d)", cid, targetId)
		errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	}
	// Done
	return errors
}

// Translate a "defpermutation" declaration.
func (t *translator) translateDefPermutation(decl *DefPermutation, module util.Path) []SyntaxError {
	var (
		errors   []SyntaxError
		context  tr.Context = tr.VoidContext[uint]()
		firstCid uint
	)
	//
	targets := make([]sc.Column, len(decl.Sources))
	signs := make([]bool, len(decl.Sources))
	sources := make([]uint, len(decl.Sources))
	//
	for i := 0; i < len(decl.Sources); i++ {
		targetPath := module.Extend(decl.Targets[i].Name())
		targetId := t.env.RegisterOf(targetPath)
		target := t.env.Register(targetId)
		// Construct columns
		targets[i] = sc.NewColumn(target.Context, target.Name(), target.DataType)
		sourceBinding := decl.Sources[i].Binding().(*ColumnBinding)
		sources[i] = t.env.RegisterOf(&sourceBinding.path)
		signs[i] = decl.Signs[i]
		// Record first CID
		if i == 0 {
			firstCid = targetId
		}
		// Join contexts
		context = context.Join(target.Context)
	}
	// Add the assignment and check the first identifier.
	cid := t.schema.AddAssignment(assignment.NewSortedPermutation(context, targets, signs, sources))
	// Sanity check column identifiers align.
	if cid != firstCid {
		err := fmt.Sprintf("inconsistent (permuted) column identifier (%d v %d)", cid, firstCid)
		errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	}
	// Done
	return errors
}

// Translate a "defproperty" declaration.
func (t *translator) translateDefProperty(decl *DefProperty, module util.Path) []SyntaxError {
	// Translate constraint body
	assertion, errors := t.translateExpressionInModule(decl.Assertion, module, 0)
	//
	if len(errors) == 0 {
		context := assertion.Context(t.schema)
		// Add translated constraint
		t.schema.AddPropertyAssertion(decl.Handle, context, assertion)
	}
	// Done
	return errors
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateOptionalExpressionInModule(expr Expr, module util.Path,
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
func (t *translator) translateUnitExpressionsInModule(exprs []Expr, module util.Path,
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
func (t *translator) translateExpressionsInModule(exprs []Expr, module util.Path,
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

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for column access, function
// invocations, etc).
func (t *translator) translateExpressionInModule(expr Expr, module util.Path, shift int) (hir.Expr, []SyntaxError) {
	switch e := expr.(type) {
	case *ArrayAccess:
		// Lookup underlying column info
		registerId, errors := t.registerOfColumnAccess(e)
		// Done
		return &hir.ColumnAccess{Column: registerId, Shift: shift}, errors
	case *Add:
		args, errs := t.translateExpressionsInModule(e.Args, module, shift)
		return &hir.Add{Args: args}, errs
	case *Constant:
		var val fr.Element
		// Initialise field from bigint
		val.SetBigInt(&e.Val)
		//
		return &hir.Constant{Val: val}, nil
	case *Exp:
		return t.translateExpInModule(e, module, shift)
	case *If:
		args, errs := t.translateExpressionsInModule([]Expr{e.Condition, e.TrueBranch, e.FalseBranch}, module, shift)
		// Construct appropriate if form
		if e.IsIfZero() {
			return &hir.IfZero{Condition: args[0], TrueBranch: args[1], FalseBranch: args[2]}, errs
		} else if e.IsIfNotZero() {
			// In this case, switch the ordering.
			return &hir.IfZero{Condition: args[0], TrueBranch: args[2], FalseBranch: args[1]}, errs
		}
		// Should be unreachable
		return nil, t.srcmap.SyntaxErrors(expr, "unresolved conditional encountered during translation")
	case *List:
		args, errs := t.translateExpressionsInModule(e.Args, module, shift)
		return &hir.List{Args: args}, errs
	case *Mul:
		args, errs := t.translateExpressionsInModule(e.Args, module, shift)
		return &hir.Mul{Args: args}, errs
	case *Normalise:
		arg, errs := t.translateExpressionInModule(e.Arg, module, shift)
		return &hir.Normalise{Arg: arg}, errs
	case *Sub:
		args, errs := t.translateExpressionsInModule(e.Args, module, shift)
		return &hir.Sub{Args: args}, errs
	case *Shift:
		return t.translateShiftInModule(e, module, shift)
	case *VariableAccess:
		return t.translateVariableAccessInModule(e, shift)
	default:
		return nil, t.srcmap.SyntaxErrors(expr, "unknown expression encountered during translation")
	}
}

func (t *translator) translateExpInModule(expr *Exp, module util.Path, shift int) (hir.Expr, []SyntaxError) {
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

func (t *translator) translateShiftInModule(expr *Shift, module util.Path, shift int) (hir.Expr, []SyntaxError) {
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

func (t *translator) translateVariableAccessInModule(expr *VariableAccess, shift int) (hir.Expr, []SyntaxError) {
	if binding, ok := expr.Binding().(*ColumnBinding); ok {
		// Lookup column binding
		register_id := t.env.RegisterOf(binding.AbsolutePath())
		// Done
		return &hir.ColumnAccess{Column: register_id, Shift: shift}, nil
	} else if binding, ok := expr.Binding().(*ConstantBinding); ok {
		// Just fill in the constant.
		var constant fr.Element
		// Initialise field from bigint
		constant.SetBigInt(binding.value.AsConstant())
		//
		return &hir.Constant{Val: constant}, nil
	}
	// error
	return nil, t.srcmap.SyntaxErrors(expr, "unbound variable")
}

// Determine the underlying register for a symbol which represents a column access.
func (t *translator) registerOfColumnAccess(symbol Symbol) (uint, []SyntaxError) {
	switch e := symbol.(type) {
	case *ArrayAccess:
		return t.registerOfArrayAccess(e)
	case *VariableAccess:
		return t.registerOfVariableAccess(e)
	}
	//
	return math.MaxUint, t.srcmap.SyntaxErrors(symbol, "invalid column access")
}

func (t *translator) registerOfVariableAccess(expr *VariableAccess) (uint, []SyntaxError) {
	if binding, ok := expr.Binding().(*ColumnBinding); ok {
		// Lookup column binding
		return t.env.RegisterOf(binding.AbsolutePath()), nil
	}
	//
	return math.MaxUint, t.srcmap.SyntaxErrors(expr, "invalid column access")
}
func (t *translator) registerOfArrayAccess(expr *ArrayAccess) (uint, []SyntaxError) {
	var (
		errors []SyntaxError
		min    uint = 0
		max    uint = math.MaxUint
	)
	// Lookup the column
	binding, ok := expr.Binding().(*ColumnBinding)
	// Did we find it?
	if !ok {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "invalid array index encountered during translation"))
	} else if arr_t, ok := binding.dataType.(*ArrayType); ok {
		min = arr_t.min
		max = arr_t.max
	}
	// Array index should be statically known
	index := expr.arg.AsConstant()
	//
	if index == nil {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "expected constant array index"))
	} else if i := uint(index.Uint64()); i < min || i > max {
		errors = append(errors, *t.srcmap.SyntaxError(expr.arg, "array index out-of-bounds"))
	}
	// Error check
	if len(errors) > 0 {
		return math.MaxUint, errors
	}
	// Construct real column name
	path := &binding.path
	name := fmt.Sprintf("%s_%d", path.Tail(), index.Uint64())
	path = path.Parent().Extend(name)
	// Lookup underlying column info
	return t.env.RegisterOf(path), errors
}

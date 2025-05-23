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
package compiler

import (
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/source"
)

// SchemaBuilder is used within this translator for building the final mixed MIR
// schema.
type SchemaBuilder = ir.SchemaBuilder[mir.Constraint, mir.Term]

// ModuleBuilder is used within this translator for building the various modules
// which are contained within the mixed MIR schema.
type ModuleBuilder = ir.ModuleBuilder[mir.Constraint, mir.Term]

// TranslateCircuit translates the components of a Corset circuit and add them
// to the schema.  By the time we get to this point, all malformed source files
// should have been rejected already and the translation should go through
// easily.  Thus, whilst syntax errors can be returned here, this should never
// happen.  The mechanism is supported, however, to simplify development of new
// features, etc.
func TranslateCircuit[M schema.Module](
	env Environment,
	srcmap *source.Maps[ast.Node],
	circuit *ast.Circuit,
	externs ...M) (schema.MixedSchema[M, mir.Module], []SyntaxError) {
	//
	builder := ir.NewSchemaBuilder[mir.Constraint, mir.Term]()
	t := translator{env, srcmap, builder}
	// Allocate all modules into schema
	t.translateModules(circuit)
	// Translate everything else
	if errs := t.translateDeclarations(circuit); len(errs) > 0 {
		return schema.MixedSchema[M, mir.Module]{}, errs
	}
	// Finally, construct the mixed schema
	return schema.NewMixedSchema(externs, t.schema.Build()), nil
}

// Translator packages up information necessary for translating a circuit into
// the schema form required for the HIR level.
type translator struct {
	// Environment is needed for determining the identifiers for modules and
	// registers.
	env Environment
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *source.Maps[ast.Node]
	// Represents the schema being constructed by this translator.
	schema SchemaBuilder
}

func (t *translator) translateModules(circuit *ast.Circuit) {
	// Add root module
	t.translateModule("")
	// Add nested modules
	for _, m := range circuit.Modules {
		// Translate module condition (if applicable)
		if m.Condition != nil {
			panic("conditional modules not supported")
		}
		//
		t.translateModule(m.Name)
	}
}

func (t *translator) translateModule(name string) {
	//
	mid := t.schema.NewModule(name)
	info := t.env.Module(name)
	// Sanity check everything lines up.
	if info.Id != mid {
		// NOTE: this should fail now
		panic(fmt.Sprintf("Invalid module identifier: %d vs %d", mid, info.Id))
	}
	// Allocate module registers
	module := t.schema.Module(mid)
	// Process each register in turn.
	for _, regIndex := range t.env.RegistersOf(name) {
		regInfo := t.env.Register(regIndex)
		// Declare corresponding register
		module.NewRegister(schema.NewInputRegister(regInfo.Name(), regInfo.Bitwidth))
		// Prove underlying types (as necessary)
		t.translateTypeConstraints(*regInfo, module)
	}
}

// Translate any type constraints applicable for the given register.  Type
// constraints are determined by the source-level registers and, hence, there are
// several cases to consider:
//
// (1) none of the source-level registers allocated to this register was marked
// provable. Therefore, no need to do anything.
//
// (2) all source-level registers allocated to this register which are marked
// provable have the same type which, furthermore, is the largest type of any
// register allocated to this register.  In this case, we can use a single
// (global) constraint for the entire register.
//
// (3) source-level registers allocated to this register which are marked provable
// have the same type, but this is not the largest of any allocated to this
// register.  In fact, only binary@prove is supported here and we can assume
// each register is allocated to a different perspective.
//
// Any other cases are considered to be erroneous register allocations, and will
// lead to a panic.
func (t *translator) translateTypeConstraints(reg Register, mod *ModuleBuilder) {
	required := false
	// Check for provability
	for _, col := range reg.Sources {
		if col.MustProve {
			required = true
			break
		}
	}
	// Apply provability (if it is required)
	if required {
		reg_width := reg.Bitwidth
		// For now, enforce all source registers have matching bitwidth.
		for _, col := range reg.Sources {
			// Determine bitwidth
			col_width := col.Bitwidth
			// Sanity check (for now)
			if col.MustProve && col_width != reg_width {
				// Currently, mixed-width proving types are not supported.
				panic("cannot (currently) prove type of mixed-width register")
			}
		}
		// Add appropriate type constraint
		constraint := mir.NewRangeConstraint(reg.Name(),
			reg.Context,
			mod.RegisterAccessOf(reg.Name(), 0),
			reg.Bitwidth)
		//
		mod.AddConstraint(constraint)
	}
}

// Translate all assignment or constraint declarations in the circuit.
func (t *translator) translateDeclarations(circuit *ast.Circuit) []SyntaxError {
	errors := t.translateDeclarationsInModule("", circuit.Declarations)
	// Translate each module
	for _, m := range circuit.Modules {
		errs := t.translateDeclarationsInModule(m.Name, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate all assignment or constraint declarations in a given module within
// the circuit.
func (t *translator) translateDeclarationsInModule(module string, decls []ast.Declaration) []SyntaxError {
	var (
		errors []SyntaxError
		mod    = t.schema.ModuleOf(module)
	)
	//
	for _, d := range decls {
		errs := t.translateDeclaration(d, mod)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate an assignment or constraint declarartion which occurs within a
// given module.
func (t *translator) translateDeclaration(decl ast.Declaration, module *ModuleBuilder) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *ast.DefAliases:
		// Not an assignment or a constraint, hence ignore.
	case *ast.DefComputed:
		errors = t.translateDefComputed(d, module)
	case *ast.DefColumns:
		// Not an assignment or a constraint, hence ignore.
	case *ast.DefConst:
		// For now, constants are always compiled out when going down to mir.
	case *ast.DefConstraint:
		errors = t.translateDefConstraint(d, module)
	case *ast.DefFun:
		// For now, functions are always compiled out when going down to mir.
		// In the future, this might change if we add support for macros to mir.
	case *ast.DefInRange:
		errors = t.translateDefInRange(d, module)
	case *ast.DefInterleaved:
		errors = t.translateDefInterleaved(d, module)
	case *ast.DefLookup:
		errors = t.translateDefLookup(d, module)
	case *ast.DefPermutation:
		errors = t.translateDefPermutation(d, module)
	case *ast.DefPerspective:
		// As for defregisters, nothing generated here.
	case *ast.DefProperty:
		errors = t.translateDefProperty(d, module)
	case *ast.DefSorted:
		errors = t.translateDefSorted(d, module)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Translate a "defcomputed" declaration.
func (t *translator) translateDefComputed(decl *ast.DefComputed, module *ModuleBuilder) []SyntaxError {
	// var (
	// 	errors   []SyntaxError
	// 	context  tr.Context = tr.VoidContext[uint]()
	// 	firstCid uint
	// )
	// //
	// targets := make([]sc.Register, len(decl.Targets))
	// sources := make([]uint, len(decl.Sources))
	// // Identify source registers
	// for i := 0; i < len(decl.Sources); i++ {
	// 	ith := decl.Sources[i].Binding().(*ast.RegisterBinding)
	// 	sources[i] = t.env.RegisterOf(&ith.Path)
	// }
	// // Identify target registers
	// for i := 0; i < len(decl.Targets); i++ {
	// 	targetPath := module.Extend(decl.Targets[i].Name())
	// 	targetId := t.env.RegisterOf(targetPath)
	// 	target := t.env.Register(targetId)
	// 	// Construct registers
	// 	targets[i] = sc.NewRegister(target.Context, target.Name(), target.DataType)
	// 	// Record first CID
	// 	if i == 0 {
	// 		firstCid = targetId
	// 	}
	// 	// Join contexts
	// 	context = context.Join(target.Context)
	// }
	// // Extract the binding
	// binding := decl.Function.Binding().(*NativeDefinition)
	// // Add the assignment and check the first identifier.
	// cid := t.schema.AddAssignment(assignment.NewComputation(context, binding.name, targets, sources))
	// // Sanity check register identifiers align.
	// if cid != firstCid {
	// 	err := fmt.Sprintf("inconsistent (computed) register identifier (%d v %d)", cid, firstCid)
	// 	errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	// }
	// // Done
	// return errors
	panic("todo")
}

// Translate a "defconstraint" declaration.
func (t *translator) translateDefConstraint(decl *ast.DefConstraint, module *ModuleBuilder) []SyntaxError {
	// Translate expr body
	expr, errors := t.translateLogical(decl.Constraint, module, 0)
	// Translate (optional) guard
	guard, guard_errors := t.translateOptionalLogical(decl.Guard, module, 0)
	// Translate (optional) perspective selector
	selector, selector_errors := t.translateSelectorInModule(decl.Perspective, module)
	// Combine errors
	errors = append(errors, guard_errors...)
	errors = append(errors, selector_errors...)
	// Apply guard
	if expr == nil {
		// NOTE: in this case, the constraint itself has been translated as nil.
		// This means there is no constraint (e.g. its a debug constraint, but
		// debug mode is not enabled).
		return errors
	}
	// Apply guard (if applicable)
	if guard != nil {
		// guard = ir.Equals[mir.LogicalTerm, mir.Term](guard, ir.Const64[mir.Term](0))
		// expr = ir.IfElse(guard, nil, expr)
		panic("todo")
	}
	// Apply perspective selector (if applicable)
	if selector != nil {
		// selector = mir.Equals(selector, ir.Const64[mir.Term](0))
		// expr = ir.IfElse(selector, nil, expr)
		panic("todo")
	}
	//
	if len(errors) == 0 {
		// FIXME: this could be more efficient!!
		context := t.env.ContextOf(decl.Constraint.Context())
		// Add translated constraint
		module.AddConstraint(mir.NewVanishingConstraint(decl.Handle, context, decl.Domain, expr))
	}
	// Done
	return errors
}

// Translate the selector for the perspective of a defconstraint.  Observe that
// a defconstraint may not be part of a perspective and, hence, would have no
// selector.
func (t *translator) translateSelectorInModule(perspective *ast.PerspectiveName,
	module *ModuleBuilder) (mir.Term, []SyntaxError) {
	//
	if perspective != nil {
		return t.translateExpression(perspective.InnerBinding().Selector, module, 0)
	}
	//
	return nil, nil
}

// Translate a "deflookup" declaration.
func (t *translator) translateDefLookup(decl *ast.DefLookup, module *ModuleBuilder) []SyntaxError {
	var (
		errors     []SyntaxError
		srcContext trace.Context
		dstContext trace.Context
		sources    []mir.Term
		targets    []mir.Term
	)
	// Determine source and target modules for this lookup.
	srcAstContext, i := ast.ContextOfExpressions(decl.Sources...)
	dstAstContext, j := ast.ContextOfExpressions(decl.Targets...)
	// Translate source expressions whilst checking for a conflicting context.
	// This can arise here, rather than in the resolve, in some unusual
	// situations (e.g. source expression is a function).
	if srcAstContext.IsConflicted() {
		errors = append(errors, *t.srcmap.SyntaxError(decl.Sources[i], "conflicting context"))
	} else {
		var errs []SyntaxError
		//
		srcContext = t.env.ContextOf(srcAstContext)
		srcModule := t.schema.Module(srcContext.ModuleId)
		//
		sources, errs = t.translateUnitExpressions(decl.Sources, srcModule, 0)
		errors = append(errors, errs...)
	}
	// Translate target expressions whilst again checking for a conflicting
	// context.
	if dstAstContext.IsConflicted() {
		errors = append(errors, *t.srcmap.SyntaxError(decl.Targets[j], "conflicting context"))
	} else {
		var errs []SyntaxError
		//
		dstContext = t.env.ContextOf(dstAstContext)
		dstModule := t.schema.Module(dstContext.ModuleId)
		//
		targets, errs = t.translateUnitExpressions(decl.Sources, dstModule, 0)
		errors = append(errors, errs...)
	}
	// Sanity check whether we can construct the constraint, or not.
	if len(errors) == 0 {
		// Add translated constraint
		module.AddConstraint(mir.NewLookupConstraint(decl.Handle,
			srcContext,
			dstContext,
			sources,
			targets))
	}
	// Done
	return errors
}

// Translate a "definrange" declaration.
func (t *translator) translateDefInRange(decl *ast.DefInRange, module *ModuleBuilder) []SyntaxError {
	// Translate constraint body
	expr, errors := t.translateExpression(decl.Expr, module, 0)
	//
	if len(errors) == 0 {
		// FIXME: this could be more efficient!!
		context := t.env.ContextOf(decl.Expr.Context())
		// Add translated constraint
		module.AddConstraint(mir.NewRangeConstraint("", context, expr, decl.Bitwidth))
	}
	// Done
	return errors
}

// Translate a "definterleaved" declaration.
func (t *translator) translateDefInterleaved(decl *ast.DefInterleaved, module *ModuleBuilder) []SyntaxError {
	// var errors []SyntaxError
	// //
	// sources := make([]uint, len(decl.Sources))
	// // Lookup target register info
	// targetPath := module.Extend(decl.Target.Name())
	// targetId := t.env.RegisterOf(targetPath)
	// target := t.env.Register(targetId)
	// // Determine source register identifiers
	// for i, source := range decl.Sources {
	// 	var errs []SyntaxError
	// 	sources[i], errs = t.registerOfRegisterAccess(source)
	// 	errors = append(errors, errs...)
	// }
	// // Register assignment
	// cid := t.schema.AddAssignment(assignment.NewInterleaving(target.Context, target.Name(), sources, target.DataType))
	// // Sanity check register identifiers align.
	// if cid != targetId {
	// 	err := fmt.Sprintf("inconsitent (interleaved) register identifier (%d v %d)", cid, targetId)
	// 	errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	// }
	// // Done
	// return errors
	panic("todo")
}

// Translate a "defpermutation" declaration.
func (t *translator) translateDefPermutation(decl *ast.DefPermutation, module *ModuleBuilder) []SyntaxError {
	// var (
	// 	errors   []SyntaxError
	// 	context  tr.Context = tr.VoidContext[uint]()
	// 	firstCid uint
	// )
	// //
	// targets := make([]sc.Register, len(decl.Sources))
	// sources := make([]uint, len(decl.Sources))
	// //
	// for i := 0; i < len(decl.Sources); i++ {
	// 	targetPath := module.Extend(decl.Targets[i].Name())
	// 	targetId := t.env.RegisterOf(targetPath)
	// 	target := t.env.Register(targetId)
	// 	// Construct registers
	// 	targets[i] = sc.NewRegister(target.Context, target.Name(), target.DataType)
	// 	sourceBinding := decl.Sources[i].Binding().(*ast.RegisterBinding)
	// 	sources[i] = t.env.RegisterOf(&sourceBinding.Path)
	// 	// Record first CID
	// 	if i == 0 {
	// 		firstCid = targetId
	// 	}
	// 	// Join contexts
	// 	context = context.Join(target.Context)
	// }
	// // Clone the signs
	// signs := slices.Clone(decl.Signs)
	// // Add the assignment and check the first identifier.
	// cid := t.schema.AddAssignment(assignment.NewSortedPermutation(context, targets, signs, sources))
	// // Sanity check register identifiers align.
	// if cid != firstCid {
	// 	err := fmt.Sprintf("inconsistent (permuted) register identifier (%d v %d)", cid, firstCid)
	// 	errors = append(errors, *t.srcmap.SyntaxError(decl, err))
	// }
	// // Done
	// return errors
	panic("todo")
}

// Translate a "defproperty" declaration.
func (t *translator) translateDefProperty(decl *ast.DefProperty, module *ModuleBuilder) []SyntaxError {
	// // Translate constraint body
	// assertion, errors := t.translateExpressionInModule(decl.Assertion, module, 0)
	// //
	// if len(errors) == 0 {
	// 	context := assertion.Context(t.schema)
	// 	// Add translated constraint
	// 	t.schema.AddPropertyAssertion(decl.Handle, context, assertion)
	// }
	// // Done
	// return errors
	panic("todo")
}

// Translate a "defsorted" declaration.
func (t *translator) translateDefSorted(decl *ast.DefSorted, module *ModuleBuilder) []SyntaxError {
	var selector util.Option[mir.Term]
	// Translate source expressions
	sources, errors := t.translateUnitExpressions(decl.Sources, module, 0)
	// Translate (optional) selector expression
	if decl.Selector.HasValue() {
		sel, errs := t.translateExpression(decl.Selector.Unwrap(), module, 0)
		selector = util.Some(sel)
		//
		errors = append(errors, errs...)
	}
	// Determine source context
	src_ctx, i := ast.ContextOfExpressions(decl.Sources...)
	// Sanity check
	if src_ctx.IsConflicted() {
		errors = append(errors, *t.srcmap.SyntaxError(decl.Sources[i], "conflicting context"))
	}
	// Create construct (assuming no errors thus far)
	if len(errors) == 0 {
		context := t.env.ContextOf(src_ctx)
		// Clone the signs
		signs := slices.Clone(decl.Signs)
		bitwidth := determineMaxBitwidth(module, sources[:len(signs)])
		// Add translated constraint
		module.AddConstraint(
			mir.NewSortedConstraint(decl.Handle, context, bitwidth, selector, sources, signs, decl.Strict))
	}
	// Done
	return errors
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateUnitExpressions(exprs []ast.Expr, module *ModuleBuilder,
	shift int) ([]mir.Term, []SyntaxError) {
	//
	errors := []SyntaxError{}
	hirExprs := make([]mir.Term, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			expr, errs := t.translateExpression(e, module, shift)
			errors = append(errors, errs...)
			hirExprs[i] = expr
		}
	}
	// Done
	return hirExprs, errors
}

// Translate a sequence of zero or more expressions enclosed in a given module.
func (t *translator) translateExpressions(module *ModuleBuilder, shift int,
	exprs ...ast.Expr) ([]mir.Term, []SyntaxError) {
	//
	errors := []SyntaxError{}
	nexprs := make([]mir.Term, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		if e != nil {
			var errs []SyntaxError
			nexprs[i], errs = t.translateExpression(e, module, shift)
			errors = append(errors, errs...)
		} else {
			// Strictly speaking, this assignment is unnecessary.  However, the
			// purpose is just to make it clear what's going on.
			nexprs[i] = nil
		}
	}
	//
	return nexprs, errors
}

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for register access, function
// invocations, etc).
func (t *translator) translateExpression(expr ast.Expr, module *ModuleBuilder, shift int) (mir.Term, []SyntaxError) {
	switch e := expr.(type) {
	case *ast.ArrayAccess:
		// Lookup underlying register info
		return t.registerOfRegisterAccess(e, shift)
	case *ast.Add:
		args, errs := t.translateExpressions(module, shift, e.Args...)
		return ir.Sum(args...), errs
	case *ast.Cast:
		arg, errs := t.translateExpression(e.Arg, module, shift)
		//
		if !e.Unsafe {
			// safe casts are compiled out since they have already been checked
			// by the type checker.
			return arg, errs
		} else if int_t, ok := e.Type.(*ast.IntType); ok {
			// unsafe casts cannot be checked by the type checker, but can be
			// exploited for the purposes of optimisation.
			return ir.CastOf(arg, int_t.BitWidth()), errs
		}
		// Should be unreachable.
		msg := fmt.Sprintf("cannot translate cast (%s)", e.Type.String())
		//
		return nil, t.srcmap.SyntaxErrors(expr, msg)
	case *ast.Constant:
		var val fr.Element
		// Initialise field from bigint
		val.SetBigInt(&e.Val)
		//
		return ir.Const[mir.Term](val), nil
	case *ast.Exp:
		return t.translateExp(e, module, shift)
	case *ast.If:
		return t.translateIf(e, module, shift)
	case *ast.Mul:
		args, errs := t.translateExpressions(module, shift, e.Args...)
		return ir.Product(args...), errs
	case *ast.Normalise:
		arg, errs := t.translateExpression(e.Arg, module, shift)
		return ir.Normalise(arg), errs
	case *ast.Sub:
		args, errs := t.translateExpressions(module, shift, e.Args...)
		return ir.Subtract(args...), errs
	case *ast.Shift:
		return t.translateShift(e, module, shift)
	case *ast.VariableAccess:
		return t.translateVariableAccess(e, shift)
	default:
		typeStr := reflect.TypeOf(expr).String()
		msg := fmt.Sprintf("unknown expression encountered during translation (%s)", typeStr)
		//
		return nil, t.srcmap.SyntaxErrors(expr, msg)
	}
}

func (t *translator) translateExp(expr *ast.Exp, module *ModuleBuilder, shift int) (mir.Term, []SyntaxError) {
	arg, errs := t.translateExpression(expr.Arg, module, shift)
	pow := expr.Pow.AsConstant()
	// Identity constant for pow
	if pow == nil {
		errs = append(errs, *t.srcmap.SyntaxError(expr.Pow, "expected constant power"))
	} else if !pow.IsUint64() {
		errs = append(errs, *t.srcmap.SyntaxError(expr.Pow, "constant power too large"))
	}
	// Sanity check errors
	if len(errs) == 0 {
		return ir.Exponent(arg, pow.Uint64()), errs
	}
	//
	return nil, errs
}

func (t *translator) translateIf(expr *ast.If, module *ModuleBuilder, shift int) (mir.Term, []SyntaxError) {
	// fall-back translation condition
	args, errs := t.translateExpressions(module, shift, expr.Condition, expr.TrueBranch, expr.FalseBranch)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Propagate emptiness (if applicable)
	if args[1] == nil && args[2] == nil {
		return nil, nil
	}
	// Construct appropriate if form
	return ir.IfElse(args[0], args[1], args[2]), nil
}

func (t *translator) translateShift(expr *ast.Shift, mod *ModuleBuilder, shift int) (mir.Term, []SyntaxError) {
	constant := expr.Shift.AsConstant()
	// Determine the shift constant
	if constant == nil {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "expected constant shift")
	} else if !constant.IsInt64() {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "constant shift too large")
	}
	// Now translate target expression with updated shift.
	return t.translateExpression(expr.Arg, mod, shift+int(constant.Int64()))
}

func (t *translator) translateVariableAccess(expr *ast.VariableAccess, shift int) (mir.Term, []SyntaxError) {
	if _, ok := expr.Binding().(*ast.ColumnBinding); ok {
		return t.registerOfVariableAccess(expr, shift)
	} else if binding, ok := expr.Binding().(*ast.ConstantBinding); ok {
		// Just fill in the constant.
		var constant fr.Element
		// Initialise field from bigint
		constant.SetBigInt(binding.Value.AsConstant())
		// Handle externalised constants slightly differently.
		if binding.Extern {
			//
			return ir.LabelledConstant[mir.Term](binding.Path.String(), constant), nil
		}
		//
		return ir.Const[mir.Term](constant), nil
	}
	// error
	return nil, t.srcmap.SyntaxErrors(expr, "unbound variable")
}

// Translate a sequence of zero or more logical expressions enclosed in a given module.
func (t *translator) translateLogicals(module *ModuleBuilder, shift int,
	exprs ...ast.Expr) ([]mir.LogicalTerm, []SyntaxError) {
	//
	errors := []SyntaxError{}
	logicals := make([]mir.LogicalTerm, len(exprs))
	// Iterate each expression in turn
	for i, e := range exprs {
		var errs []SyntaxError
		logicals[i], errs = t.translateLogical(e, module, shift)
		errors = append(errors, errs...)
	}
	//
	return logicals, errors
}

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for register access, function
// invocations, etc).
func (t *translator) translateLogical(expr ast.Expr, mod *ModuleBuilder, shift int) (mir.LogicalTerm, []SyntaxError) {
	switch e := expr.(type) {
	case *ast.Connective:
		args, errs := t.translateLogicals(mod, shift, e.Args...)
		//
		if e.Sign {
			return ir.Disjunction(args...), errs
		}
		//
		return ir.Conjunction(args...), errs
	case *ast.Equation:
		lhs, errs1 := t.translateExpression(e.Lhs, mod, shift)
		rhs, errs2 := t.translateExpression(e.Rhs, mod, shift)
		errs := append(errs1, errs2...)
		//
		if len(errs) > 0 {
			return nil, errs
		}
		//
		switch e.Kind {
		case ast.EQUALS:
			return ir.Equals[mir.LogicalTerm](lhs, rhs), nil
		case ast.NOT_EQUALS:
			return ir.NotEquals[mir.LogicalTerm](lhs, rhs), nil
		case ast.LESS_THAN, ast.LESS_THAN_EQUALS, ast.GREATER_THAN, ast.GREATER_THAN_EQUALS:
			panic("inequality not currently supported")
		default:
			panic("unreachable")
		}
	case *ast.List:
		args, errs := t.translateLogicals(mod, shift, e.Args...)
		// Sanity check void
		if len(args) == 0 {
			return nil, errs
		}
		//
		return ir.Conjunction(args...), errs
	case *ast.Not:
		arg, errs := t.translateLogical(e.Arg, mod, shift)
		return ir.Negate(arg), errs
	default:
		typeStr := reflect.TypeOf(expr).String()
		msg := fmt.Sprintf("unknown expression encountered during translation (%s)", typeStr)
		//
		return nil, t.srcmap.SyntaxErrors(expr, msg)
	}
}

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateOptionalLogical(expr ast.Expr, module *ModuleBuilder,
	shift int) (mir.LogicalTerm, []SyntaxError) {
	//
	if expr != nil {
		return t.translateLogical(expr, module, shift)
	}

	return nil, nil
}

// Determine the underlying register for a symbol which represents a register access.
func (t *translator) registerOfRegisterAccess(symbol ast.Symbol, shift int) (mir.Term, []SyntaxError) {
	switch e := symbol.(type) {
	case *ast.ArrayAccess:
		return t.registerOfArrayAccess(e, shift)
	case *ast.VariableAccess:
		return t.registerOfVariableAccess(e, shift)
	}
	//
	return nil, t.srcmap.SyntaxErrors(symbol, "invalid register access")
}

func (t *translator) registerOfVariableAccess(expr *ast.VariableAccess, shift int) (mir.Term, []SyntaxError) {
	if binding, ok := expr.Binding().(*ast.ColumnBinding); ok {
		// Lookup register binding
		return t.registerOf(binding.AbsolutePath(), shift), nil
	}
	//
	return nil, t.srcmap.SyntaxErrors(expr, "invalid register access")
}
func (t *translator) registerOfArrayAccess(expr *ast.ArrayAccess, shift int) (mir.Term, []SyntaxError) {
	var (
		errors []SyntaxError
		min    uint = 0
		max    uint = math.MaxUint
	)
	// Lookup the register
	binding, ok := expr.Binding().(*ast.ColumnBinding)
	// Did we find it?
	if !ok {
		errors = append(errors, *t.srcmap.SyntaxError(expr.Arg, "invalid array index encountered during translation"))
	} else if arr_t, ok := binding.DataType.(*ast.ArrayType); ok {
		min = arr_t.MinIndex()
		max = arr_t.MaxIndex()
	}
	// Array index should be statically known
	index := expr.Arg.AsConstant()
	//
	if index == nil {
		errors = append(errors, *t.srcmap.SyntaxError(expr.Arg, "expected constant array index"))
	} else if i := uint(index.Uint64()); i < min || i > max {
		errors = append(errors, *t.srcmap.SyntaxError(expr.Arg, "array index out-of-bounds"))
	}
	// Error check
	if len(errors) > 0 {
		return nil, errors
	}
	// Construct real register name
	path := &binding.Path
	name := fmt.Sprintf("%s_%d", path.Tail(), index.Uint64())
	path = path.Parent().Extend(name)
	//
	return t.registerOf(path, shift), errors
}

// Map registers to appropriate module register identifiers.
func (t *translator) registerOf(path *util.Path, shift int) mir.Term {
	// Determine register id
	rid := t.env.RegisterOf(path)
	//
	reg := t.env.Register(rid)
	// Lookup corresponding module builder
	module := t.schema.Module(reg.Context.ModuleId)
	//
	return module.RegisterAccessOf(reg.Name(), shift)
}

func determineMaxBitwidth(module *ModuleBuilder, sources []mir.Term) uint {
	// Sanity check bitwidth
	bitwidth := uint(0)
	//
	for _, e := range sources {
		// Determine bitwidth of nth term
		switch e := e.(type) {
		case *ir.RegisterAccess[mir.Term]:
			reg := module.Register(e.Register)
			//
			if reg.Width > bitwidth {
				bitwidth = reg.Width
			}
		default:
			// For now, we only supports simple column accesses.
			panic("bitwidth calculation only supported for column accesses")
		}
	}
	//
	return bitwidth
}

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
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/ir/assignment"
	"github.com/consensys/go-corset/pkg/ir/mir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/collection/array"
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
	builder := ir.NewSchemaBuilder[mir.Constraint, mir.Term](externs...)
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

// Translate the given Corset module into a family of one (or more) MIR modules.
// Normally, every Corset module corresponds to exactly one MIR module. More
// specifically, there will be one module for each distinct length multiplier.
// Thus, in the presence of interleavings, a Corset module will map to more than
// one MIR module.
func (t *translator) translateModule(name string) {
	// Always include module with base multiplier (even if empty).
	t.schema.NewModule(name, 1)
	// Initialise the corresponding family of MIR modules.
	for _, regIndex := range t.env.RegistersOf(name) {
		var (
			// Identify register info
			regInfo = t.env.Register(regIndex)
			// Determine corresponding module name
			moduleName = regInfo.Context.ModuleName()
		)
		// Check whether module created this already (or not)
		if _, ok := t.schema.HasModule(moduleName); !ok {
			// No, therefore create new module.
			t.schema.NewModule(moduleName, regInfo.Context.LengthMultiplier())
		}
	}
	// Translate all corset registers in this module into MIR registers across
	// the corresponding *family* of modules.
	t.translateModuleRegisters(t.env.RegistersOf(name))
}

// Add all registers defined in the given Corset module into registers in one
// (or more) MIR modules.
func (t *translator) translateModuleRegisters(corsetRegisters []uint) {
	// Process each register in turn.
	for _, regIndex := range corsetRegisters {
		var (
			// Identify register info
			regInfo = t.env.Register(regIndex)
			// Identify enclosing MIR module
			module = t.schema.ModuleOf(regInfo.Context.ModuleName())
			//
			reg schema.Register
		)
		// Declare corresponding register
		if regInfo.IsInput() {
			reg = schema.NewInputRegister(regInfo.Name(), regInfo.Bitwidth)
		} else {
			reg = schema.NewComputedRegister(regInfo.Name(), regInfo.Bitwidth)
		}
		// Add the register
		module.NewRegister(reg)
		// Add range constraints for underlying types (as necessary)
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
		regWidth := reg.Bitwidth
		// For now, enforce all source registers have matching bitwidth.
		for _, col := range reg.Sources {
			// Determine bitwidth
			colWidth := col.Bitwidth
			// Sanity check (for now)
			if col.MustProve && colWidth != regWidth {
				// Currently, mixed-width proving types are not supported.
				panic("cannot (currently) prove type of mixed-width register")
			}
		}
		// Add appropriate type constraint
		constraint := mir.NewRangeConstraint(reg.Name(),
			mod.Id(),
			mod.RegisterAccessOf(reg.Name(), 0),
			reg.Bitwidth)
		//
		mod.AddConstraint(constraint)
	}
}

// Translate all assignment or constraint declarations in the circuit.
func (t *translator) translateDeclarations(circuit *ast.Circuit) []SyntaxError {
	rootPath := util.NewAbsolutePath()
	errors := t.translateDeclarationsInModule(rootPath, circuit.Declarations)
	// Translate each module
	for _, m := range circuit.Modules {
		modPath := rootPath.Extend(m.Name)
		errs := t.translateDeclarationsInModule(*modPath, m.Declarations)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate all assignment or constraint declarations in a given module within
// the circuit.
func (t *translator) translateDeclarationsInModule(path util.Path, decls []ast.Declaration) []SyntaxError {
	var errors []SyntaxError
	//
	for _, d := range decls {
		errs := t.translateDeclaration(d, path)
		errors = append(errors, errs...)
	}
	// Done
	return errors
}

// Translate an assignment or constraint declarartion which occurs within a
// given module.
func (t *translator) translateDeclaration(decl ast.Declaration, path util.Path) []SyntaxError {
	var errors []SyntaxError
	//
	switch d := decl.(type) {
	case *ast.DefAliases:
		// Not an assignment or a constraint, hence ignore.
	case *ast.DefComputed:
		return t.translateDefComputed(d, path)
	case *ast.DefColumns:
		// Not an assignment or a constraint, hence ignore.
	case *ast.DefConst:
		// For now, constants are always compiled out when going down to mir.
	case *ast.DefConstraint:
		errors = t.translateDefConstraint(d)
	case *ast.DefFun:
		// For now, functions are always compiled out when going down to mir.
		// In the future, this might change if we add support for macros to mir.
	case *ast.DefInRange:
		errors = t.translateDefInRange(d)
	case *ast.DefInterleaved:
		errors = t.translateDefInterleaved(d, path)
	case *ast.DefLookup:
		errors = t.translateDefLookup(d)
	case *ast.DefPermutation:
		t.translateDefPermutation(d, path)
	case *ast.DefPerspective:
		// As for defregisters, nothing generated here.
	case *ast.DefProperty:
		errors = t.translateDefProperty(d)
	case *ast.DefSorted:
		errors = t.translateDefSorted(d)
	default:
		// Error handling
		panic("unknown declaration")
	}
	//
	return errors
}

// Translate a "defcomputed" declaration.
func (t *translator) translateDefComputed(decl *ast.DefComputed, path util.Path) []SyntaxError {
	var context ast.Context = ast.VoidContext()
	//
	targets := make([]schema.RegisterRef, len(decl.Targets))
	sources := make([]schema.RegisterRef, len(decl.Sources))
	// Identify source registers
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i].Binding().(*ast.ColumnBinding)
		source := t.env.Register(t.env.RegisterOf(&ith.Path))
		sources[i] = t.registerRefOf(&ith.Path)
		// Join contexts
		context = context.Join(source.Context)
	}
	// Identify target registers
	for i := 0; i < len(decl.Targets); i++ {
		targetPath := path.Extend(decl.Targets[i].Name())
		target := t.env.Register(t.env.RegisterOf(targetPath))
		targets[i] = t.registerRefOf(targetPath)
		// Join contexts
		context = context.Join(target.Context)
	}
	// Extract the binding
	binding := decl.Function.Binding().(*NativeDefinition)
	// Sanity check
	if context.IsConflicted() || context.IsVoid() {
		return t.srcmap.SyntaxErrors(decl, "conflicting (or void) constraint context")
	}
	// Determine enclosing module
	module := t.moduleOf(context)
	// Add the assignment and check the first identifier.
	module.AddAssignment(assignment.NewComputation(binding.name, targets, sources))
	//
	return nil
}

// Translate a "defconstraint" declaration.
func (t *translator) translateDefConstraint(decl *ast.DefConstraint) []SyntaxError {
	var (
		module = t.moduleOf(decl.Constraint.Context())
		// Translate expr body
		expr, errors = t.translateLogical(decl.Constraint, module, 0)
	)
	// Apply guard
	if expr == nil {
		// NOTE: in this case, the constraint itself has been translated as nil.
		// This means there is no constraint (e.g. its a debug constraint, but
		// debug mode is not enabled).
		return errors
	}
	// Apply guard (if applicable)
	if decl.Guard != nil {
		// Translate (optional) guard
		gexpr, guardErrors := t.translateOptionalExpression(decl.Guard, module, 0)
		guard := ir.Equals[mir.LogicalTerm](gexpr, ir.Const64[mir.Term](0))
		expr = ir.IfThenElse(guard, nil, expr)
		// Combine errors
		errors = append(errors, guardErrors...)
	}
	// Apply perspective selector (if applicable)
	if decl.Perspective != nil {
		// Translate (optional) perspective selector
		sexpr, selectorErrors := t.translateSelectorInModule(decl.Perspective, module)
		selector := ir.Equals[mir.LogicalTerm](sexpr, ir.Const64[mir.Term](0))
		expr = ir.IfThenElse(selector, nil, expr)
		// Combine errors
		errors = append(errors, selectorErrors...)
	}
	// Sanity check
	if len(errors) == 0 {
		// Add translated constraint
		module.AddConstraint(mir.NewVanishingConstraint(decl.Handle, module.Id(), decl.Domain, expr))
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
func (t *translator) translateDefLookup(decl *ast.DefLookup) []SyntaxError {
	var (
		errors  []SyntaxError
		context ast.Context
		sources []ir.Enclosed[[]mir.Term]
		targets []ir.Enclosed[[]mir.Term]
	)
	// Translate sources
	for i, ith := range decl.Targets {
		ith_targets, _, errs := t.translateDefLookupSources(decl.TargetSelectors[i], ith)
		targets = append(targets, ith_targets)
		errors = append(errors, errs...)
	}
	// Translate targets
	for i, ith := range decl.Sources {
		ith_sources, ctx, errs := t.translateDefLookupSources(decl.SourceSelectors[i], ith)
		sources = append(sources, ith_sources)
		errors = append(errors, errs...)
		//
		if i == 0 {
			context = ctx
		}
	}
	// Sanity check whether we can construct the constraint, or not.
	if len(errors) == 0 {
		module := t.moduleOf(context)
		// Add translated constraint
		module.AddConstraint(mir.NewLookupConstraint(decl.Handle, targets, sources))
	}
	// Done
	return errors
}

func (t *translator) translateDefLookupSources(selector ast.Expr,
	sources []ast.Expr) (ir.Enclosed[[]mir.Term], ast.Context, []SyntaxError) {
	// Determine context of ith set of targets
	context, j := ast.ContextOfExpressions(sources...)
	// Include selector (when present)
	if selector != nil {
		context = context.Join(selector.Context())
	}
	// Translate target expressions whilst again checking for a conflicting
	// context.
	if context.IsConflicted() {
		return ir.Enclosed[[]mir.Term]{}, context, t.srcmap.SyntaxErrors(sources[j], "conflicting context")
	}
	// Determine enclosing module
	module := t.moduleOf(context)
	// Translate source expressions
	terms, errors := t.translateUnitExpressions(sources, module, 0)
	// handle selector
	if selector != nil {
		s, errs := t.translateExpression(selector, module, 0)
		errors = append(errors, errs...)
		terms = array.Prepend(s, terms)
	} else {
		// Selector is 1
		terms = array.Prepend(ir.Const64[mir.Term](1), terms)
	}
	// Return enclosed terms
	return ir.Enclose(module.Id(), terms), context, errors
}

// Translate a "definrange" declaration.
func (t *translator) translateDefInRange(decl *ast.DefInRange) []SyntaxError {
	module := t.moduleOf(decl.Expr.Context())
	// Translate constraint body
	expr, errors := t.translateExpression(decl.Expr, module, 0)
	//
	if len(errors) == 0 {
		// Add translated constraint
		module.AddConstraint(mir.NewRangeConstraint("", module.Id(), expr, decl.Bitwidth))
	}
	// Done
	return errors
}

// Translate a "definterleaved" declaration.
// nolint
func (t *translator) translateDefInterleaved(decl *ast.DefInterleaved, path util.Path) []SyntaxError {
	//
	var (
		errors []SyntaxError
		//
		sources = make([]schema.RegisterRef, len(decl.Sources))
		targets = make([]schema.RegisterRef, 1)
		//
		sourceContext ast.Context
		sourceTerms   = make([]mir.Term, len(decl.Sources))
		// Lookup target register info
		targetPath = path.Extend(decl.Target.Name())
		targetId   = t.env.RegisterOf(targetPath)
		target     = t.env.Register(targetId)
	)
	// Determine source context
	for _, source := range decl.Sources {
		sourceBinding := source.Binding().(*ast.ColumnBinding)
		sourceContext = sourceContext.Join(sourceBinding.Context())
	}
	// Determine enclosing tgtModule
	tgtModule := t.moduleOf(target.Context)
	srcModule := t.moduleOf(sourceContext)
	// Determine source register refs
	for i, source := range decl.Sources {
		ith, errs := t.registerOfRegisterAccess(source, 0)
		//
		if len(errs) == 0 {
			sources[i] = schema.NewRegisterRef(srcModule.Id(), ith.Register)
			sourceTerms[i] = ith
		}
		//
		errors = append(errors, errs...)
	}
	// Determine target register refs
	targets[0] = schema.NewRegisterRef(tgtModule.Id(), t.registerIndexOf(targetPath))
	targetTerm := t.registerOf(targetPath, 0)
	// Register constraint
	tgtModule.AddConstraint(
		mir.NewInterleavingConstraint("", tgtModule.Id(), srcModule.Id(), targetTerm, sourceTerms),
	)
	// Register assignment
	tgtModule.AddAssignment(
		assignment.NewComputation("interleave", targets, sources))

	// Done
	return errors
}

// Translate a "defpermutation" declaration.
func (t *translator) translateDefPermutation(decl *ast.DefPermutation, path util.Path) []SyntaxError {
	//
	var (
		context     ast.Context = ast.VoidContext()
		targets                 = make([]schema.RegisterId, len(decl.Sources))
		targetTerms             = make([]mir.Term, len(decl.Sources))
		sources                 = make([]schema.RegisterId, len(decl.Sources))
		handle      strings.Builder
	)
	//
	for i := range decl.Sources {
		targetPath := path.Extend(decl.Targets[i].Name())
		targets[i] = t.registerIndexOf(targetPath)
		targetTerms[i] = t.registerOf(targetPath, 0)
		//
		target := t.env.Register(t.env.RegisterOf(targetPath))
		sourceBinding := decl.Sources[i].Binding().(*ast.ColumnBinding)
		sources[i] = t.registerIndexOf(&sourceBinding.Path)
		// Join contexts
		context = context.Join(target.Context)
		// Construct handle
		if i >= len(decl.Signs) {
			// No nothing
		} else if decl.Signs[i] {
			handle.WriteString("+")
		} else {
			handle.WriteString("-")
		}
		//
		handle.WriteString(target.Name())
	}

	if context.IsConflicted() || context.IsVoid() {
		return t.srcmap.SyntaxErrors(decl, "conflicting (or void) constraint context")
	}
	//
	module := t.moduleOf(context)
	// Clone the signs
	signs := slices.Clone(decl.Signs)
	bitwidth := determineMaxBitwidth(module, targetTerms[:len(signs)])
	// Add assignment for computing the sorted permutation
	module.AddAssignment(assignment.NewSortedPermutation(module.Id(), targets, signs, sources))
	// Add Permutation Constraint
	module.AddConstraint(mir.NewPermutationConstraint(handle.String(), module.Id(), targets, sources))
	// Add Sorting Constraint
	module.AddConstraint(
		mir.NewSortedConstraint(handle.String(), module.Id(), bitwidth, util.None[mir.Term](), targetTerms, signs, false))
	//
	return nil
}

// Translate a "defproperty" declaration.
func (t *translator) translateDefProperty(decl *ast.DefProperty) []SyntaxError {
	module := t.moduleOf(decl.Assertion.Context())
	// Translate constraint body
	assertion, errors := t.translateLogical(decl.Assertion, module, 0)
	//
	if len(errors) == 0 {
		// Add translated constraint
		module.AddConstraint(mir.NewAssertion(decl.Handle, module.Id(), assertion))
	}
	// Done
	return errors
}

// Translate a "defsorted" declaration.
func (t *translator) translateDefSorted(decl *ast.DefSorted) []SyntaxError {
	var (
		selector util.Option[mir.Term]
		// Determine source context
		context, _ = ast.ContextOfExpressions(decl.Sources...)
		//
		module = t.moduleOf(context)
	)

	// Translate source expressions
	sources, errors := t.translateUnitExpressions(decl.Sources, module, 0)
	// Translate (optional) selector expression
	if decl.Selector.HasValue() {
		sel, errs := t.translateExpression(decl.Selector.Unwrap(), module, 0)
		selector = util.Some(sel)
		//
		errors = append(errors, errs...)
	}
	// Create construct (assuming no errors thus far)
	if len(errors) == 0 {
		// Clone the signs
		signs := slices.Clone(decl.Signs)
		bitwidth := determineMaxBitwidth(module, sources[:len(signs)])
		// Add translated constraint
		module.AddConstraint(
			mir.NewSortedConstraint(decl.Handle, module.Id(), bitwidth, selector, sources, signs, decl.Strict))
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

// Translate an optional expression in a given context.  That is an expression
// which maybe nil (i.e. doesn't exist).  In such case, nil is returned (i.e.
// without any errors).
func (t *translator) translateOptionalExpression(expr ast.Expr, module *ModuleBuilder,
	shift int) (mir.Term, []SyntaxError) {
	//
	if expr != nil {
		return t.translateExpression(expr, module, shift)
	}

	return nil, nil
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
		} else if intType, ok := e.Type.(*ast.IntType); ok {
			// unsafe casts cannot be checked by the type checker, but can be
			// exploited for the purposes of optimisation.
			return ir.CastOf(arg, intType.BitWidth()), errs
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
	case *ast.VectorAccess:
		return t.translateVectorAccess(e, shift)
	default:
		typeStr := reflect.TypeOf(expr).String()
		msg := fmt.Sprintf("unknown arithmetic expression encountered during translation (%s)", typeStr)
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
	// Translate condition as a logical
	cond, condErrs := t.translateLogical(expr.Condition, module, shift)
	// Translate optional true / false branches
	args, argErrs := t.translateExpressions(module, shift, expr.TrueBranch, expr.FalseBranch)
	//
	errs := append(condErrs, argErrs...)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Propagate emptiness (if applicable)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	// Construct appropriate if form
	return ir.IfElse(cond, args[0], args[1]), nil
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

func (t *translator) translateVectorAccess(expr *ast.VectorAccess, shift int) (mir.Term, []SyntaxError) {
	var (
		limbs  []*mir.RegisterAccess = make([]*mir.RegisterAccess, len(expr.Vars))
		errors []SyntaxError
	)
	//
	for i, v := range expr.Vars {
		var (
			ith, errs = t.translateVariableAccess(v, shift)
		)
		// Sanity check it was a real register access
		if ra, ok := ith.(*mir.RegisterAccess); ok {
			limbs[i] = ra
		} else if len(errs) == 0 {
			errors = append(errors, *t.srcmap.SyntaxError(v, "invalid register access"))
		}
		//
		errors = append(errors, errs...)
	}
	//
	return ir.NewVectorAccess(limbs), errors
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

// Translate an expression situated in a given context.  The context is
// necessary to resolve unqualified names (e.g. for register access, function
// invocations, etc).
func (t *translator) translateLogical(expr ast.Expr, mod *ModuleBuilder, shift int) (mir.LogicalTerm, []SyntaxError) {
	switch e := expr.(type) {
	case *ast.Cast:
		if e.Type != ast.BOOL_TYPE {
			// This should be unreachable, since type checking should have
			// caught this already.  However, potentially, issues with the
			// preprocessor could result in some weird scenario.
			panic("malformed logical expression")
		}
		// Just ignore
		return t.translateLogical(e.Arg, mod, shift)
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
		case ast.LESS_THAN:
			return ir.LessThan[mir.LogicalTerm](lhs, rhs), nil
		case ast.LESS_THAN_EQUALS:
			return ir.LessThanOrEquals[mir.LogicalTerm](lhs, rhs), nil
		case ast.GREATER_THAN:
			return ir.GreaterThan[mir.LogicalTerm](lhs, rhs), nil
		case ast.GREATER_THAN_EQUALS:
			return ir.GreaterThanOrEquals[mir.LogicalTerm](lhs, rhs), nil
		default:
			panic("unreachable")
		}
	case *ast.If:
		return t.translateIte(e, mod, shift)
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
		return ir.Negation(arg), errs
	case *ast.Shift:
		return t.translateLogicalShift(e, mod, shift)
	default:
		typeStr := reflect.TypeOf(expr).String()
		msg := fmt.Sprintf("unknown logical expression encountered during translation (%s)", typeStr)
		//
		return nil, t.srcmap.SyntaxErrors(expr, msg)
	}
}

func (t *translator) translateIte(expr *ast.If, module *ModuleBuilder, shift int) (mir.LogicalTerm, []SyntaxError) {
	// Translate condition as a logical
	cond, errs := t.translateLogical(expr.Condition, module, shift)
	// Translate optional true / false branches
	truebranch, trueErrs := t.translateOptionalLogical(expr.TrueBranch, module, shift)
	// Translate optional true / false branches
	falsebranch, falseErrs := t.translateOptionalLogical(expr.FalseBranch, module, shift)
	//
	errs = append(errs, trueErrs...)
	errs = append(errs, falseErrs...)
	//
	if len(errs) > 0 {
		return nil, errs
	}
	// Propagate emptiness (if applicable)
	if truebranch == nil && falsebranch == nil {
		return nil, nil
	}
	// Construct appropriate if form
	return ir.IfThenElse(cond, truebranch, falsebranch), nil
}

func (t *translator) translateLogicalShift(expr *ast.Shift, mod *ModuleBuilder,
	shift int) (mir.LogicalTerm, []SyntaxError) {
	//
	constant := expr.Shift.AsConstant()
	// Determine the shift constant
	if constant == nil {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "expected constant shift")
	} else if !constant.IsInt64() {
		return nil, t.srcmap.SyntaxErrors(expr.Shift, "constant shift too large")
	}
	// Now translate target expression with updated shift.
	return t.translateLogical(expr.Arg, mod, shift+int(constant.Int64()))
}

// Determine the underlying register for a symbol which represents a register access.
func (t *translator) registerOfRegisterAccess(symbol ast.Symbol, shift int) (*mir.RegisterAccess, []SyntaxError) {
	switch e := symbol.(type) {
	case *ast.ArrayAccess:
		return t.registerOfArrayAccess(e, shift)
	case *ast.VariableAccess:
		return t.registerOfVariableAccess(e, shift)
	}
	//
	return nil, t.srcmap.SyntaxErrors(symbol, "invalid register access")
}

func (t *translator) registerOfVariableAccess(expr *ast.VariableAccess,
	shift int) (*mir.RegisterAccess, []SyntaxError) {
	//
	if binding, ok := expr.Binding().(*ast.ColumnBinding); ok {
		// Lookup register binding
		return t.registerOf(binding.AbsolutePath(), shift), nil
	}
	//
	return nil, t.srcmap.SyntaxErrors(expr, "invalid register access")
}

func (t *translator) registerOfArrayAccess(expr *ast.ArrayAccess, shift int) (*mir.RegisterAccess, []SyntaxError) {
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
	} else if arrType, ok := binding.DataType.(*ast.ArrayType); ok {
		min = arrType.MinIndex()
		max = arrType.MaxIndex()
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

// Determine the appropriate name for a given module based on a module context.
func (t *translator) moduleOf(context ast.Context) *ModuleBuilder {
	if context.IsVoid() {
		// NOTE: the intuition behing the choice to return nil here is allow for
		// situations where there is no context (e.g. constant expressions,
		// etc).  As such, return nil is safe as, for such expressions, the
		// module should never be accessed during their translation.
		return nil
	}
	//
	return t.schema.ModuleOf(context.ModuleName())
}

// Map columns to appropriate module register identifiers.
func (t *translator) registerOf(path *util.Path, shift int) *mir.RegisterAccess {
	// Determine register id
	rid := t.env.RegisterOf(path)
	//
	reg := t.env.Register(rid)
	// Lookup corresponding module builder
	module := t.moduleOf(reg.Context)
	//
	return module.RegisterAccessOf(reg.Name(), shift)
}

// Map columns to appropriate module register identifiers.
func (t *translator) registerIndexOf(path *util.Path) schema.RegisterId {
	// Determine register id
	rid := t.env.RegisterOf(path)
	//
	reg := t.env.Register(rid)
	// Lookup corresponding module builder
	module := t.moduleOf(reg.Context)
	//
	if rid, ok := module.HasRegister(reg.Name()); ok {
		return rid
	}
	//
	panic("unreachable")
}

func (t *translator) registerRefOf(path *util.Path) schema.RegisterRef {
	// Determine register id
	rid := t.env.RegisterOf(path)
	//
	reg := t.env.Register(rid)
	// Lookup corresponding module builder
	module := t.moduleOf(reg.Context)
	//
	if rid, ok := module.HasRegister(reg.Name()); ok {
		return schema.NewRegisterRef(module.Id(), rid)
	}
	//
	panic("unreachable")
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

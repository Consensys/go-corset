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
	"math/big"

	"github.com/consensys/go-corset/pkg/corset/ast"
	"github.com/consensys/go-corset/pkg/util/collection/iter"
	"github.com/consensys/go-corset/pkg/util/source"
)

// DeclPredicate is a shorthand notation.
type DeclPredicate = iter.Predicate[ast.Declaration]

// ResolveCircuit resolves all symbols declared and used within a circuit,
// producing an environment which can subsequently be used to look up the
// relevant module or column identifiers.  This process can fail, of course, it
// a symbol (e.g. a column) is referred to which doesn't exist.  Likewise, if
// two modules or columns with identical names are declared in the same scope,
// etc.
func ResolveCircuit(srcmap *source.Maps[ast.Node], circuit *ast.Circuit) (*ModuleScope, []SyntaxError) {
	// Construct top-level scope
	scope := NewModuleScope(nil)
	// Define natives
	for _, i := range NATIVES {
		scope.Define(&i)
	}
	// Define intrinsics
	for _, i := range INTRINSICS {
		scope.Define(&i)
	}
	// Register modules
	for _, m := range circuit.Modules {
		scope.Declare(m.Name, nil)
	}
	// Construct resolver
	r := resolver{srcmap}
	// Initialise all columns
	if errs := r.initialiseDeclarations(scope, circuit); len(errs) > 0 {
		return nil, errs
	}
	// Finalise all columns / declarations
	if errs := r.resolveAssignments(scope, circuit); len(errs) > 0 {
		return nil, errs
	}
	//
	if errs := r.resolveConstraints(scope, circuit); len(errs) > 0 {
		return nil, errs
	}
	// Done
	return scope, nil
}

// Resolver packages up information necessary for resolving a circuit and
// checking that everything makes sense.
type resolver struct {
	// Source maps nodes in the circuit back to the spans in their original
	// source files.  This is needed when reporting syntax errors to generate
	// highlights of the relevant source line(s) in question.
	srcmap *source.Maps[ast.Node]
}

// Initialise all columns from their declaring constructs.
func (r *resolver) initialiseDeclarations(scope *ModuleScope, circuit *ast.Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.initialiseDeclarationsInModule(scope, circuit.Declarations)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.initialiseDeclarationsInModule(scope.Enter(m.Name), m.Declarations)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Initialise all declarations in the given module scope.  That means allocating
// all bindings into the scope, whilst also ensuring that we never have two
// bindings for the same symbol, etc.  The key is that, at this stage, all
// bindings are potentially "non-finalised".  That means they may be missing key
// information which is yet to be determined (e.g. information about types, or
// contexts, etc).
func (r *resolver) initialiseDeclarationsInModule(scope *ModuleScope, decls []ast.Declaration) []SyntaxError {
	errors := make([]SyntaxError, 0)
	// First, initialise any perspectives as submodules of the given scope.  Its
	// slightly frustrating that we have to do this separately, but the
	// non-lexical nature of perspectives forces our hand.
	for _, d := range decls {
		if def, ok := d.(*ast.DefPerspective); ok {
			// Attempt to declare the perspective.  Note, we don't need to check
			// whether or not this succeeds here as, if it fails, this will be
			// caught below.
			scope.Declare(def.Name(), def.Selector)
		}
	}
	// Second, initialise all symbol (e.g. column) definitions.
	for _, d := range decls {
		for iter := d.Definitions(); iter.HasNext(); {
			def := iter.Next()
			// Attempt to declare symbol
			if !scope.Define(def) {
				msg := fmt.Sprintf("symbol %s already declared", def.Path())
				err := r.srcmap.SyntaxError(def, msg)
				errors = append(errors, *err)
			}
		}
	}
	// Third, intialise aliases
	if errors := r.initialiseAliasesInModule(scope, decls); len(errors) > 0 {
		return errors
	}
	//
	return errors
}

// Initialise all alias declarations in the given module scope.  This means
// declaring them within the module scope, whilst also supporting aliases of
// aliases, etc.  Since the order of aliases is unspecified, this means we have
// to iterate the alias declarations until a fixed point is reached.  Once that
// is done, if there are any aliases left unallocated then they indicate errors.
func (r *resolver) initialiseAliasesInModule(scope *ModuleScope, decls []ast.Declaration) []SyntaxError {
	// Apply any aliases
	errors := make([]SyntaxError, 0)
	visited := make(map[string]ast.Declaration)
	changed := true
	// Iterate aliases to fixed point (i.e. until no new aliases discovered)
	for changed {
		changed = false
		// Look for all aliases
		for _, d := range decls {
			if a, ok := d.(*ast.DefAliases); ok {
				for i, alias := range a.Aliases {
					symbol := a.Symbols[i]
					if _, ok := visited[alias.Name]; !ok {
						// Attempt to make the alias
						if change := scope.Alias(alias.Name, symbol); change {
							visited[alias.Name] = d
							changed = true
						}
					}
				}
			}
		}
	}
	// Check for any aliases which remain incomplete
	for _, decl := range decls {
		if a, ok := decl.(*ast.DefAliases); ok {
			for i, alias := range a.Aliases {
				symbol := a.Symbols[i]
				// Check whether it already exists (or not)
				if d, ok := visited[alias.Name]; ok && d == decl {
					continue
				} else if scope.Binding(alias.Name, symbol.IsFunction()) != nil {
					err := r.srcmap.SyntaxError(alias, "symbol already exists")
					errors = append(errors, *err)
				} else {
					err := r.srcmap.SyntaxError(symbol, "unknown symbol")
					errors = append(errors, *err)
				}
			}
		}
	}
	// Done
	return errors
}

// Process all assignment column declarations.  These are more complex than for
// input columns, since there can be dependencies between them.  Thus, we cannot
// simply resolve them in one linear scan.
func (r *resolver) resolveAssignments(scope *ModuleScope, circuit *ast.Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.finaliseDeclarationsInModule(scope, circuit.Declarations, isAssigmentDeclaration)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.finaliseDeclarationsInModule(scope.Enter(m.Name), m.Declarations, isAssigmentDeclaration)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Process all remaining declarations, such as constraint declarations.  These
// are more complex than for input columns, since there can be dependencies
// between them.  Thus, we cannot simply resolve them in one linear scan.
func (r *resolver) resolveConstraints(scope *ModuleScope, circuit *ast.Circuit) []SyntaxError {
	// Input columns must be allocated before assignemts, since the hir.Schema
	// separates these out.
	errs := r.finaliseDeclarationsInModule(scope, circuit.Declarations, isNotAssigmentDeclaration)
	//
	for _, m := range circuit.Modules {
		// Process all declarations in the module
		merrs := r.finaliseDeclarationsInModule(scope.Enter(m.Name), m.Declarations, isNotAssigmentDeclaration)
		// Package up all errors
		errs = append(errs, merrs...)
	}
	//
	return errs
}

// Determines whether a given declaration is an "assignment" or not.
// Specifically, an assignment is a declaration which defines one or more
// computed (i.e. assigned) columns.
func isAssigmentDeclaration(decl ast.Declaration) bool {
	return decl.IsAssignment()
}

// Determines whether a given declaration is not an "assignment" (see above).
func isNotAssigmentDeclaration(decl ast.Declaration) bool {
	return !decl.IsAssignment()
}

// Finalise a subset of declarations in a given module.  This requires an
// iterative process as we cannot finalise an arbitrary declaration until all of
// its dependencies have been themselves finalised.  For example, a function
// which depends upon an interleaved column.  Until the interleaved column is
// finalised, its type won't be available and, hence, we cannot type the
// function.
func (r *resolver) finaliseDeclarationsInModule(scope *ModuleScope, decls []ast.Declaration,
	includes DeclPredicate) []SyntaxError {
	// Changed indicates whether or not a new assignment was finalised during a
	// given iteration.  This is important to know since, if the assignment is
	// not complete and we didn't finalise any more assignments --- then, we've
	// reached a fixed point where the final assignment is incomplete (i.e.
	// there is some error somewhere).
	changed := true
	// Complete tells us whether or not the assignment is complete.  The
	// assignment is not complete if there it at least one declaration which is
	// not yet finalised.
	complete := false
	// For an incomplete assignment, this identifies the last declaration that
	// could not be finalised (i.e. as an example so we have at least one for
	// error reporting).
	var (
		incomplete ast.Node = nil
		counter    uint     = 32
		errors     []SyntaxError
		// Failed indicates declarations which are already considered to have
		// failed.
		failed = make([]bool, len(decls))
	)
	//
	for changed && !complete && counter > 0 {
		changed = false
		complete = true
		//
		for i, decl := range decls {
			// Check whether included and already finalised
			if includes(decl) && !failed[i] && !decl.IsFinalised() {
				// No, so attempt to finalise
				ready, errs := r.declarationDependenciesAreFinalised(scope, decl)
				// Check what we found
				if errs != nil {
					errors = append(errors, errs...)
					failed[i] = true
				} else if ready {
					// Finalise declaration and handle errors
					errs := r.finaliseDeclaration(scope, decl)
					errors = append(errors, errs...)
					// Record that a new assignment is available.
					changed = changed || len(errs) == 0
					failed[i] = (len(errs) != 0)
				} else {
					// ast.Declaration not ready yet
					complete = false
					incomplete = decl
				}
			}
		}
		// Decrement counter
		counter--
	}
	// Check whether we actually finished the allocation.
	if len(errors) > 0 {
		return errors
	} else if counter == 0 {
		err := r.srcmap.SyntaxError(incomplete, "unable to complete resolution")
		return []SyntaxError{*err}
	} else if !complete {
		// No, we didn't.  So, something is wrong and we now have to figure out
		// what exactly.
		return r.determineFinalisationErrors(decls, includes)
	}
	// Done
	return nil
}

// Check that a given set of symbols have been finalised.  This is important,
// since we cannot finalise a declaration until all of its dependencies have
// themselves been finalised.
func (r *resolver) declarationDependenciesAreFinalised(scope *ModuleScope,
	decl ast.Declaration) (bool, []SyntaxError) {
	var (
		errors    []SyntaxError
		finalised bool = true
	)
	// DefConstraints require special handling because they can be associated
	// with a perspective.  Perspectives are challenging here because they are
	// effectively non-lexical scopes, which is not a good fit for the module
	// tree structure used.
	if dc, ok := decl.(*ast.DefConstraint); ok && dc.Perspective != nil {
		if dc.Perspective.IsResolved() || scope.Bind(dc.Perspective) {
			// Temporarily enter the perspective for the purposes of resolving
			// symbols within this declaration.
			scope = scope.Enter(dc.Perspective.Name())
		}
	}
	//
	for iter := decl.Dependencies(); iter.HasNext(); {
		symbol := iter.Next()
		// Attempt to resolve
		if !symbol.IsResolved() && !scope.Bind(symbol) {
			errors = append(errors, *r.srcmap.SyntaxError(symbol, "unknown symbol"))
			// not finalised yet
			finalised = false
		} else {
			// Check whether this declaration defines this symbol (because if it
			// does, we cannot expect it to be finalised yet :)
			selfdefinition := decl.Defines(symbol)
			// Check whether this symbol is already finalised.
			symbol_finalised := symbol.Binding().IsFinalised()
			// Final check
			if !selfdefinition && !symbol_finalised {
				// Ok, not ready for finalisation yet.
				finalised = false
			}
		}
	}
	//
	return finalised, errors
}

// For each included declaration, identify which dependencies are unresolved and
// report specific errors for them.
func (r *resolver) determineFinalisationErrors(decls []ast.Declaration, includes DeclPredicate) []SyntaxError {
	var errors []SyntaxError
	//
	for _, decl := range decls {
		// Look for an included, but unfinalised declaration
		if includes(decl) && !decl.IsFinalised() {
			for iter := decl.Dependencies(); iter.HasNext(); {
				symbol := iter.Next()
				// Check whether this dependency is a problem
				if !symbol.Binding().IsFinalised() {
					// Yes, so report error
					errors = append(errors, *r.srcmap.SyntaxError(symbol, "unresolved symbol"))
				}
			}
		}
	}
	//
	return errors
}

// Finalise a declaration.
func (r *resolver) finaliseDeclaration(scope *ModuleScope, decl ast.Declaration) []SyntaxError {
	switch d := decl.(type) {
	case *ast.DefComputed:
		return r.finaliseDefComputedInModule(d)
	case *ast.DefConst:
		return r.finaliseDefConstInModule(scope, d)
	case *ast.DefConstraint:
		return r.finaliseDefConstraintInModule(scope, d)
	case *ast.DefFun:
		return r.finaliseDefFunInModule(scope, d)
	case *ast.DefInRange:
		return r.finaliseDefInRangeInModule(scope, d)
	case *ast.DefInterleaved:
		return r.finaliseDefInterleavedInModule(d)
	case *ast.DefLookup:
		return r.finaliseDefLookupInModule(scope, d)
	case *ast.DefPermutation:
		return r.finaliseDefPermutationInModule(d)
	case *ast.DefPerspective:
		return r.finaliseDefPerspectiveInModule(scope, d)
	case *ast.DefProperty:
		return r.finaliseDefPropertyInModule(scope, d)
	case *ast.DefSorted:
		return r.finaliseDefSortedInModule(scope, d)
	}
	//
	return nil
}

func (r *resolver) finaliseDefComputedInModule(decl *ast.DefComputed) []SyntaxError {
	var (
		errors    []SyntaxError
		arguments []NativeColumn = make([]NativeColumn, len(decl.Sources))
		binding   *NativeDefinition
	)
	// Initialise arguments
	for i := 0; i < len(decl.Sources); i++ {
		// FIXME: sanity check that these things make sense.
		ith := decl.Sources[i].Binding().(*ast.ColumnBinding)
		arguments[i] = NativeColumn{ith.DataType, ith.Multiplier}
	}
	// Extract binding
	binding = decl.Function.Binding().(*NativeDefinition)
	//
	if !binding.HasArity(uint(len(arguments))) {
		msg := fmt.Sprintf("incorrect number of arguments (found %d)", len(arguments))
		errors = append(errors, *r.srcmap.SyntaxError(decl.Function, msg))
	} else {
		// Apply definition to determine geometry of assignment
		assignments := binding.Apply(arguments)
		//
		if len(assignments) > len(decl.Targets) {
			msg := fmt.Sprintf("not enough target columns (expected %d)", len(assignments))
			errors = append(errors, *r.srcmap.SyntaxError(decl.Function, msg))
		} else if len(assignments) < len(decl.Targets) {
			msg := fmt.Sprintf("too many target columns (expected %d)", len(assignments))
			errors = append(errors, *r.srcmap.SyntaxError(decl.Function, msg))
		} else {
			// Finalise each target column
			for i := 0; i < len(decl.Targets); i++ {
				// Finalise ith target column
				target := decl.Targets[i].Binding().(*ast.ColumnBinding)
				// Update with completed information
				target.Multiplier = assignments[i].multiplier
				target.DataType = assignments[i].datatype
			}
		}
	}
	// Done
	return errors
}

// Finalise one or more constant definitions within a given module.
// Specifically, we need to check that the constant values provided are indeed
// constants.
func (r *resolver) finaliseDefConstInModule(enclosing Scope, decl *ast.DefConst) []SyntaxError {
	var (
		errors []SyntaxError
		zero   = big.NewInt(0)
	)
	//
	for _, c := range decl.Constants {
		scope := NewLocalScope(enclosing, false, true, true)
		// Resolve constant body
		errs := r.finaliseExpressionInModule(scope, c.ConstBinding.Value)
		// Accumulate errors
		errors = append(errors, errs...)
		//
		if len(errs) == 0 {
			// Check it is indeed constant!
			if constant := c.ConstBinding.Value.AsConstant(); constant != nil {
				datatype := c.ConstBinding.DataType
				// Sanity check explicit type (if given)
				if datatype != nil && datatype.AsUnderlying().AsUint() != nil {
					uintType := datatype.AsUnderlying().AsUint()
					uintBound := uintType.IntBound()
					// bounds check
					if uintType != nil && constant.Cmp(&uintBound) >= 0 {
						// error, constant value outside bounds of given type!
						errors = append(errors, *r.srcmap.SyntaxError(c, "constant out-of-bounds (overflow)"))
						continue
					} else if uintType != nil && constant.Cmp(zero) < 0 {
						// unsigned integer cannot be negative.
						errors = append(errors, *r.srcmap.SyntaxError(c, "constant out-of-bounds (underflow)"))
						continue
					}
				}
				// Finalise constant binding.  Note, no need to register a syntax
				// error for the error case, because it would have already been
				// accounted for during resolution.
				c.ConstBinding.Finalise()
			}
		}
	}
	//
	return errors
}

// Finalise a vanishing constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefConstraintInModule(enclosing *ModuleScope, decl *ast.DefConstraint) []SyntaxError {
	var guard_errors []SyntaxError
	// Identifiery enclosing perspective (if applicable)
	if decl.Perspective != nil {
		// As before, we must temporarily enter the perspective here.
		perspective := decl.Perspective.Name()
		enclosing = enclosing.Enter(perspective)
	}
	// Construct scope in which to resolve constraint
	scope := NewLocalScope(enclosing, false, false, false)
	// Resolve guard
	if decl.Guard != nil {
		guard_errors = r.finaliseExpressionInModule(scope, decl.Guard)
	}
	// Resolve constraint body
	constraint_errors := r.finaliseExpressionInModule(scope, decl.Constraint)
	//
	if len(guard_errors) == 0 && len(constraint_errors) == 0 {
		// Finalise declaration.
		decl.Finalise()
	}
	// Done
	return append(guard_errors, constraint_errors...)
}

// Finalise an interleaving assignment.  Since the assignment would already been
// initialised, all we need to do is determine the appropriate type and length
// multiplier for the interleaved column.  This can still result in an error,
// for example, if the multipliers between interleaved columns are incompatible,
// etc.
func (r *resolver) finaliseDefInterleavedInModule(decl *ast.DefInterleaved) []SyntaxError {
	var (
		// Length multiplier being determined
		length_multiplier uint
		// Column type being determined
		datatype ast.Type
		// Errors discovered
		errors []SyntaxError
	)
	// Determine type and length multiplier
	for _, source := range decl.Sources {
		// Lookup binding of column being interleaved.
		if binding, ok := source.Binding().(*ast.ColumnBinding); !ok {
			// Columns to be interleaved must have the same length multiplier.
			err := r.srcmap.SyntaxError(source, "invalid source column")
			errors = append(errors, *err)
		} else if datatype == nil {
			length_multiplier = binding.Multiplier
			datatype = source.Type()
		} else if binding.Multiplier != length_multiplier {
			// Columns to be interleaved must have the same length multiplier.
			err := r.srcmap.SyntaxError(source, "incompatible length multiplier")
			errors = append(errors, *err)
		} else {
			// Combine datatypes.
			datatype = ast.GreatestLowerBound(datatype, source.Type())
		}
	}
	// Finalise details only if no errors
	if len(errors) == 0 {
		// Determine actual length multiplier
		length_multiplier *= uint(len(decl.Sources))
		// Lookup existing declaration
		binding := decl.Target.Binding().(*ast.ColumnBinding)
		// Finalise column binding
		binding.Finalise(length_multiplier, datatype)
	}
	// Done
	return errors
}

// Finalise a permutation assignment after all symbols have been resolved.  This
// requires checking the contexts of all columns is consistent.
func (r *resolver) finaliseDefPermutationInModule(decl *ast.DefPermutation) []SyntaxError {
	var (
		multiplier uint = 0
		errors     []SyntaxError
		started    bool
	)
	// Finalise each column in turn
	for i := 0; i < len(decl.Sources); i++ {
		ith := decl.Sources[i]
		// Lookup source of column being permuted
		if source, ok := ith.Binding().(*ast.ColumnBinding); !ok {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "invalid source column"))
			return errors
		} else if !started && source.DataType.AsUnderlying().AsUint() == nil {
			errors = append(errors, *r.srcmap.SyntaxError(ith, "fixed-width type required"))
		} else if started && multiplier != source.Multiplier {
			// Problem
			errors = append(errors, *r.srcmap.SyntaxError(ith, "incompatible length multiplier"))
		} else {
			// All good, finalise target column
			target := decl.Targets[i].Binding().(*ast.ColumnBinding)
			// Update with completed information
			target.Multiplier = source.Multiplier
			target.DataType = source.DataType
			multiplier = source.Multiplier
			started = true
		}
	}
	// Done
	return errors
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPerspectiveInModule(enclosing Scope, decl *ast.DefPerspective) []SyntaxError {
	scope := NewLocalScope(enclosing, false, false, false)
	// Resolve assertion
	errors := r.finaliseExpressionInModule(scope, decl.Selector)
	// Error check
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Finalise a range constraint declaration after all symbols have been
// resolved. This involves: (a) checking the context is valid; (b) checking the
// expressions are well-typed.
func (r *resolver) finaliseDefInRangeInModule(enclosing Scope, decl *ast.DefInRange) []SyntaxError {
	var scope = NewLocalScope(enclosing, false, false, false)
	// Resolve property body
	errors := r.finaliseExpressionInModule(scope, decl.Expr)
	// Error check
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Finalise a function definition after all symbols have been resolved. This
// involves: (a) checking the context is valid for the body; (b) checking the
// body is well-typed; (c) for pure functions checking that no columns are
// accessed; (d) finally, resolving any parameters used within the body of this
// function.
func (r *resolver) finaliseDefFunInModule(enclosing Scope, decl *ast.DefFun) []SyntaxError {
	var scope = NewLocalScope(enclosing, true, decl.IsPure(), false)
	// Declare parameters in local scope
	for _, p := range decl.Parameters() {
		scope.DeclareLocal(p.Binding.Name, &p.Binding)
	}
	// Resolve property body
	errors := r.finaliseExpressionInModule(scope, decl.Body())
	// Finalise declaration
	if len(errors) == 0 {
		decl.Finalise()
	}
	// Done
	return errors
}

// Resolve those variables appearing in the body of this lookup constraint.
func (r *resolver) finaliseDefLookupInModule(enclosing Scope, decl *ast.DefLookup) []SyntaxError {
	// Resolve source expressions
	sourceErrors := r.finaliseLookupVectorInModule(enclosing, decl.Source)
	// Resolve target expressions
	targetErrors := r.finaliseLookupVectorInModule(enclosing, decl.Target)
	//
	return append(sourceErrors, targetErrors...)
}

func (r *resolver) finaliseLookupVectorInModule(enclosing Scope, vec ast.LookupVector) []SyntaxError {
	var (
		selectorErrs []SyntaxError
		scope        = NewLocalScope(enclosing, true, false, false)
	)
	// Resolve selector (if applicable)
	if vec.Selector != nil {
		selectorErrs = r.finaliseExpressionInModule(scope, vec.Selector)
	}
	// Resolve terms
	sourceErrs := r.finaliseExpressionsInModule(scope, vec.Terms)
	//
	return append(selectorErrs, sourceErrs...)
}

// Resolve those variables appearing in the body of this property assertion.
func (r *resolver) finaliseDefPropertyInModule(enclosing Scope, decl *ast.DefProperty) []SyntaxError {
	scope := NewLocalScope(enclosing, false, false, false)
	// Resolve assertion
	return r.finaliseExpressionInModule(scope, decl.Assertion)
}

func (r *resolver) finaliseDefSortedInModule(enclosing Scope, decl *ast.DefSorted) []SyntaxError {
	var (
		scope = NewLocalScope(enclosing, false, false, false)
	)
	// Resolve source expressions
	errors := r.finaliseExpressionsInModule(scope, decl.Sources)
	// Resolve (optional) selector expression
	if decl.Selector.HasValue() {
		r.finaliseExpressionInModule(scope, decl.Selector.Unwrap())
	}
	// Sanity check length multipliers
	for _, e := range decl.Sources {
		// Sanity check multiplier has size 1
		if e.Context().Multiplier != 1 {
			errors = append(errors, *r.srcmap.SyntaxError(e, "interleaved column access not permitted"))
		}
	}
	// Error check
	if len(errors) == 0 {
		decl.Finalise()
	}
	//
	return errors
}

// Resolve a sequence of zero or more expressions within a given module.  This
// simply resolves each of the arguments in turn, collecting any errors arising.
func (r *resolver) finaliseExpressionsInModule(scope LocalScope, args []ast.Expr) []SyntaxError {
	var errors []SyntaxError
	// Visit each argument
	for _, arg := range args {
		if arg != nil {
			errs := r.finaliseExpressionInModule(scope, arg)
			errors = append(errors, errs...)
		}
	}
	// Done
	return errors
}

// Resolve any variable accesses with this expression (which is declared in a
// given module).  The enclosing module is required to resolve unqualified
// variable accesses.  As above, the goal is ensure variable refers to something
// that was declared and, more specifically, what kind of access it is (e.g.
// column access, constant access, etc).
//
//nolint:staticcheck
func (r *resolver) finaliseExpressionInModule(scope LocalScope, expr ast.Expr) []SyntaxError {
	switch v := expr.(type) {
	case *ast.ArrayAccess:
		return r.finaliseArrayAccessInModule(scope, v)
	case *ast.Add:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *ast.Cast:
		return r.finaliseExpressionInModule(scope, v.Arg)
	case *ast.Constant:
		return nil
	case *ast.Debug:
		return r.finaliseExpressionInModule(scope, v.Arg)
	case *ast.Exp:
		constscope := scope.NestedConstScope()
		arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		pow_errs := r.finaliseExpressionInModule(constscope, v.Pow)
		// combine errors
		return append(arg_errs, pow_errs...)
	case *ast.For:
		nestedscope := scope.NestedScope()
		// Declare local variable
		nestedscope.DeclareLocal(v.Binding.Name, &v.Binding)
		// Continue resolution
		return r.finaliseExpressionInModule(nestedscope, v.Body)
	case *ast.If:
		return r.finaliseExpressionsInModule(scope, []ast.Expr{v.Condition, v.TrueBranch, v.FalseBranch})
	case *ast.Invoke:
		return r.finaliseInvokeInModule(scope, v)
	case *ast.Let:
		return r.finaliseLetInModule(scope, v)
	case *ast.List:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *ast.Mul:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *ast.Normalise:
		return r.finaliseExpressionInModule(scope, v.Arg)
	case *ast.Reduce:
		return r.finaliseReduceInModule(scope, v)
	case *ast.Shift:
		constscope := scope.NestedConstScope()
		arg_errs := r.finaliseExpressionInModule(scope, v.Arg)
		shf_errs := r.finaliseExpressionInModule(constscope, v.Shift)
		// combine errors
		return append(arg_errs, shf_errs...)
	case *ast.Sub:
		return r.finaliseExpressionsInModule(scope, v.Args)
	case *ast.VariableAccess:
		return r.finaliseVariableInModule(scope, v)
	default:
		return r.srcmap.SyntaxErrors(expr, "unknown expression encountered during resolution")
	}
}

// Resolve a specific array access contained within some expression which, in
// turn, is contained within some module.
func (r *resolver) finaliseArrayAccessInModule(scope LocalScope, expr *ast.ArrayAccess) []SyntaxError {
	// Resolve argument
	errors := r.finaliseExpressionInModule(scope, expr.Arg)
	//
	if !expr.IsResolved() && !scope.Bind(expr) {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown array column"))
	} else if _, ok := expr.Binding().(*ast.ColumnBinding); !ok {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown array column"))
	}
	// All good
	return errors
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseInvokeInModule(scope LocalScope, expr *ast.Invoke) []SyntaxError {
	// Resolve arguments
	errors := r.finaliseExpressionsInModule(scope, expr.Args)
	// Lookup the corresponding function definition.
	if !expr.Name.IsResolved() && !scope.Bind(expr.Name) {
		return append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
	}
	// Following must be true if we get here.
	binding := expr.Name.Binding().(ast.FunctionBinding)
	// Check purity
	if scope.IsPure() && !binding.IsPure() {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "not permitted in pure context"))
	}
	// Check provide correct number of arguments
	if !binding.HasArity(uint(len(expr.Args))) {
		msg := fmt.Sprintf("incorrect number of arguments (found %d)", len(expr.Args))
		errors = append(errors, *r.srcmap.SyntaxError(expr, msg))
	}
	//
	return errors
}

func (r *resolver) finaliseLetInModule(scope LocalScope, expr *ast.Let) []SyntaxError {
	nestedscope := scope.NestedScope()
	// Declare assigned variable(s)
	for i, letvar := range expr.Vars {
		nestedscope.DeclareLocal(letvar.Name, &expr.Vars[i])
	}
	// Finalise assigned expressions
	args_errs := r.finaliseExpressionsInModule(scope, expr.Args)
	// Finalise body
	body_errs := r.finaliseExpressionInModule(nestedscope, expr.Body)
	//
	return append(args_errs, body_errs...)
}

// Resolve a specific invocation contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseReduceInModule(scope LocalScope, expr *ast.Reduce) []SyntaxError {
	// Resolve arguments
	errors := r.finaliseExpressionInModule(scope, expr.Arg)
	// Lookup the corresponding function definition.
	if !expr.Name.IsResolved() && !scope.Bind(expr.Name) {
		errors = append(errors, *r.srcmap.SyntaxError(expr, "unknown function"))
	} else {
		// Following must be true if we get here.
		binding := expr.Name.Binding().(ast.FunctionBinding)

		if scope.IsPure() && !binding.IsPure() {
			errors = append(errors, *r.srcmap.SyntaxError(expr, "not permitted in pure context"))
		}
	}
	// Done
	return errors
}

// Resolve a specific variable access contained within some expression which, in
// turn, is contained within some module.  Note, qualified accesses are only
// permitted in a global context.
func (r *resolver) finaliseVariableInModule(scope LocalScope, expr *ast.VariableAccess) []SyntaxError {
	// Check whether this is a qualified access, or not.
	if !scope.IsGlobal() && expr.Path().IsAbsolute() {
		return r.srcmap.SyntaxErrors(expr, "qualified access not permitted here")
	}
	// Symbol should be resolved at this point, but we'd better sanity check this.
	if !expr.IsResolved() && !scope.Bind(expr) {
		// Unable to resolve variable
		return r.srcmap.SyntaxErrors(expr, "unresolved symbol")
	}
	// Check what we've got.
	if binding, ok := expr.Binding().(*ast.ColumnBinding); ok {
		// For column bindings, we still need to sanity check the context is
		// compatible.
		if !scope.FixContext(binding.Context()) {
			return r.srcmap.SyntaxErrors(expr, "conflicting context")
		} else if scope.IsPure() {
			return r.srcmap.SyntaxErrors(expr, "not permitted in pure context")
		}
		//
		return nil
	} else if binding, ok := expr.Binding().(*ast.ConstantBinding); ok {
		// Constant
		if binding.Extern && scope.IsConstant() {
			return r.srcmap.SyntaxErrors(expr, "not permitted in const context")
		}
		//
		return nil
	} else if _, ok := expr.Binding().(*ast.LocalVariableBinding); ok {
		// Parameter, for or let variable
		return nil
	} else if _, ok := expr.Binding().(ast.FunctionBinding); ok {
		// Function doesn't makes sense here.
		return r.srcmap.SyntaxErrors(expr, "refers to a function")
	}
	// Should be unreachable.
	return r.srcmap.SyntaxErrors(expr, "unknown symbol kind")
}

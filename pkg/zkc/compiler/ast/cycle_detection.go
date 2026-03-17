package ast

import (
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// Dependencies returns the declaration indices that the given declaration
// depends on. Returns nil for declarations that do not participate in
// dependency graphs (e.g. functions, memories).
func Dependencies(d decl.Resolved, program Program) []uint {
	switch d := d.(type) {
	case *decl.ResolvedTypeAlias:
		return typeAliasDependencies(d, program)
	case *decl.ResolvedConstant:
		return constantDependencies(d)
	default:
		return nil
	}
}

func typeAliasDependencies(d *decl.ResolvedTypeAlias, _ Program) []uint {
	alias, ok := d.DataType.(*data.ResolvedAlias)
	if !ok {
		return nil
	}

	return []uint{alias.Name.Index}
}

func constantDependencies(d *decl.ResolvedConstant) []uint {
	uses := d.ConstExpr.ExternUses()
	if uses == nil {
		return nil
	}

	var indices []uint

	for it := uses.Iter(); it.HasNext(); {
		sym := it.Next()
		if sym.Kind == symbol.CONSTANT {
			indices = append(indices, sym.Index)
		}
	}

	return indices
}

// findCycle performs DFS from start. If a cycle is found, returns the set of
// indices involved in that cycle. Returns nil if no cycle.
func findCycle(start uint, program Program, processed map[uint]bool) map[uint]bool {

	decl := program.Components()[start]
	deps := Dependencies(decl, program)

	if len(deps) == 0 {
		return nil
	} else if processed[start] {
		return processed
	}

	processed[start] = true
	return findCycle(deps[0], program, processed)
}

// CycleDetection traverses the program and detects cyclic definitions in
// type aliases and constants. Returns syntax errors for any cycles found.
func CycleDetection(program Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		errors []source.SyntaxError
		// TODO use a set
		processed = make(map[uint]bool)
	)

	for i, decl := range program.Components() {
		if processed[uint(i)] {
			continue
		}

		if findCycle(uint(i), program, processed) != nil {
			errors = append(errors, srcmaps.SyntaxErrors(decl, "cyclic definition for "+decl.Name())...)
		}
	}

	return errors
}

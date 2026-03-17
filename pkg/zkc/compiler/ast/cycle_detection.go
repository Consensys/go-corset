package ast

import (
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/symbol"
)

// dependencies returns the declaration indices that the given declaration
// depends on. Returns nil for declarations that are not checked for cycles.
func dependencies(d decl.Resolved) []uint {
	switch d := d.(type) {
	case *decl.ResolvedTypeAlias:
		return typeAliasDependencies(d)
	case *decl.ResolvedConstant:
		return constantDependencies(d)
	default:
		return nil
	}
}

func typeAliasDependencies(d *decl.ResolvedTypeAlias) []uint {
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

// findCycle performs DFS from start. It returns the set of
// indices involved in that cycle if a cycle is found, else it returns nil.
func findCycle(start uint, program Program, visited map[uint]bool) map[uint]bool {
	d := program.Components()[start]
	deps := dependencies(d)

	if len(deps) == 0 {
		return nil
	} else if visited[start] {
		return visited
	}

	visited[start] = true

	return findCycle(deps[0], program, visited)
}

// CycleDetection traverses the program and detects cyclic definitions in
// type aliases and constants.
func CycleDetection(program Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		errors  []source.SyntaxError
		visited = make(map[uint]bool)
	)

	for i, d := range program.Components() {
		if visited[uint(i)] {
			continue
		}

		if findCycle(uint(i), program, visited) != nil {
			errors = append(errors, srcmaps.SyntaxErrors(d, "cyclic definition for "+d.Name())...)
		}
	}

	return errors
}

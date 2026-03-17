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
	symSet := d.ConstExpr.ExternUses()
	if symSet == nil {
		return nil
	}

	var deps []uint

	for sym := symSet.Iter(); sym.HasNext(); {
		s := sym.Next()
		if s.Kind == symbol.CONSTANT {
			deps = append(deps, s.Index)
		}
	}

	return deps
}

// findCycle performs DFS from start. It returns the set of indices involved in a
// cycle if one is found, else nil.
func findCycle(start uint, program Program, path, visited map[uint]bool) map[uint]bool {
	if visited[start] {
		return nil
	}

	if path[start] {
		for j := range path {
			visited[j] = true
		}

		return map[uint]bool{start: true}
	}

	d := program.Components()[start]
	deps := dependencies(d)

	if len(deps) == 0 {
		visited[start] = true
		return nil
	}

	path[start] = true

	for _, k := range deps {
		if findCycle(k, program, path, visited) != nil {
			for l := range path {
				visited[l] = true
			}

			path[start] = false

			return map[uint]bool{start: true}
		}
	}

	path[start] = false
	visited[start] = true

	return nil
}

// CycleDetection traverses the program and detects cyclic definitions in
// type constants and aliases.
func CycleDetection(program Program, srcmaps source.Maps[any]) []source.SyntaxError {
	var (
		errors  []source.SyntaxError
		visited = make(map[uint]bool)
	)

	for i, d := range program.Components() {
		if visited[uint(i)] {
			continue
		}

		path := make(map[uint]bool)
		if findCycle(uint(i), program, path, visited) != nil {
			errors = append(errors, srcmaps.SyntaxErrors(d, "cyclic definition for "+d.Name())...)
		}
	}

	return errors
}

package ast

import (
	"github.com/consensys/go-corset/pkg/util/source"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/data"
	"github.com/consensys/go-corset/pkg/zkc/compiler/ast/decl"
)

func CycleDetection(program Program, srcmaps source.Maps[any]) []source.SyntaxError {
	env := program.Environment()
	for _, d := range program.Components() {
		if decl, okTypeAlias := d.(*decl.ResolvedTypeAlias); okTypeAlias {
			aliasSet := map[string]struct{}{}
			datatype := decl.DataType

			alias, okAlias := datatype.(*data.ResolvedAlias)

			for okAlias {
				_, in := aliasSet[datatype.String(env)]

				if in {
					return srcmaps.SyntaxErrors(decl, "cyclic definition for "+decl.Name())
				}

				aliasSet[datatype.String(env)] = struct{}{}
				datatype = alias.Resolve(env)
				alias, okAlias = datatype.(*data.ResolvedAlias)
			}
		}
	}
	return nil
}
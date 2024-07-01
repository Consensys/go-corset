package json

import (
	"fmt"
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
)

// ToJsonString converts a trace into a JSON string.
func ToJsonString(tr trace.Trace) string {
	var builder strings.Builder

	columns := tr.Columns()
	//
	builder.WriteString("{")
	//
	for i := uint(0); i < columns.Len(); i++ {
		ith := columns.Get(i)
		mod := tr.Modules().Get(ith.Module())
		// Determine fully qualified column name
		name := ith.Name()
		// Prepend module name (if applicable)
		if mod.Name() != "" {
			name = fmt.Sprintf("%s.%s", mod.Name(), name)
		}
		//
		if i != 0 {
			builder.WriteString(", ")
		}
		//
		builder.WriteString("\"")
		builder.WriteString(name)
		builder.WriteString("\": [")

		for j := 0; j < int(ith.Height()); j++ {
			if j != 0 {
				builder.WriteString(", ")
			}

			builder.WriteString(ith.Get(j).String())
		}
		builder.WriteString("]")
	}
	//
	builder.WriteString("}")
	// Done
	return builder.String()
}

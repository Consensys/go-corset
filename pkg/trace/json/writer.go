package json

import (
	"strings"

	"github.com/consensys/go-corset/pkg/trace"
)

// ToJsonString converts a trace into a JSON string.
func ToJsonString(tr trace.Trace) string {
	var builder strings.Builder
	//
	builder.WriteString("{")
	//
	for i := uint(0); i < tr.Width(); i++ {
		if i != 0 {
			builder.WriteString(", ")
		}

		ith := tr.Column(i)

		builder.WriteString("\"")
		builder.WriteString(ith.Name())
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

package hir

// MacroDefinition represents something which can be called, and that will be
// inlined at the point of call.
type MacroDefinition struct {
	// Enclosing module
	module uint
	// Name of the macro
	name string
	// Parameters of the macro
	params []string
	// Body of the macro
	body Expr
	// Indicates whether or not this macro is "pure".  More specifically, pure
	// macros can only refer to parameters (i.e. cannot access enclosing columns
	// directly).
	pure bool
}

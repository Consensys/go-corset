package corset

import (
	"fmt"

	sc "github.com/consensys/go-corset/pkg/schema"
)

// Type embodies a richer notion of type found at the Corset level, compared
// with that found at lower levels (e.g. HIR). below.
type Type interface {
	// Determines whether or not this type supports "loobean" semantics.  If so,
	// this means that 0 is treated as true, with anything else being false.
	HasLoobeanSemantics() bool

	// Determines whether or not this type supports "boolean" semantics.  If so,
	// this means that 0 is treated as false, with anything else being true.
	HasBooleanSemantics() bool

	// Construct a variant of this type which employs loobean semantics.  This
	// will panic if the type has already been given boolean semantics.
	WithLoobeanSemantics() Type

	// Construct a variant of this type which employs boolean semantics.  This
	// will panic if the type has already been given loobean semantics.
	WithBooleanSemantics() Type

	// Access an underlying representation of this type (should one exist).  If
	// this doesn't exist, then nil is returned.
	AsUnderlying() sc.Type

	// Produce a string representation of this type.
	String() string
}

// NewFieldType constructs a native field type which, initially, has no semantic
// specified.
func NewFieldType() Type {
	return &NativeType{&sc.FieldType{}, false, false}
}

// NewUintType constructs a native uint type of the given width which,
// initially, has no semantic specified.
func NewUintType(nbits uint) Type {
	return &NativeType{sc.NewUintType(nbits), false, false}
}

// Join computes the Least Upper Bound of two types.  For example, the lub of
// u16 and u128 is u128, etc.  Observe that the type with no semantics is above
// those which have semantics.  Thus, joining a loobean with a boolean
// necessarily leads to a type without a semantic specifier.
func Join(lhs Type, rhs Type) Type {
	var (
		l_loobean bool = lhs.HasLoobeanSemantics()
		r_loobean bool = rhs.HasLoobeanSemantics()
		l_boolean bool = lhs.HasBooleanSemantics()
		r_boolean bool = rhs.HasBooleanSemantics()
	)
	// Determine join of underlying types
	underlying := sc.Join(lhs.AsUnderlying(), rhs.AsUnderlying())
	// Check whether semantics match or not.
	if l_loobean == r_loobean && l_boolean == r_boolean {
		return &NativeType{underlying, l_loobean, r_boolean}
	}
	//
	return &NativeType{underlying, false, false}
}

// NativeType simply wraps one of the types available at the HIR level (and below).
type NativeType struct {
	// The underlying type
	datatype sc.Type
	// Determines whether or not this type supports "loobean" semantics.  If so,
	// this means that 0 is treated as true, with anything else being false.
	loobean bool
	// Determines whether or not this type supports "boolean" semantics.  If so,
	// this means that 0 is treated as false, with anything else being true.
	boolean bool
}

// HasLoobeanSemantics indicates whether or not this type supports "loobean"
// semantics or not. If so, this means that 0 is treated as true, with anything
// else being false.
func (p *NativeType) HasLoobeanSemantics() bool {
	return p.loobean
}

// HasBooleanSemantics indicates whether or not this type supports "boolean"
// semantics. If so, this means that 0 is treated as false, with anything else
// being true.
func (p *NativeType) HasBooleanSemantics() bool {
	return p.boolean
}

// WithLoobeanSemantics constructs a variant of this type which employs loobean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *NativeType) WithLoobeanSemantics() Type {
	if p.HasBooleanSemantics() {
		panic("type already given boolean semantics")
	}
	// Done
	return &NativeType{p.datatype, true, false}
}

// WithBooleanSemantics constructs a variant of this type which employs boolean
// semantics.  This will panic if the type has already been given boolean
// semantics.
func (p *NativeType) WithBooleanSemantics() Type {
	if p.HasLoobeanSemantics() {
		panic("type already given loobean semantics")
	}
	// Done
	return &NativeType{p.datatype, false, true}
}

// AsUnderlying attempts to convert this type into an underlying type.  If this
// is not possible, then nil is returned.
func (p *NativeType) AsUnderlying() sc.Type {
	return p.datatype
}

func (p *NativeType) String() string {
	if p.loobean {
		return fmt.Sprintf("%s@loob", p.datatype.String())
	}
	//
	return p.datatype.String()
}

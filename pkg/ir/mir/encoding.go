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
package mir

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
)

const (
	// Constraints
	assertionTag       = byte(0)
	interleavingTag    = byte(1)
	lookupTag          = byte(2)
	permutationTag     = byte(3)
	rangeTag           = byte(4)
	sortedTag          = byte(5)
	sortedSelectionTag = byte(6)
	vanishingTag       = byte(7)
	// Logicals
	conjunctTag   = byte(10)
	disjunctTag   = byte(11)
	equalTag      = byte(12)
	notEqualTag   = byte(13)
	lessThanTag   = byte(14)
	lessThanEqTag = byte(15)
	negationTag   = byte(16)
	iteTagTF      = byte(17)
	iteTagT       = byte(18)
	iteTagF       = byte(19)
	// Expressions
	addTag              = byte(30)
	castTag             = byte(31)
	constantTag         = byte(32)
	ifZeroTag           = byte(33)
	labelledConstantTag = byte(34)
	registerAccessTag   = byte(35)
	// NOTE: unused registers only required for optional lookup selectors.
	unusedRegisterAccessTag = byte(36)
	expTag                  = byte(37)
	mulTag                  = byte(38)
	normTag                 = byte(39)
	subTag                  = byte(40)
	vectorAccessTag         = byte(41)
)

func encode_constraint(constraint schema.Constraint) ([]byte, error) {
	switch c := constraint.(type) {
	case Assertion:
		return encode_assertion(c)
	case InterleavingConstraint:
		return encode_interleaving(c)
	case LookupConstraint:
		return encode_lookup(c)
	case PermutationConstraint:
		return encode_permutation(c)
	case SortedConstraint:
		return encode_sorted(c)
	case RangeConstraint:
		return encode_range(c)
	case VanishingConstraint:
		return encode_vanishing(c)
	default:
		return nil, errors.New("unknown constraint")
	}
}

func encode_assertion(c Assertion) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Tag
	if _, err := buffer.Write([]byte{assertionTag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Context
	if err := gobEncoder.Encode(c.Context); err != nil {
		return nil, err
	}
	// Constraint
	err := encode_logical(c.Property, &buffer)
	// Done
	return buffer.Bytes(), err
}

func encode_interleaving(c InterleavingConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Tag
	if _, err := buffer.Write([]byte{interleavingTag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Target Context
	if err := gobEncoder.Encode(c.TargetContext); err != nil {
		return nil, err
	}
	// Target term
	if err := encode_term(c.Target, &buffer); err != nil {
		return nil, err
	}
	// Source Context
	if err := gobEncoder.Encode(c.SourceContext); err != nil {
		return nil, err
	}
	// Source terms
	if err := encode_nary(encode_term, &buffer, c.Sources); err != nil {
		return nil, err
	}
	//
	return buffer.Bytes(), nil
}

func encode_lookup(c LookupConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Tag
	if _, err := buffer.Write([]byte{lookupTag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Target terms
	if err := encode_nary(encode_lookup_vector, &buffer, c.Targets); err != nil {
		return nil, err
	}
	// Sources
	if err := encode_nary(encode_lookup_vector, &buffer, c.Sources); err != nil {
		return nil, err
	}
	//
	return buffer.Bytes(), nil
}

func encode_lookup_vector(terms lookup.Vector[Term], buffer *bytes.Buffer) error {
	var (
		gobEncoder = gob.NewEncoder(buffer)
	)
	// Source Context
	if err := gobEncoder.Encode(terms.Module); err != nil {
		return err
	}
	//
	if terms.HasSelector() {
		panic("todo")
	}
	// Source terms
	return encode_nary(encode_term, buffer, terms.Terms)
}

func encode_permutation(c PermutationConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Tag
	if _, err := buffer.Write([]byte{permutationTag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Column Context
	if err := gobEncoder.Encode(c.Context); err != nil {
		return nil, err
	}
	// Target terms
	if err := gobEncoder.Encode(c.Targets); err != nil {
		return nil, err
	}
	// Source terms
	if err := gobEncoder.Encode(c.Sources); err != nil {
		return nil, err
	}
	//
	return buffer.Bytes(), nil
}

func encode_sorted(c SortedConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
		tag        byte
	)
	//
	if c.Selector.HasValue() {
		tag = sortedSelectionTag
	} else {
		tag = sortedTag
	}
	// Tag
	if _, err := buffer.Write([]byte{tag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Context
	if err := gobEncoder.Encode(c.Context); err != nil {
		return nil, err
	}
	// Bitwidth
	if err := gobEncoder.Encode(c.BitWidth); err != nil {
		return nil, err
	}
	// Signs
	if err := gobEncoder.Encode(c.Signs); err != nil {
		return nil, err
	}
	// Strict
	if err := gobEncoder.Encode(c.Strict); err != nil {
		return nil, err
	}
	// Optional Selector
	if c.Selector.HasValue() {
		// Constraint
		if err := encode_term(c.Selector.Unwrap(), &buffer); err != nil {
			return nil, err
		}
	}
	// Sources
	err := encode_nary(encode_term, &buffer, c.Sources)
	//
	return buffer.Bytes(), err
}

func encode_vanishing(c VanishingConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	// Tag
	if _, err := buffer.Write([]byte{vanishingTag}); err != nil {
		return nil, err
	}
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Context
	if err := gobEncoder.Encode(c.Context); err != nil {
		return nil, err
	}
	// Domain
	if err := gobEncoder.Encode(&c.Domain); err != nil {
		return nil, err
	}
	// Constraint
	err := encode_logical(c.Constraint, &buffer)
	// Done
	return buffer.Bytes(), err
}

func encode_range(c RangeConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	//
	buffer.Write([]byte{rangeTag})
	// Handle
	if err := gobEncoder.Encode(c.Handle); err != nil {
		return nil, err
	}
	// Context
	if err := gobEncoder.Encode(c.Context); err != nil {
		return nil, err
	}
	// Bitwidth
	if err := gobEncoder.Encode(c.Bitwidth); err != nil {
		return nil, err
	}
	// Expression
	err := encode_term(c.Expr, &buffer)
	// Done
	return buffer.Bytes(), err
}

func decode_constraint(bytes []byte) (schema.Constraint, error) {
	switch bytes[0] {
	case assertionTag:
		return decode_assertion(bytes[1:])
	case interleavingTag:
		return decode_interleaving(bytes[1:])
	case lookupTag:
		return decode_lookup(bytes[1:])
	case permutationTag:
		return decode_permutation(bytes[1:])
	case rangeTag:
		return decode_range(bytes[1:])
	case sortedTag:
		return decode_sorted(false, bytes[1:])
	case sortedSelectionTag:
		return decode_sorted(true, bytes[1:])
	case vanishingTag:
		return decode_vanishing(bytes[1:])
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", bytes[0])
	}
}

func decode_assertion(data []byte) (schema.Constraint, error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		assertion  Assertion
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&assertion.Handle); err != nil {
		return assertion, err
	}
	// Context
	if err = gobDecoder.Decode(&assertion.Context); err != nil {
		return assertion, err
	}
	//
	assertion.Property, err = decode_logical(buffer)
	// Success!
	return assertion, err
}

func decode_interleaving(data []byte) (schema.Constraint, error) {
	var (
		buffer       = bytes.NewBuffer(data)
		gobDecoder   = gob.NewDecoder(buffer)
		interleaving InterleavingConstraint
		err          error
	)
	// Handle
	if err = gobDecoder.Decode(&interleaving.Handle); err != nil {
		return interleaving, err
	}
	// Target Context
	if err = gobDecoder.Decode(&interleaving.TargetContext); err != nil {
		return interleaving, err
	}
	// Targets
	if interleaving.Target, err = decode_term(buffer); err != nil {
		return interleaving, err
	}
	// Source Context
	if err = gobDecoder.Decode(&interleaving.SourceContext); err != nil {
		return interleaving, err
	}
	// Sources
	if interleaving.Sources, err = decode_nary(decode_term, buffer); err != nil {
		return interleaving, err
	}
	//
	return interleaving, nil
}

func decode_lookup(data []byte) (schema.Constraint, error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		lookup     LookupConstraint
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&lookup.Handle); err != nil {
		return lookup, err
	}
	// Targets
	if lookup.Targets, err = decode_nary(decode_lookup_vector, buffer); err != nil {
		return lookup, err
	}
	// Sources
	if lookup.Sources, err = decode_nary(decode_lookup_vector, buffer); err != nil {
		return lookup, err
	}
	//
	return lookup, nil
}

func decode_lookup_vector(buf *bytes.Buffer) (lookup.Vector[Term], error) {
	var (
		gobDecoder = gob.NewDecoder(buf)
		terms      lookup.Vector[Term]
		err        error
	)
	// Context
	if err = gobDecoder.Decode(&terms.Module); err != nil {
		return terms, err
	}
	// Contents
	if terms.Terms, err = decode_nary(decode_term, buf); err != nil {
		return terms, err
	}
	//
	//return terms, nil
	panic("todo: handle selectors")
}

func decode_permutation(data []byte) (schema.Constraint, error) {
	var (
		buffer      = bytes.NewBuffer(data)
		gobDecoder  = gob.NewDecoder(buffer)
		permutation PermutationConstraint
		err         error
	)
	// Handle
	if err = gobDecoder.Decode(&permutation.Handle); err != nil {
		return permutation, err
	}
	// Column Context
	if err = gobDecoder.Decode(&permutation.Context); err != nil {
		return permutation, err
	}
	// Target terms
	if err = gobDecoder.Decode(&permutation.Targets); err != nil {
		return permutation, err
	}
	// Source terms
	if err = gobDecoder.Decode(&permutation.Sources); err != nil {
		return permutation, err
	}
	//
	return permutation, nil
}

func decode_sorted(selector bool, data []byte) (schema.Constraint, error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		sorted     SortedConstraint
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&sorted.Handle); err != nil {
		return nil, err
	}
	// Context
	if err = gobDecoder.Decode(&sorted.Context); err != nil {
		return nil, err
	}
	// Bitwidth
	if err := gobDecoder.Decode(&sorted.BitWidth); err != nil {
		return nil, err
	}
	// Signs
	if err := gobDecoder.Decode(&sorted.Signs); err != nil {
		return nil, err
	}
	// Strict
	if err := gobDecoder.Decode(&sorted.Strict); err != nil {
		return nil, err
	}
	// Optional Selector
	if selector {
		var term Term
		//
		if term, err = decode_term(buffer); err != nil {
			return nil, err
		}
		//
		sorted.Selector = util.Some(term)
	}
	// Sources
	sorted.Sources, err = decode_nary(decode_term, buffer)
	// Done
	return sorted, err
}

func decode_range(data []byte) (schema.Constraint, error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		constraint RangeConstraint
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&constraint.Handle); err != nil {
		return constraint, err
	}
	// Context
	if err = gobDecoder.Decode(&constraint.Context); err != nil {
		return constraint, err
	}
	// Bitwidth
	if err = gobDecoder.Decode(&constraint.Bitwidth); err != nil {
		return constraint, err
	}
	//
	constraint.Expr, err = decode_term(buffer)
	// Success!
	return constraint, err
}

func decode_vanishing(data []byte) (schema.Constraint, error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		vanishing  VanishingConstraint
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&vanishing.Handle); err != nil {
		return vanishing, err
	}
	// Context
	if err = gobDecoder.Decode(&vanishing.Context); err != nil {
		return vanishing, err
	}
	// Domain
	if err = gobDecoder.Decode(&vanishing.Domain); err != nil {
		return vanishing, err
	}
	//
	vanishing.Constraint, err = decode_logical(buffer)
	// Success!
	return vanishing, err
}

// ============================================================================
// Logical Terms (encoding)
// ============================================================================

func encode_logical(term LogicalTerm, buf *bytes.Buffer) error {
	switch t := term.(type) {
	case *Conjunct:
		return encode_tagged_nary_logicals(conjunctTag, buf, t.Args...)
	case *Disjunct:
		return encode_tagged_nary_logicals(disjunctTag, buf, t.Args...)
	case *Equal:
		return encode_tagged_terms(equalTag, buf, t.Lhs, t.Rhs)
	case *Ite:
		return encode_ite(t, buf)
	case *Negate:
		return encode_tagged_logicals(negationTag, buf, t.Arg)
	case *NotEqual:
		return encode_tagged_terms(notEqualTag, buf, t.Lhs, t.Rhs)
	case *Inequality:
		if t.Strict {
			return encode_tagged_terms(lessThanTag, buf, t.Lhs, t.Rhs)
		}
		//
		return encode_tagged_terms(lessThanEqTag, buf, t.Lhs, t.Rhs)
	default:
		return fmt.Errorf("unknown logical term encountered (%s)", term.Lisp(nil).String(false))
	}
}

func encode_tagged_nary_logicals(tag byte, buf *bytes.Buffer, terms ...LogicalTerm) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_nary(encode_logical, buf, terms)
}

func encode_tagged_logicals(tag byte, buf *bytes.Buffer, terms ...LogicalTerm) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_n(encode_logical, buf, terms...)
}

func encode_ite(term *Ite, buf *bytes.Buffer) error {
	switch {
	case term.FalseBranch != nil && term.TrueBranch != nil:
		return encode_tagged_logicals(iteTagTF, buf, term.Condition, term.TrueBranch, term.FalseBranch)
	case term.FalseBranch == nil:
		return encode_tagged_logicals(iteTagT, buf, term.Condition, term.TrueBranch)
	case term.TrueBranch == nil:
		return encode_tagged_logicals(iteTagF, buf, term.Condition, term.FalseBranch)
	default:
		panic("unreachable")
	}
}

// ============================================================================
// Logical Terms (decoding)
// ============================================================================

func decode_logical(buf *bytes.Buffer) (LogicalTerm, error) {
	tag, err := buf.ReadByte()
	//
	if err != nil {
		return nil, err
	}
	//
	switch tag {
	case conjunctTag:
		return decode_nary_logicals(conjunctionConstructor, buf)
	case disjunctTag:
		return decode_nary_logicals(disjunctionConstructor, buf)
	case equalTag:
		return decode_terms(2, equalConstructor, buf)
	case iteTagTF, iteTagT, iteTagF:
		return decode_ite(tag, buf)
	case negationTag:
		return decode_logicals(1, negationConstructor, buf)
	case notEqualTag:
		return decode_terms(2, notEqualConstructor, buf)
	case lessThanTag:
		return decode_terms(2, lessThanConstructor, buf)
	case lessThanEqTag:
		return decode_terms(2, lessThanOrEqualsConstructor, buf)
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", tag)
	}
}

// Decode a variable number of terms, as determined by the leading byte.
func decode_nary_logicals(constructor func([]LogicalTerm) LogicalTerm, buf *bytes.Buffer) (LogicalTerm, error) {
	terms, err := decode_nary(decode_logical, buf)
	return constructor(terms), err
}

// Decode exactly n logicals terms
func decode_logicals[S any](n uint, constructor func([]LogicalTerm) S, buf *bytes.Buffer) (S, error) {
	terms, err := decode_n(n, decode_logical, buf)
	return constructor(terms), err
}

func decode_ite(tag byte, buf *bytes.Buffer) (LogicalTerm, error) {
	//
	switch tag {
	case iteTagTF:
		return decode_logicals(3, iteTrueFalseConstructor, buf)
	case iteTagT:
		return decode_logicals(2, iteTrueConstructor, buf)
	case iteTagF:
		return decode_logicals(2, iteFalseConstructor, buf)
	default:
		panic("unreachable")
	}
}

// ============================================================================
// Arithmetic Terms (encoding)
// ============================================================================

func encode_term(term Term, buf *bytes.Buffer) error {
	switch t := term.(type) {
	case *Add:
		return encode_tagged_nary_terms(addTag, buf, t.Args...)
	case *Cast:
		return encode_cast(*t, buf)
	case *Constant:
		return encode_constant(*t, buf)
	case *Exp:
		return encode_exponent(*t, buf)
	case *IfZero:
		return encode_ifZero(*t, buf)
	case *LabelledConst:
		return encode_labelled_constant(*t, buf)
	case *Mul:
		return encode_tagged_nary_terms(mulTag, buf, t.Args...)
	case *Norm:
		return encode_tagged_terms(normTag, buf, t.Arg)
	case *RegisterAccess:
		return encode_reg_access(*t, buf)
	case *Sub:
		return encode_tagged_nary_terms(subTag, buf, t.Args...)
	case *VectorAccess:
		return encode_vec_access(*t, buf)
	default:
		return fmt.Errorf("unknown arithmetic term encountered (%s)", term.Lisp(nil).String(false))
	}
}

func encode_tagged_nary_terms(tag byte, buf *bytes.Buffer, terms ...Term) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_nary(encode_term, buf, terms)
}

func encode_tagged_terms(tag byte, buf *bytes.Buffer, terms ...Term) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_n(encode_term, buf, terms...)
}

func encode_cast(term Cast, buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(castTag); err != nil {
		return err
	}
	// Bitwidth
	if err := binary.Write(buf, binary.BigEndian, uint16(term.BitWidth)); err != nil {
		return err
	}
	// term
	return encode_term(term.Arg, buf)
}

func encode_constant(term Constant, buf *bytes.Buffer) error {
	bytes := term.Value.Bytes()
	// Write tag
	if err := buf.WriteByte(constantTag); err != nil {
		return err
	}
	// Write value as 32bytes
	_, err := buf.Write(bytes[:])
	//
	return err
}

func encode_ifZero(term IfZero, buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(ifZeroTag); err != nil {
		return err
	}
	// Write condition
	if err := encode_logical(term.Condition, buf); err != nil {
		return err
	}
	// Write true + false branches
	return encode_n(encode_term, buf, term.TrueBranch, term.FalseBranch)
}

func encode_labelled_constant(term LabelledConst, buf *bytes.Buffer) error {
	var (
		str_bytes   = []byte(term.Label)
		str_len     = uint16(len(str_bytes))
		const_bytes = term.Value.Bytes()
	)
	// Write tag
	if err := buf.WriteByte(labelledConstantTag); err != nil {
		return err
	}
	// Write label length
	if err := binary.Write(buf, binary.BigEndian, str_len); err != nil {
		return err
	}
	// Write label contents
	if n, err := buf.Write(str_bytes); err != nil {
		return err
	} else if n != len(str_bytes) {
		return fmt.Errorf("failed encoding constant label (%d versus %d bytes)", n, len(str_bytes))
	}
	// Write value as 32bytes
	_, err := buf.Write(const_bytes[:])
	//
	return err
}

func encode_exponent(term Exp, buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(expTag); err != nil {
		return err
	}
	// Exponent
	if err := binary.Write(buf, binary.BigEndian, term.Pow); err != nil {
		return err
	}
	// term
	return encode_term(term.Arg, buf)
}

func encode_reg_access(term RegisterAccess, buf *bytes.Buffer) error {
	// Write (appropriate) tag
	if !term.Register.IsUsed() {
		return buf.WriteByte(unusedRegisterAccessTag)
	} else if err := buf.WriteByte(registerAccessTag); err != nil {
		return err
	}
	//
	return encode_raw_access(&term, buf)
}

func encode_vec_access(term VectorAccess, buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(vectorAccessTag); err != nil {
		return err
	}
	//
	return encode_nary(encode_raw_access, buf, term.Vars)
}

func encode_raw_access(term *RegisterAccess, buf *bytes.Buffer) error {
	// Register Index
	if err := binary.Write(buf, binary.BigEndian, uint16(term.Register.Unwrap())); err != nil {
		return err
	}
	// Shift
	if err := binary.Write(buf, binary.BigEndian, int16(term.Shift)); err != nil {
		return err
	}
	//
	return nil
}

// ============================================================================
// Arithmetic Terms (decoding)
// ============================================================================

// Decode an arbitrary term from the buffer.
func decode_term(buf *bytes.Buffer) (Term, error) {
	tag, err := buf.ReadByte()
	//
	if err != nil {
		return nil, err
	}
	//
	switch tag {
	case addTag:
		return decode_nary_terms(addConstructor, buf)
	case castTag:
		return decode_cast(buf)
	case constantTag:
		return decode_constant(buf)
	case expTag:
		return decode_exponent(buf)
	case ifZeroTag:
		return decode_ifzero(buf)
	case labelledConstantTag:
		return decode_labelled_constant(buf)
	case registerAccessTag:
		return decode_reg_access(buf)
	case unusedRegisterAccessTag:
		rid := schema.NewUnusedRegisterId()
		return ir.NewRegisterAccess[Term](rid, 0), nil
	case mulTag:
		return decode_nary_terms(mulConstructor, buf)
	case normTag:
		return decode_terms(1, normConstructor, buf)
	case subTag:
		return decode_nary_terms(subConstructor, buf)
	case vectorAccessTag:
		return decode_vec_access(buf)
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", tag)
	}
}

// Decode a variable number of terms, as determined by the leading byte.
func decode_nary_terms(constructor func([]Term) Term, buf *bytes.Buffer) (Term, error) {
	terms, err := decode_nary(decode_term, buf)
	return constructor(terms), err
}

// Decode exactly n terms
func decode_terms[S any](n uint, constructor func([]Term) S, buf *bytes.Buffer) (S, error) {
	terms, err := decode_n(n, decode_term, buf)
	return constructor(terms), err
}

func decode_cast(buf *bytes.Buffer) (Term, error) {
	var (
		bitwidth uint16
		term     Term
		err      error
	)
	// Exponent
	if err := binary.Read(buf, binary.BigEndian, &bitwidth); err != nil {
		return nil, err
	}
	// Term
	if term, err = decode_term(buf); err != nil {
		return term, err
	}
	// Done
	return ir.CastOf(term, uint(bitwidth)), nil
}

func decode_constant(buf *bytes.Buffer) (Term, error) {
	var (
		bytes   [32]byte
		element fr.Element
	)
	//
	if n, err := buf.Read(bytes[:]); err != nil {
		return nil, err
	} else if n != 32 {
		return nil, errors.New("failed decoding constant")
	}
	//
	element.SetBytes(bytes[:])
	//
	return ir.Const[Term](element), nil
}

func decode_exponent(buf *bytes.Buffer) (Term, error) {
	var (
		exponent uint64
		term     Term
		err      error
	)
	// Exponent
	if err := binary.Read(buf, binary.BigEndian, &exponent); err != nil {
		return nil, err
	}
	// Term
	if term, err = decode_term(buf); err != nil {
		return term, err
	}
	// Done
	return ir.Exponent(term, exponent), nil
}

func decode_ifzero(buf *bytes.Buffer) (Term, error) {
	var (
		condition LogicalTerm
		branches  []Term
		err       error
	)
	// Condition
	if condition, err = decode_logical(buf); err != nil {
		return &IfZero{}, err
	}
	// True / false branches
	if branches, err = decode_n(2, decode_term, buf); err != nil {
		return &IfZero{}, err
	}
	// Done
	return ir.IfElse(condition, branches[0], branches[1]), nil
}

func decode_labelled_constant(buf *bytes.Buffer) (Term, error) {
	var (
		str_bytes   []byte
		str_len     uint16
		const_bytes [32]byte
		element     fr.Element
	)
	// Label length
	if err := binary.Read(buf, binary.BigEndian, &str_len); err != nil {
		return nil, err
	}
	// Label contents
	str_bytes = make([]byte, str_len)
	if n, err := buf.Read(str_bytes); err != nil {
		return nil, err
	} else if n != int(str_len) {
		return nil, errors.New("failed decoding labelled constant")
	}
	// Constant
	if n, err := buf.Read(const_bytes[:]); err != nil {
		return nil, err
	} else if n != 32 {
		return nil, errors.New("failed decoding labelled constant")
	}
	//
	element.SetBytes(const_bytes[:])
	//
	return ir.LabelledConstant[Term](string(str_bytes), element), nil
}

func decode_reg_access(buf *bytes.Buffer) (*RegisterAccess, error) {
	var (
		index uint16
		shift int16
	)
	// Register index
	if err := binary.Read(buf, binary.BigEndian, &index); err != nil {
		return nil, err
	}
	// Register shift
	if err := binary.Read(buf, binary.BigEndian, &shift); err != nil {
		return nil, err
	}
	// Construct raw register id
	rid := schema.NewRegisterId(uint(index))
	// Done
	return &ir.RegisterAccess[Term]{Register: rid, Shift: int(shift)}, nil
}

func decode_vec_access(buf *bytes.Buffer) (Term, error) {
	vars, err := decode_nary(decode_reg_access, buf)
	//
	return &ir.VectorAccess[Term]{Vars: vars}, err
}

// ============================================================================
// Helpers
// ============================================================================

func encode_nary[T any](encoder func(T, *bytes.Buffer) error, buf *bytes.Buffer, terms []T) error {
	var n byte = byte(len(terms))
	// Write n
	if err := buf.WriteByte(n); err != nil {
		return err
	}
	//
	return encode_n(encoder, buf, terms...)
}

func encode_n[T any](encoder func(T, *bytes.Buffer) error, buf *bytes.Buffer, terms ...T) error {
	//
	for _, t := range terms {
		if err := encoder(t, buf); err != nil {
			return err
		}
	}
	//
	return nil
}

func decode_nary[T any](decoder func(*bytes.Buffer) (T, error), buf *bytes.Buffer) ([]T, error) {
	var (
		// NOTE: hard limit enforced here that we have at most 256 terms.
		n, err = buf.ReadByte()
	)
	//
	if err != nil {
		return nil, err
	}
	//
	return decode_n(uint(n), decoder, buf)
}

func decode_n[T any](n uint, decoder func(*bytes.Buffer) (T, error), buf *bytes.Buffer) ([]T, error) {
	var (
		terms = make([]T, n)
		err   error
	)
	//
	for i := range terms {
		if terms[i], err = decoder(buf); err != nil {
			return nil, err
		}
	}
	//
	return terms, nil
}

// ============================================================================
// Constructors
// ============================================================================

func addConstructor(terms []Term) Term {
	return ir.Sum(terms...)
}

func conjunctionConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.Conjunction(terms...)
}

func disjunctionConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.Disjunction(terms...)
}

func equalConstructor(terms []Term) LogicalTerm {
	return ir.Equals[LogicalTerm](terms[0], terms[1])
}

func iteTrueFalseConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.IfThenElse(terms[0], terms[1], terms[2])
}

func iteTrueConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.IfThenElse(terms[0], terms[1], nil)
}

func iteFalseConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.IfThenElse(terms[0], nil, terms[1])
}

func mulConstructor(terms []Term) Term {
	return ir.Product(terms...)
}

func negationConstructor(terms []LogicalTerm) LogicalTerm {
	return ir.Negation[LogicalTerm](terms[0])
}

func notEqualConstructor(terms []Term) LogicalTerm {
	return ir.NotEquals[LogicalTerm](terms[0], terms[1])
}

func lessThanConstructor(terms []Term) LogicalTerm {
	return ir.LessThan[LogicalTerm](terms[0], terms[1])
}

func lessThanOrEqualsConstructor(terms []Term) LogicalTerm {
	return ir.LessThanOrEquals[LogicalTerm](terms[0], terms[1])
}

func subConstructor(terms []Term) Term {
	return ir.Subtract(terms...)
}

func normConstructor(terms []Term) Term {
	return ir.Normalise(terms[0])
}

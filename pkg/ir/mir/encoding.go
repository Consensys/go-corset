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

	"github.com/consensys/go-corset/pkg/ir"
	"github.com/consensys/go-corset/pkg/schema"
	"github.com/consensys/go-corset/pkg/schema/constraint/lookup"
	"github.com/consensys/go-corset/pkg/util"
	"github.com/consensys/go-corset/pkg/util/field"
)

const (
	// Constraints
	assertionTag        = byte(0)
	interleavingTag     = byte(1)
	lookupTag           = byte(2)
	permutationTag      = byte(3)
	rangeTag            = byte(4)
	sortedUnfilteredTag = byte(5)
	sortedFilteredTag   = byte(6)
	vanishingTag        = byte(7)
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
	expTag              = byte(36)
	mulTag              = byte(37)
	normTag             = byte(38)
	subTag              = byte(39)
	vectorAccessTag     = byte(40)
)

func encode_constraint[F field.Element[F]](constraint schema.Constraint[F]) ([]byte, error) {
	switch c := constraint.(type) {
	case Assertion[F]:
		return encode_assertion(c)
	case InterleavingConstraint[F]:
		return encode_interleaving(c)
	case LookupConstraint[F]:
		return encode_lookup(c)
	case PermutationConstraint[F]:
		return encode_permutation(c)
	case SortedConstraint[F]:
		return encode_sorted(c)
	case RangeConstraint[F]:
		return encode_range(c)
	case VanishingConstraint[F]:
		return encode_vanishing(c)
	default:
		return nil, errors.New("unknown constraint")
	}
}

func encode_assertion[F field.Element[F]](c Assertion[F]) ([]byte, error) {
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
	// Domain
	if err := gobEncoder.Encode(&c.Domain); err != nil {
		return nil, err
	}
	// Constraint
	err := encode_logical(c.Property, &buffer)
	// Done
	return buffer.Bytes(), err
}

func encode_interleaving[F field.Element[F]](c InterleavingConstraint[F]) ([]byte, error) {
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
	if err := encode_nary(encode_term[F], &buffer, c.Sources); err != nil {
		return nil, err
	}
	//
	return buffer.Bytes(), nil
}

func encode_lookup[F field.Element[F]](c LookupConstraint[F]) ([]byte, error) {
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

func encode_lookup_vector[F field.Element[F]](vector lookup.Vector[F, Term[F]], buffer *bytes.Buffer) error {
	var (
		gobEncoder = gob.NewEncoder(buffer)
		selector   = vector.HasSelector()
	)
	// Source Context
	if err := gobEncoder.Encode(vector.Module); err != nil {
		return err
	}
	// HasSelector flag
	if err := gobEncoder.Encode(selector); err != nil {
		return err
	}
	// Selector itself (if applicable)
	if selector {
		if err := encode_term(vector.Selector.Unwrap(), buffer); err != nil {
			return err
		}
	}
	// Source terms
	return encode_nary(encode_term[F], buffer, vector.Terms)
}

func encode_permutation[F field.Element[F]](c PermutationConstraint[F]) ([]byte, error) {
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

func encode_sorted[F field.Element[F]](c SortedConstraint[F]) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
		tag        byte
	)
	//
	if c.Selector.HasValue() {
		tag = sortedFilteredTag
	} else {
		tag = sortedUnfilteredTag
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
	err := encode_nary(encode_term[F], &buffer, c.Sources)
	//
	return buffer.Bytes(), err
}

func encode_vanishing[F field.Element[F]](c VanishingConstraint[F]) ([]byte, error) {
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

func encode_range[F field.Element[F]](c RangeConstraint[F]) ([]byte, error) {
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

func decode_constraint[F field.Element[F]](bytes []byte) (schema.Constraint[F], error) {
	switch bytes[0] {
	case assertionTag:
		return decode_assertion[F](bytes[1:])
	case interleavingTag:
		return decode_interleaving[F](bytes[1:])
	case lookupTag:
		return decode_lookup[F](bytes[1:])
	case permutationTag:
		return decode_permutation[F](bytes[1:])
	case rangeTag:
		return decode_range[F](bytes[1:])
	case sortedUnfilteredTag:
		return decode_sorted[F](false, bytes[1:])
	case sortedFilteredTag:
		return decode_sorted[F](true, bytes[1:])
	case vanishingTag:
		return decode_vanishing[F](bytes[1:])
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", bytes[0])
	}
}

func decode_assertion[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		assertion  Assertion[F]
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
	// Domain
	if err = gobDecoder.Decode(&assertion.Domain); err != nil {
		return assertion, err
	}
	//
	assertion.Property, err = decode_logical[F](buffer)
	// Success!
	return assertion, err
}

func decode_interleaving[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer       = bytes.NewBuffer(data)
		gobDecoder   = gob.NewDecoder(buffer)
		interleaving InterleavingConstraint[F]
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
	if interleaving.Target, err = decode_term[F](buffer); err != nil {
		return interleaving, err
	}
	// Source Context
	if err = gobDecoder.Decode(&interleaving.SourceContext); err != nil {
		return interleaving, err
	}
	// Sources
	if interleaving.Sources, err = decode_nary(decode_term[F], buffer); err != nil {
		return interleaving, err
	}
	//
	return interleaving, nil
}

func decode_lookup[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		lookup     LookupConstraint[F]
		err        error
	)
	// Handle
	if err = gobDecoder.Decode(&lookup.Handle); err != nil {
		return lookup, err
	}
	// Targets
	if lookup.Targets, err = decode_nary(decode_lookup_vector[F], buffer); err != nil {
		return lookup, err
	}
	// Sources
	if lookup.Sources, err = decode_nary(decode_lookup_vector[F], buffer); err != nil {
		return lookup, err
	}
	//
	return lookup, nil
}

func decode_lookup_vector[F field.Element[F]](buf *bytes.Buffer) (lookup.Vector[F, Term[F]], error) {
	var (
		gobDecoder  = gob.NewDecoder(buf)
		vector      lookup.Vector[F, Term[F]]
		hasSelector bool
		selector    Term[F]
		err         error
	)
	// Context
	if err = gobDecoder.Decode(&vector.Module); err != nil {
		return vector, err
	}
	// HasSelector
	if err = gobDecoder.Decode(&hasSelector); err != nil {
		return vector, err
	}
	// Selector (if applicable)
	if hasSelector {
		if selector, err = decode_term[F](buf); err != nil {
			return vector, err
		}
		// Wrap selector
		vector.Selector = util.Some(selector)
	}
	// Contents
	if vector.Terms, err = decode_nary(decode_term[F], buf); err != nil {
		return vector, err
	}
	// Done
	return vector, nil
}

func decode_permutation[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer      = bytes.NewBuffer(data)
		gobDecoder  = gob.NewDecoder(buffer)
		permutation PermutationConstraint[F]
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

func decode_sorted[F field.Element[F]](selector bool, data []byte) (schema.Constraint[F], error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		sorted     SortedConstraint[F]
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
		var term Term[F]
		//
		if term, err = decode_term[F](buffer); err != nil {
			return nil, err
		}
		//
		sorted.Selector = util.Some(term)
	}
	// Sources
	sorted.Sources, err = decode_nary(decode_term[F], buffer)
	// Done
	return sorted, err
}

func decode_range[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		constraint RangeConstraint[F]
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
	constraint.Expr, err = decode_term[F](buffer)
	// Success!
	return constraint, err
}

func decode_vanishing[F field.Element[F]](data []byte) (schema.Constraint[F], error) {
	var (
		buffer     = bytes.NewBuffer(data)
		gobDecoder = gob.NewDecoder(buffer)
		vanishing  VanishingConstraint[F]
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
	vanishing.Constraint, err = decode_logical[F](buffer)
	// Success!
	return vanishing, err
}

// ============================================================================
// Logical Terms (encoding)
// ============================================================================

func encode_logical[F field.Element[F]](term LogicalTerm[F], buf *bytes.Buffer) error {
	switch t := term.(type) {
	case *Conjunct[F]:
		return encode_tagged_nary_logicals(conjunctTag, buf, t.Args...)
	case *Disjunct[F]:
		return encode_tagged_nary_logicals(disjunctTag, buf, t.Args...)
	case *Equal[F]:
		return encode_tagged_terms(equalTag, buf, t.Lhs, t.Rhs)
	case *Ite[F]:
		return encode_ite(t, buf)
	case *Negate[F]:
		return encode_tagged_logicals(negationTag, buf, t.Arg)
	case *NotEqual[F]:
		return encode_tagged_terms(notEqualTag, buf, t.Lhs, t.Rhs)
	case *Inequality[F]:
		if t.Strict {
			return encode_tagged_terms(lessThanTag, buf, t.Lhs, t.Rhs)
		}
		//
		return encode_tagged_terms(lessThanEqTag, buf, t.Lhs, t.Rhs)
	default:
		return fmt.Errorf("unknown logical term encountered (%s)", term.Lisp(false, nil).String(false))
	}
}

func encode_tagged_nary_logicals[F field.Element[F]](tag byte, buf *bytes.Buffer, terms ...LogicalTerm[F]) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_nary(encode_logical[F], buf, terms)
}

func encode_tagged_logicals[F field.Element[F]](tag byte, buf *bytes.Buffer, terms ...LogicalTerm[F]) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_n(encode_logical[F], buf, terms...)
}

func encode_ite[F field.Element[F]](term *Ite[F], buf *bytes.Buffer) error {
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

func decode_logical[F field.Element[F]](buf *bytes.Buffer) (LogicalTerm[F], error) {
	tag, err := buf.ReadByte()
	//
	if err != nil {
		return nil, err
	}
	//
	switch tag {
	case conjunctTag:
		return decode_nary_logicals[F](conjunctionConstructor, buf)
	case disjunctTag:
		return decode_nary_logicals[F](disjunctionConstructor, buf)
	case equalTag:
		return decode_terms[F](2, equalConstructor, buf)
	case iteTagTF, iteTagT, iteTagF:
		return decode_ite[F](tag, buf)
	case negationTag:
		return decode_logicals[F](1, negationConstructor, buf)
	case notEqualTag:
		return decode_terms[F](2, notEqualConstructor, buf)
	case lessThanTag:
		return decode_terms[F](2, lessThanConstructor, buf)
	case lessThanEqTag:
		return decode_terms[F](2, lessThanOrEqualsConstructor, buf)
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", tag)
	}
}

// Decode a variable number of terms, as determined by the leading byte.
func decode_nary_logicals[F field.Element[F]](constructor func([]LogicalTerm[F]) LogicalTerm[F], buf *bytes.Buffer,
) (LogicalTerm[F], error) {
	//
	terms, err := decode_nary(decode_logical[F], buf)
	return constructor(terms), err
}

// Decode exactly n logicals terms
func decode_logicals[F field.Element[F], S any](n uint, constructor func([]LogicalTerm[F]) S, buf *bytes.Buffer,
) (S, error) {
	//
	terms, err := decode_n(n, decode_logical[F], buf)
	return constructor(terms), err
}

func decode_ite[F field.Element[F]](tag byte, buf *bytes.Buffer) (LogicalTerm[F], error) {
	//
	switch tag {
	case iteTagTF:
		return decode_logicals[F](3, iteTrueFalseConstructor, buf)
	case iteTagT:
		return decode_logicals[F](2, iteTrueConstructor, buf)
	case iteTagF:
		return decode_logicals[F](2, iteFalseConstructor, buf)
	default:
		panic("unreachable")
	}
}

// ============================================================================
// Arithmetic Terms (encoding)
// ============================================================================

func encode_term[F field.Element[F]](term Term[F], buf *bytes.Buffer) error {
	//
	switch t := term.(type) {
	case *Add[F]:
		return encode_tagged_nary_terms(addTag, buf, t.Args...)
	case *Cast[F]:
		return encode_cast(*t, buf)
	case *Constant[F]:
		return encode_constant(*t, buf)
	case *Exp[F]:
		return encode_exponent(*t, buf)
	case *IfZero[F]:
		return encode_ifZero(*t, buf)
	case *LabelledConst[F]:
		return encode_labelled_constant(*t, buf)
	case *Mul[F]:
		return encode_tagged_nary_terms(mulTag, buf, t.Args...)
	case *Norm[F]:
		return encode_tagged_terms(normTag, buf, t.Arg)
	case *RegisterAccess[F]:
		return encode_reg_access(*t, buf)
	case *Sub[F]:
		return encode_tagged_nary_terms(subTag, buf, t.Args...)
	case *VectorAccess[F]:
		return encode_vec_access(*t, buf)
	default:
		return fmt.Errorf("unknown arithmetic term encountered (%s)", term.Lisp(false, nil).String(false))
	}
}

func encode_tagged_nary_terms[F field.Element[F]](tag byte, buf *bytes.Buffer, terms ...Term[F]) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_nary(encode_term[F], buf, terms)
}

func encode_tagged_terms[F field.Element[F]](tag byte, buf *bytes.Buffer, terms ...Term[F]) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	return encode_n(encode_term[F], buf, terms...)
}

func encode_cast[F field.Element[F]](term Cast[F], buf *bytes.Buffer) error {
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

func encode_constant[F field.Element[F]](term Constant[F], buf *bytes.Buffer) error {
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

func encode_ifZero[F field.Element[F]](term IfZero[F], buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(ifZeroTag); err != nil {
		return err
	}
	// Write condition
	if err := encode_logical(term.Condition, buf); err != nil {
		return err
	}
	// Write true + false branches
	return encode_n(encode_term[F], buf, term.TrueBranch, term.FalseBranch)
}

func encode_labelled_constant[F field.Element[F]](term LabelledConst[F], buf *bytes.Buffer) error {
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

func encode_exponent[F field.Element[F]](term Exp[F], buf *bytes.Buffer) error {
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

func encode_reg_access[F field.Element[F]](term RegisterAccess[F], buf *bytes.Buffer) error {
	// Write (appropriate) tag
	if err := buf.WriteByte(registerAccessTag); err != nil {
		return err
	}
	//
	return encode_raw_access(&term, buf)
}

func encode_vec_access[F field.Element[F]](term VectorAccess[F], buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(vectorAccessTag); err != nil {
		return err
	}
	//
	return encode_nary(encode_raw_access, buf, term.Vars)
}

func encode_raw_access[F field.Element[F]](term *RegisterAccess[F], buf *bytes.Buffer) error {
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
func decode_term[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	tag, err := buf.ReadByte()
	//
	if err != nil {
		return nil, err
	}
	//
	switch tag {
	case addTag:
		return decode_nary_terms[F](addConstructor, buf)
	case castTag:
		return decode_cast[F](buf)
	case constantTag:
		return decode_constant[F](buf)
	case expTag:
		return decode_exponent[F](buf)
	case ifZeroTag:
		return decode_ifzero[F](buf)
	case labelledConstantTag:
		return decode_labelled_constant[F](buf)
	case registerAccessTag:
		return decode_reg_access[F](buf)
	case mulTag:
		return decode_nary_terms[F](mulConstructor, buf)
	case normTag:
		return decode_terms[F](1, normConstructor, buf)
	case subTag:
		return decode_nary_terms[F](subConstructor, buf)
	case vectorAccessTag:
		return decode_vec_access[F](buf)
	default:
		return nil, fmt.Errorf("unknown constraint (tag %d)", tag)
	}
}

// Decode a variable number of terms, as determined by the leading byte.
func decode_nary_terms[F field.Element[F]](constructor func([]Term[F]) Term[F], buf *bytes.Buffer) (Term[F], error) {
	terms, err := decode_nary(decode_term[F], buf)
	return constructor(terms), err
}

// Decode exactly n terms
func decode_terms[F field.Element[F], S any](n uint, constructor func([]Term[F]) S, buf *bytes.Buffer) (S, error) {
	terms, err := decode_n(n, decode_term[F], buf)
	return constructor(terms), err
}

func decode_cast[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	var (
		bitwidth uint16
		term     Term[F]
		err      error
	)
	// Exponent
	if err := binary.Read(buf, binary.BigEndian, &bitwidth); err != nil {
		return nil, err
	}
	// Term
	if term, err = decode_term[F](buf); err != nil {
		return term, err
	}
	// Done
	return ir.CastOf(term, uint(bitwidth)), nil
}

func decode_constant[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	var (
		bytes   [32]byte
		element F
	)
	//
	if n, err := buf.Read(bytes[:]); err != nil {
		return nil, err
	} else if n != 32 {
		return nil, errors.New("failed decoding constant")
	}
	//
	element = element.SetBytes(bytes[:])
	//
	return ir.Const[F, Term[F]](element), nil
}

func decode_exponent[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	var (
		exponent uint64
		term     Term[F]
		err      error
	)
	// Exponent
	if err := binary.Read(buf, binary.BigEndian, &exponent); err != nil {
		return nil, err
	}
	// Term
	if term, err = decode_term[F](buf); err != nil {
		return term, err
	}
	// Done
	return ir.Exponent(term, exponent), nil
}

func decode_ifzero[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	var (
		condition LogicalTerm[F]
		branches  []Term[F]
		err       error
	)
	// Condition
	if condition, err = decode_logical[F](buf); err != nil {
		return &IfZero[F]{}, err
	}
	// True / false branches
	if branches, err = decode_n(2, decode_term[F], buf); err != nil {
		return &IfZero[F]{}, err
	}
	// Done
	return ir.IfElse(condition, branches[0], branches[1]), nil
}

func decode_labelled_constant[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	var (
		str_bytes   []byte
		str_len     uint16
		const_bytes [32]byte
		element     F
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
	element = element.SetBytes(const_bytes[:])
	//
	return ir.LabelledConstant[F, Term[F]](string(str_bytes), element), nil
}

func decode_reg_access[F field.Element[F]](buf *bytes.Buffer) (*RegisterAccess[F], error) {
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
	return &ir.RegisterAccess[F, Term[F]]{Register: rid, Shift: int(shift)}, nil
}

func decode_vec_access[F field.Element[F]](buf *bytes.Buffer) (Term[F], error) {
	vars, err := decode_nary(decode_reg_access[F], buf)
	//
	return &ir.VectorAccess[F, Term[F]]{Vars: vars}, err
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

func addConstructor[F field.Element[F]](terms []Term[F]) Term[F] {
	return ir.Sum(terms...)
}

func conjunctionConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.Conjunction(terms...)
}

func disjunctionConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.Disjunction(terms...)
}

func equalConstructor[F field.Element[F]](terms []Term[F]) LogicalTerm[F] {
	return ir.Equals[F, LogicalTerm[F]](terms[0], terms[1])
}

func iteTrueFalseConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.IfThenElse(terms[0], terms[1], terms[2])
}

func iteTrueConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.IfThenElse(terms[0], terms[1], nil)
}

func iteFalseConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.IfThenElse(terms[0], nil, terms[1])
}

func mulConstructor[F field.Element[F]](terms []Term[F]) Term[F] {
	return ir.Product(terms...)
}

func negationConstructor[F field.Element[F]](terms []LogicalTerm[F]) LogicalTerm[F] {
	return ir.Negation(terms[0])
}

func notEqualConstructor[F field.Element[F]](terms []Term[F]) LogicalTerm[F] {
	return ir.NotEquals[F, LogicalTerm[F]](terms[0], terms[1])
}

func lessThanConstructor[F field.Element[F]](terms []Term[F]) LogicalTerm[F] {
	return ir.LessThan[F, LogicalTerm[F]](terms[0], terms[1])
}

func lessThanOrEqualsConstructor[F field.Element[F]](terms []Term[F]) LogicalTerm[F] {
	return ir.LessThanOrEquals[F, LogicalTerm[F]](terms[0], terms[1])
}

func subConstructor[F field.Element[F]](terms []Term[F]) Term[F] {
	return ir.Subtract(terms...)
}

func normConstructor[F field.Element[F]](terms []Term[F]) Term[F] {
	return ir.Normalise(terms[0])
}

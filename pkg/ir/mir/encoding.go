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
)

const (
	// Constraints
	lookupTag      = byte(1)
	permutationTag = byte(2)
	sortedTag      = byte(3)
	rangeTag       = byte(4)
	vanishingTag   = byte(5)
	// Logicals
	conjunctTag = byte(10)
	disjunctTag = byte(11)
	equalTag    = byte(12)
	notEqualTag = byte(13)
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
)

func encode_constraint(constraint schema.Constraint) ([]byte, error) {
	switch c := constraint.(type) {
	case LookupConstraint:
		panic("todo")
	case PermutationConstraint:
		panic("todo")
	case SortedConstraint:
		panic("todo")
	case RangeConstraint:
		panic("todo")
	case VanishingConstraint:
		return encode_vanishing(c)
	default:
		return nil, errors.New("unknown MIR constraint")
	}
}

func encode_vanishing(c VanishingConstraint) ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder = gob.NewEncoder(&buffer)
	)
	//
	buffer.Write([]byte{vanishingTag})
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

func decode_constraint(bytes []byte) (schema.Constraint, error) {
	switch bytes[0] {
	case lookupTag:
		panic("todo")
	case permutationTag:
		panic("todo")
	case rangeTag:
		panic("todo")
	case sortedTag:
		panic("todo")
	case vanishingTag:
		return decode_vanishing(bytes[1:])
	default:
		return nil, fmt.Errorf("unknown MIR constraint (tag %d)", bytes[0])
	}
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
// Logical Terms
// ============================================================================

func encode_logical(term LogicalTerm, buf *bytes.Buffer) error {
	switch t := term.(type) {
	case *Equal:
		return encode_terms(equalTag, buf, t.Lhs, t.Rhs)
	case *NotEqual:
		return encode_terms(notEqualTag, buf, t.Lhs, t.Rhs)
	default:
		return errors.New("unknown MIR term encountered")
	}
}

func decode_logical(buf *bytes.Buffer) (LogicalTerm, error) {
	tag, err := buf.ReadByte()
	//
	if err != nil {
		return nil, err
	}
	//
	switch tag {
	case conjunctTag:
		panic("todo")
	case disjunctTag:
		panic("todo")
	case equalTag:
		return decode_terms(2, equalConstructor, buf)
	case notEqualTag:
		return decode_terms(2, notEqualConstructor, buf)
	default:
		return nil, fmt.Errorf("unknown MIR constraint (tag %d)", tag)
	}
}

// ============================================================================
// Arithmetic Terms (encoding)
// ============================================================================

func encode_term(term Term, buf *bytes.Buffer) error {
	switch t := term.(type) {
	case *Add:
		return encode_nary_terms(addTag, buf, t.Args...)
	case *Cast:
		panic("todo")
	case *Constant:
		return encode_const(*t, buf)
	case *Exp:
		panic("todo")
	case *IfZero:
		panic("todo")
	case *LabelledConst:
		panic("todo")
	case *Mul:
		return encode_nary_terms(mulTag, buf, t.Args...)
	case *Norm:
		panic("todo")
	case *RegisterAccess:
		return encode_reg_access(*t, buf)
	case *Sub:
		return encode_nary_terms(subTag, buf, t.Args...)
	default:
		return errors.New("unknown MIR term encountered")
	}
}

func encode_nary_terms(tag byte, buf *bytes.Buffer, terms ...Term) error {
	var n byte = byte(len(terms))
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	// Write n
	if err := buf.WriteByte(n); err != nil {
		return err
	}
	// Write terms
	for _, t := range terms {
		if err := encode_term(t, buf); err != nil {
			return err
		}
	}
	//
	return nil
}

func encode_terms(tag byte, buf *bytes.Buffer, terms ...Term) error {
	// Write tag
	if err := buf.WriteByte(tag); err != nil {
		return err
	}
	//
	for _, t := range terms {
		if err := encode_term(t, buf); err != nil {
			return err
		}
	}
	//
	return nil
}

func encode_const(term Constant, buf *bytes.Buffer) error {
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

func encode_reg_access(term RegisterAccess, buf *bytes.Buffer) error {
	// Write tag
	if err := buf.WriteByte(registerAccessTag); err != nil {
		return err
	}
	// Register Index
	if err := binary.Write(buf, binary.BigEndian, uint16(term.Register)); err != nil {
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
		panic("todo")
	case constantTag:
		return decode_constant(buf)
	case expTag:
		panic("todo")
	case ifZeroTag:
		panic("todo")
	case labelledConstantTag:
		panic("todo")
	case registerAccessTag:
		return decode_register(buf)
	case mulTag:
		return decode_nary_terms(mulConstructor, buf)
	case normTag:
		panic("todo")
	case subTag:
		return decode_nary_terms(subConstructor, buf)
	default:
		return nil, fmt.Errorf("unknown MIR constraint (tag %d)", tag)
	}
}

// Decode a variable number of terms, as determined by the leading byte.
func decode_nary_terms[S any](constructor func([]Term) S, buf *bytes.Buffer) (S, error) {
	var (
		dummy S
		// NOTE: hard limit enforced here that we have at most 256 terms.
		n, err = buf.ReadByte()
	)
	//
	if err != nil {
		return dummy, err
	}
	//
	return decode_terms(uint(n), constructor, buf)
}

// Decode exactly n terms
func decode_terms[S any](n uint, constructor func([]Term) S, buf *bytes.Buffer) (S, error) {
	var (
		dummy S
		terms = make([]Term, n)
		err   error
	)
	//
	for i := range terms {
		if terms[i], err = decode_term(buf); err != nil {
			return dummy, err
		}
	}
	//
	return constructor(terms), nil
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
		return nil, errors.New("failed decoding MIR constant")
	}
	//
	element.SetBytes(bytes[:])
	//
	return ir.Const[Term](element), nil
}

func decode_register(buf *bytes.Buffer) (Term, error) {
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
	// Done
	return ir.NewRegisterAccess[Term](uint(index), int(shift)), nil
}

// ============================================================================
// Constructors
// ============================================================================

func addConstructor(terms []Term) Term {
	return ir.Sum(terms...)
}

func equalConstructor(terms []Term) LogicalTerm {
	return ir.Equals[LogicalTerm](terms[0], terms[1])
}

func mulConstructor(terms []Term) Term {
	return ir.Product(terms...)
}

func notEqualConstructor(terms []Term) LogicalTerm {
	return ir.NotEquals[LogicalTerm](terms[0], terms[1])
}

func subConstructor(terms []Term) Term {
	return ir.Subtract(terms...)
}

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
package binfile

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/util/collection/typed"
)

// ============================================================================
// Binary File Format
// ============================================================================

// BinaryFile is a programatic represresentation of an underlying binary file.
type BinaryFile struct {
	// Header for the binary file
	Header Header
	// Attributes for the binary file.  These hold, for example, information for
	// debugging, etc.
	Attributes []Attribute
	// The HIR Schema itself.
	Schema hir.Schema
}

// NewBinaryFile constructs a new binary file with the default header for the
// currently supported version.
func NewBinaryFile(metadata []byte, attributes []Attribute, schema *hir.Schema) *BinaryFile {
	return &BinaryFile{
		Header{ZKBINARY, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION, metadata},
		attributes,
		*schema,
	}
}

// GetAttribute returns the first instance of a given attribute, or nil if none
// exists.
func GetAttribute[T Attribute](binf *BinaryFile) (T, bool) {
	var empty T
	//
	for _, attr := range binf.Attributes {
		if a, ok := attr.(T); ok {
			return a, true
		}
	}
	//
	return empty, false
}

// Attribute is essentially an abstraction allowing arbitrary
// attributes to be embedded alongside the schema.  These can be used, for
// example, for storing mapping information from source columns to allocated
// columns, etc.
type Attribute interface {
	// AttributeName returns the name of this attribute.
	AttributeName() string
}

// Header provides a structured header for the binary file format.  In
// particular, it supports versioning and embedded (binary) metadata.
type Header struct {
	Identifier   [8]byte
	MajorVersion uint16
	MinorVersion uint16
	MetaData     []byte
}

// GetMetaData attempts to parse the metadata bytes as JSON which is then
// unmarshalled into a map.  This can fail if the embedded metadata bytes are
// not, in fact, JSON.  Observe that, if there are no metadata bytes, then nil
// will be returned.
func (p *Header) GetMetaData() (typed.Map, error) {
	// Check for empty metadata
	if len(p.MetaData) == 0 {
		return typed.NewMap(nil), nil
	}
	// Attempt to unmarshal metadata bytes
	return typed.FromJsonBytes(p.MetaData)
}

// SetMetaData attempts to set the metadata bytes for this header, using a JSON
// encoding of the given map.  If this fails, an error is returned and the
// metadata bytes are unaffected.
func (p *Header) SetMetaData(metadata typed.Map) error {
	bytes, err := metadata.ToJsonBytes()
	// Check for error
	if err != nil {
		return err
	}
	// success
	p.MetaData = bytes
	//
	return nil
}

// MarshalBinary converts the BinaryFile Header into a sequence of bytes.
// Observe that we don't use GobEncoding here to avoid being tied to that
// encoding scheme.
func (p *Header) MarshalBinary() ([]byte, error) {
	var (
		buffer     bytes.Buffer
		majorBytes [2]byte
		minorBytes [2]byte
		metaLength [4]byte
	)
	// Marshall version numbers
	binary.BigEndian.PutUint16(majorBytes[:], p.MajorVersion)
	binary.BigEndian.PutUint16(minorBytes[:], p.MinorVersion)
	binary.BigEndian.PutUint32(metaLength[:], uint32(len(p.MetaData)))
	// Write identifier
	buffer.Write(p.Identifier[:])
	// Write major version
	buffer.Write(majorBytes[:])
	// Write minor version
	buffer.Write(minorBytes[:])
	// Write metadata length
	buffer.Write(metaLength[:])
	// Write metadata itself
	buffer.Write(p.MetaData)
	// Done
	return buffer.Bytes(), nil
}

// UnmarshalBinary initialises this BinaryFile Header from a given set of data bytes.
// This should match exactly the encoding above.
func (p *Header) UnmarshalBinary(buffer *bytes.Buffer) error {
	var (
		majorBytes      [2]byte
		minorBytes      [2]byte
		metaLengthBytes [4]byte
	)
	// Read identifier
	if n, err := buffer.Read(p.Identifier[:]); err != nil {
		return err
	} else if n != 8 {
		return errors.New("malformed binary file")
	}
	// Read major version
	if n, err := buffer.Read(majorBytes[:]); err != nil {
		return err
	} else if n != len(majorBytes) {
		return errors.New("malformed binary file")
	}
	// Read minor version
	if n, err := buffer.Read(minorBytes[:]); err != nil {
		return err
	} else if n != len(minorBytes) {
		return errors.New("malformed binary file")
	}
	// Read metadata length
	if n, err := buffer.Read(metaLengthBytes[:]); err != nil {
		return err
	} else if n != len(metaLengthBytes) {
		return errors.New("malformed binary file")
	}
	// Make space for the metadata
	var (
		metaLength        = binary.BigEndian.Uint32(metaLengthBytes[:])
		metaBytes  []byte = make([]byte, metaLength)
	)
	// Read metadata itself
	if n, err := buffer.Read(metaBytes[:]); err != nil {
		return err
	} else if n != len(metaBytes) {
		return errors.New("malformed binary file")
	}
	// Finally assign everything over
	p.MajorVersion = binary.BigEndian.Uint16(majorBytes[:])
	p.MinorVersion = binary.BigEndian.Uint16(minorBytes[:])
	p.MetaData = metaBytes
	// Done
	return nil
}

// IsCompatible determines whether a given binary file is compatible with this
// version of go-corset.
func (p *Header) IsCompatible() bool {
	//
	return p.Identifier == ZKBINARY &&
		p.MajorVersion == BINFILE_MAJOR_VERSION &&
		p.MinorVersion <= BINFILE_MINOR_VERSION
}

// BINFILE_MAJOR_VERSION givesn the major version of the binary file format.  No
// matter what version, we should always have the ZKBINARY identifier first,
// followed by a GOB encoding of the header.  What follows after that, however,
// is determined by the major version.
const BINFILE_MAJOR_VERSION uint16 = 3

// BINFILE_MINOR_VERSION gives the minor version of the binary file format.  The
// expected interpretation is that older versions are compatible with newer
// ones, but not vice-versa.
const BINFILE_MINOR_VERSION uint16 = 4

// ZKBINARY is used as the file identifier for binary file types.  This just
// helps us identify actual binary files from corrupted files.
var ZKBINARY [8]byte = [8]byte{'z', 'k', 'b', 'i', 'n', 'a', 'r', 'y'}

// IsBinaryFile checks whether the given data file begins with the expected
// "zkbinary" identifier.
func IsBinaryFile(data []byte) bool {
	var (
		zkbinary [8]byte
		buffer   *bytes.Buffer = bytes.NewBuffer(data)
	)
	//
	if _, err := buffer.Read(zkbinary[:]); err != nil {
		return false
	}
	// Check whether header identified
	return zkbinary == ZKBINARY
}

// MarshalBinary converts the BinaryFile into a sequence of bytes.
func (p *BinaryFile) MarshalBinary() ([]byte, error) {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Marshal header
	headerBytes, err := p.Header.MarshalBinary()
	//
	if err != nil {
		return nil, err
	}
	// Encode header
	buffer.Write(headerBytes)
	// Encode attributes
	if err := gobEncoder.Encode(p.Attributes); err != nil {
		return nil, err
	}
	// Encode schema
	if err := gobEncoder.Encode(&p.Schema); err != nil {
		return nil, err
	}
	// Done
	return buffer.Bytes(), nil
}

// UnmarshalBinary initialises this BinaryFile from a given set of data bytes.
// This should match exactly the encoding above.
func (p *BinaryFile) UnmarshalBinary(data []byte) error {
	var err error
	//
	buffer := bytes.NewBuffer(data)
	// Read header
	if err = p.Header.UnmarshalBinary(buffer); err == nil && p.Header.IsCompatible() {
		// Looks good, proceed.
		decoder := gob.NewDecoder(buffer)
		// Proceed to decoding any attributes.
		if err = decoder.Decode(&p.Attributes); err == nil {
			// Finally, decode the schema itself
			err = decoder.Decode(&p.Schema)
		}
	} else if err == nil {
		err = fmt.Errorf("incompatible binary file was v%d.%d, but expected v%d.%d)",
			p.Header.MajorVersion, p.Header.MinorVersion, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION)
	}
	//
	return err
}

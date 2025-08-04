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
package lt

import (
	"bytes"
	"fmt"

	"github.com/consensys/go-corset/pkg/trace"
)

// LT_MAJOR_VERSION givesn the major version of the binary file format.  No
// matter what version, we should always have the ZKBINARY identifier first,
// followed by a GOB encoding of the header.  What follows after that, however,
// is determined by the major version.
const LT_MAJOR_VERSION uint16 = 1

// LT_MINOR_VERSION gives the minor version of the binary file format.  The
// expected interpretation is that older versions are compatible with newer
// ones, but not vice-versa.
const LT_MINOR_VERSION uint16 = 0

// ZKTRACER is used as the file identifier for binary file types.  This just
// helps us identify actual binary files from corrupted files.
var ZKTRACER [8]byte = [8]byte{'z', 'k', 't', 'r', 'a', 'c', 'e', 'r'}

// TraceFile is a programatic represresentation of an underlying trace file.
type TraceFile struct {
	// Header for the binary file
	Header Header
	// Word pool
	Pool WordPool
	// Column data
	Columns []trace.BigEndianColumn
}

// NewTraceFile constructs a new trace file with the default header for the
// currently supported version.
func NewTraceFile(metadata []byte, pool WordPool, columns []trace.BigEndianColumn) TraceFile {
	return TraceFile{
		Header{ZKTRACER, LT_MAJOR_VERSION, LT_MINOR_VERSION, metadata},
		pool,
		columns,
	}
}

// IsTraceFile checks whether the given data file begins with the expected
// "zktracer" identifier.
func IsTraceFile(data []byte) bool {
	var (
		zktracer [8]byte
		buffer   = bytes.NewBuffer(data)
	)
	//
	if _, err := buffer.Read(zktracer[:]); err != nil {
		return false
	}
	// Check whether header identified
	return zktracer == ZKTRACER
}

// MarshalBinary converts the TraceFile into a sequence of bytes.
func (p *TraceFile) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer
	// Bytes header
	headerBytes, err := p.Header.MarshalBinary()
	// Error check
	if err != nil {
		return nil, err
	}
	// Encode header
	buffer.Write(headerBytes)
	// Bytes column data
	columnBytes, err := ToBytesLegacy(p.Columns)
	// Error check
	if err != nil {
		return nil, err
	}
	// Encode column data
	buffer.Write(columnBytes)
	// Done
	return buffer.Bytes(), nil
}

// UnmarshalBinary initialises this TraceFile from a given set of data bytes.
// This should match exactly the encoding above.
func (p *TraceFile) UnmarshalBinary(data []byte) error {
	var err error
	//
	buffer := bytes.NewBuffer(data)
	// Read header
	if err = p.Header.UnmarshalBinary(buffer); err == nil && p.Header.IsCompatible() {
		// Decode column data
		p.Pool, p.Columns, err = FromBytesLegacy(buffer.Bytes())
		// Done
		return err
	} else if err == nil {
		err = fmt.Errorf("incompatible binary file was v%d.%d, but expected v%d.%d)",
			p.Header.MajorVersion, p.Header.MinorVersion, LT_MAJOR_VERSION, LT_MINOR_VERSION)
	}
	//
	return err
}

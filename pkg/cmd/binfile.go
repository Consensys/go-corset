package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"os"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/hir"
)

// ============================================================================
// Binary File Format
// ============================================================================

// BinaryFile is a programatic represresentation of an underlying binary file.
type BinaryFile struct {
	// Header for the binary file
	Header BinaryFileHeader
	// Attributes for the binary file.  These hold, for example, information for
	// debugging, etc.
	Attributes []BinaryFileAttribute
	// The HIR Schema itself.
	Schema hir.Schema
}

// NewBinaryFile constructs a new binary file with the default header for the
// currently supported version.
func NewBinaryFile(headerdata []byte, attributes []BinaryFileAttribute, schema *hir.Schema) *BinaryFile {
	return &BinaryFile{
		BinaryFileHeader{ZKBINARY, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION, headerdata},
		attributes,
		*schema,
	}
}

// BinaryFileHeader provides a structured header for the binary file format.  In
// particular, it supports versioning and embedded (binary) metadata.
type BinaryFileHeader struct {
	Identifier   [8]byte
	MajorVersion uint16
	MinorVersion uint16
	MetaData     []byte
}

// BinaryFileAttribute is essentially an abstraction allowing arbitrary
// attributes to be embedded alongside the schema.  These can be used, for
// example, for storing mapping information from source columns to allocated
// columns, etc.
type BinaryFileAttribute interface {
}

// MarshalBinary converts the BinaryFile header into a sequence of bytes.
// Observe that we don't use GobEncoding here to avoid being tied to that
// encoding scheme.
func (p *BinaryFileHeader) MarshalBinary() ([]byte, error) {
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

// UnmarshalBinary initialises this BinaryFile from a given set of data bytes.
// This should match exactly the encoding above.
func (p *BinaryFileHeader) UnmarshalBinary(buffer *bytes.Buffer) error {
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

// Determine whether a given binary file is compatible with this version of
// go-corset.
func (p *BinaryFileHeader) isCompatible() bool {
	//
	return p.Identifier == ZKBINARY &&
		p.MajorVersion == BINFILE_MAJOR_VERSION &&
		p.MinorVersion <= BINFILE_MINOR_VERSION
}

// BINFILE_MAJOR_VERSION givesn the major version of the binary file format.  No
// matter what version, we should always have the ZKBINARY identifier first,
// followed by a GOB encoding of the header.  What follows after that, however,
// is determined by the major version.
const BINFILE_MAJOR_VERSION uint16 = 2

// BINFILE_MINOR_VERSION gives the minor version of the binary file format.  The
// expected interpretation is that older versions are compatible with newer
// ones, but not vice-versa.
const BINFILE_MINOR_VERSION uint16 = 0

// ZKBINARY is used as the file identifier for binary file types.  This just
// helps us identify actual binary files from corrupted files.
var ZKBINARY [8]byte = [8]byte{'z', 'k', 'b', 'i', 'n', 'a', 'r', 'y'}

// Read a "bin" file and extract the metadata bytes, along with the schema.
func readBinaryFile(legacy bool, filename string) *BinaryFile {
	var (
		header     BinaryFileHeader
		schema     *hir.Schema
		attributes []BinaryFileAttribute
	)
	// Read schema file
	data, err := os.ReadFile(filename)
	// Handle errors
	if err == nil && (legacy || !isBinaryFile(data)) {
		// Read the binary file
		schema, err = binfile.HirSchemaFromJson(data)
	} else if err == nil {
		// Read the Gob file
		buffer := bytes.NewBuffer(data)
		// Read header
		if err = header.UnmarshalBinary(buffer); err == nil && header.isCompatible() {
			// Looks good, proceed.
			decoder := gob.NewDecoder(buffer)
			// Proceed to decoding any attributes.
			if err = decoder.Decode(&attributes); err == nil {
				// Finally, decode the schema itself
				err = decoder.Decode(&schema)
			}
		} else if err == nil {
			err = fmt.Errorf("incompatible binary file \"%s\" (was v%d.%d, but expected v%d.%d)", filename,
				header.MajorVersion, header.MinorVersion, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION)
		}
	}
	// Return if no errors
	if err == nil {
		return &BinaryFile{header, attributes, *schema}
	}
	// Handle error & exit
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Write a binary file using a given set of metadata bytes.
//
//nolint:errcheck
func writeBinaryFile(binfile *BinaryFile, legacy bool, filename string) {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
	)
	// Sanity checks
	if legacy {
		// Currently, there is no support for this.
		fmt.Println("legacy binary format not supported for writing")
	}
	// Marshal header
	headerBytes, _ := binfile.Header.MarshalBinary()
	// Encode header
	buffer.Write(headerBytes)
	// Encode attributes
	if err := gobEncoder.Encode(binfile.Attributes); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Encode schema
	if err := gobEncoder.Encode(binfile.Schema); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Write file
	if err := os.WriteFile(filename, buffer.Bytes(), 0644); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Check whether the given data file begins with the expected "zkbinary"
// identifier.
func isBinaryFile(data []byte) bool {
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

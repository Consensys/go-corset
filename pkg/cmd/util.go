package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/consensys/go-corset/pkg/binfile"
	"github.com/consensys/go-corset/pkg/corset"
	"github.com/consensys/go-corset/pkg/hir"
	"github.com/consensys/go-corset/pkg/trace"
	"github.com/consensys/go-corset/pkg/trace/json"
	"github.com/consensys/go-corset/pkg/trace/lt"
	"github.com/consensys/go-corset/pkg/util/sexp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// GetFlag gets an expected flag, or panic if an error arises.
func GetFlag(cmd *cobra.Command, flag string) bool {
	r, err := cmd.Flags().GetBool(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return r
}

// GetInt gets an expectedsigned integer, or panic if an error arises.
func GetInt(cmd *cobra.Command, flag string) int {
	r, err := cmd.Flags().GetInt(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	return r
}

// GetUint gets an expected unsigned integer, or panic if an error arises.
func GetUint(cmd *cobra.Command, flag string) uint {
	r, err := cmd.Flags().GetUint(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetString gets an expected string, or panic if an error arises.
func GetString(cmd *cobra.Command, flag string) string {
	r, err := cmd.Flags().GetString(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// GetStringArray gets an expected string array, or panic if an error arises.
func GetStringArray(cmd *cobra.Command, flag string) []string {
	r, err := cmd.Flags().GetStringArray(flag)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	return r
}

// Write a given trace file to disk
func writeTraceFile(filename string, columns []trace.RawColumn) {
	var err error

	var bytes []byte
	// Check file extension
	ext := path.Ext(filename)
	//
	switch ext {
	case ".json":
		js := json.ToJsonString(columns)
		//
		if err = os.WriteFile(filename, []byte(js), 0644); err == nil {
			return
		}
	case ".lt":
		bytes, err = lt.ToBytes(columns)
		//
		if err == nil {
			if err = os.WriteFile(filename, bytes, 0644); err == nil {
				return
			}
		}
	default:
		err = fmt.Errorf("Unknown trace file format: %s", ext)
	}
	// Handle error
	fmt.Println(err)
	os.Exit(4)
}

// Parse a trace file using a parser based on the extension of the filename.
func readTraceFile(filename string) []trace.RawColumn {
	var tr []trace.RawColumn
	// Read data file
	bytes, err := os.ReadFile(filename)
	// Check success
	if err == nil {
		// Check file extension
		ext := path.Ext(filename)
		//
		switch ext {
		case ".json":
			tr, err = json.FromBytes(bytes)
			if err == nil {
				return tr
			}
		case ".lt":
			tr, err = lt.FromBytes(bytes)
			if err == nil {
				return tr
			}
		default:
			err = fmt.Errorf("Unknown trace file format: %s", ext)
		}
	}
	// Handle error
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil
}

// Read the constraints file, whilst optionally including the standard library.
func readSchema(stdlib bool, debug bool, legacy bool, filenames []string) *hir.Schema {
	var err error
	//
	if len(filenames) == 0 {
		fmt.Println("source or binary constraint(s) file required.")
		os.Exit(5)
	} else if len(filenames) == 1 && path.Ext(filenames[0]) == ".bin" {
		// Single (binary) file supplied
		_, schema := readBinaryFile(legacy, filenames[0])
		// Ignore header information here.
		return schema
	}
	// Recursively expand any directories given in the list of filenames.
	if filenames, err = expandSourceFiles(filenames); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Must be source files
	return readSourceFiles(stdlib, debug, filenames)
}

// Parse a set of source files and compile them into a single schema.  This can
// result, for example, in a syntax error, etc.
func readSourceFiles(stdlib bool, debug bool, filenames []string) *hir.Schema {
	srcfiles := make([]*sexp.SourceFile, len(filenames))
	// Read each file
	for i, n := range filenames {
		log.Debug(fmt.Sprintf("including source file %s", n))
		// Read source file
		bytes, err := os.ReadFile(n)
		// Sanity check for errors
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		//
		srcfiles[i] = sexp.NewSourceFile(n, bytes)
	}
	// Parse and compile source files
	schema, errs := corset.CompileSourceFiles(stdlib, debug, srcfiles)
	// Check for any errors
	if len(errs) == 0 {
		return schema
	}
	// Report errors
	for _, err := range errs {
		printSyntaxError(&err)
	}
	// Fail
	os.Exit(4)
	// unreachable
	return nil
}

// Look through the list of filenames and identify any which are directories.
// Those are then recursively expanded.
func expandSourceFiles(filenames []string) ([]string, error) {
	var expandedFilenames []string
	//
	for _, f := range filenames {
		// Lookup information on the given file.
		if info, err := os.Stat(f); err != nil {
			// Something is wrong with one of the files provided, therefore
			// terminate with an error.
			return nil, err
		} else if info.IsDir() {
			// This a directory, so read its contents
			if contents, err := expandDirectory(f); err != nil {
				return nil, err
			} else {
				expandedFilenames = append(expandedFilenames, contents...)
			}
		} else {
			// This is a single file
			expandedFilenames = append(expandedFilenames, f)
		}
	}
	//
	return expandedFilenames, nil
}

// Recursively search through a given directory looking for any lisp files.
func expandDirectory(dirname string) ([]string, error) {
	var filenames []string
	// Recursively walk the given directory.
	err := filepath.Walk(dirname, func(filename string, info os.FileInfo, err error) error {
		if !info.IsDir() && path.Ext(filename) == ".lisp" {
			filenames = append(filenames, filename)
		}
		// Continue.
		return nil
	})
	// Done
	return filenames, err
}

// Print a syntax error with appropriate highlighting.
func printSyntaxError(err *sexp.SyntaxError) {
	span := err.Span()
	line := err.FirstEnclosingLine()
	lineOffset := span.Start() - line.Start()
	// Calculate length (ensures don't overflow line)
	length := min(line.Length()-lineOffset, span.Length())
	// Print error + line number
	fmt.Printf("%s:%d:%d-%d %s\n", err.SourceFile().Filename(),
		line.Number(), 1+lineOffset, 1+lineOffset+length, err.Message())
	// Print separator line
	fmt.Println()
	// Print line
	fmt.Println(line.String())
	// Print indent (todo: account for tabs)
	fmt.Print(strings.Repeat(" ", lineOffset))
	// Print highlight
	fmt.Println(strings.Repeat("^", length))
}

func maxHeightColumns(cols []trace.RawColumn) uint {
	h := uint(0)
	// Iterate over modules
	for _, col := range cols {
		h = max(h, col.Data.Len())
	}
	// Done
	return h
}

// ============================================================================
// Binary File Format
// ============================================================================

// BinaryFile provides a structured header for the binary file format.  In
// particular, it supports versioning and embedded (binary) metadata.
type BinaryFile struct {
	Identifier   [8]byte
	MajorVersion uint16
	MinorVersion uint16
	MetaData     []byte
}

// MarshalBinary converts the BinaryFile header into a sequence of bytes.
// Observe that we don't use GobEncoding here to avoid being tied to that
// encoding scheme.
func (p *BinaryFile) MarshalBinary() ([]byte, error) {
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
func (p *BinaryFile) UnmarshalBinary(buffer *bytes.Buffer) error {
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
func (p *BinaryFile) isCompatible() bool {
	return p.Identifier == ZKBINARY &&
		p.MajorVersion == BINFILE_MAJOR_VERSION &&
		p.MinorVersion <= BINFILE_MINOR_VERSION
}

// BINFILE_MAJOR_VERSION givesn the major version of the binary file format.  No
// matter what version, we should always have the ZKBINARY identifier first,
// followed by a GOB encoding of the header.  What follows after that, however,
// is determined by the major version.
const BINFILE_MAJOR_VERSION uint16 = 1

// BINFILE_MINOR_VERSION gives the minor version of the binary file format.  The
// expected interpretation is that older versions are compatible with newer
// ones, but not vice-versa.
const BINFILE_MINOR_VERSION uint16 = 0

// ZKBINARY is used as the file identifier for binary file types.  This just
// helps us identify actual binary files from corrupted files.
var ZKBINARY [8]byte = [8]byte{'z', 'k', 'b', 'i', 'n', 'a', 'r', 'y'}

// Read a "bin" file and extract the metadata bytes, along with the schema.
func readBinaryFile(legacy bool, filename string) (*BinaryFile, *hir.Schema) {
	var (
		header BinaryFile
		schema *hir.Schema
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
			// All looks good, proceed to decode the schema itself.
			err = decoder.Decode(&schema)
		} else if err == nil {
			err = fmt.Errorf("incompatible binary file (was v%d.%d, but expected v%d.%d)",
				header.MajorVersion, header.MinorVersion, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION)
		}
	}
	// Return if no errors
	if err == nil {
		return &header, schema
	}
	// Handle error & exit
	fmt.Println(err)
	os.Exit(2)
	// unreachable
	return nil, nil
}

// Write a binary file using a given set of metadata bytes.
//
//nolint:errcheck
func writeBinaryFile(metadata []byte, schema *hir.Schema, legacy bool, filename string) {
	var (
		buffer     bytes.Buffer
		gobEncoder *gob.Encoder = gob.NewEncoder(&buffer)
		// Construct header.
		header BinaryFile = BinaryFile{ZKBINARY, BINFILE_MAJOR_VERSION, BINFILE_MINOR_VERSION, metadata}
	)
	// Sanity checks
	if legacy {
		// Currently, there is no support for this.
		fmt.Println("legacy binary format not supported for writing")
	}
	// Marshal header
	headerBytes, _ := header.MarshalBinary()
	// Encode header
	buffer.Write(headerBytes)
	// Encode schema
	if err := gobEncoder.Encode(schema); err != nil {
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

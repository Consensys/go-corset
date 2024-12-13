package util

import (
	"bufio"
	"errors"
	"io"
	"os"
)

// ReadInputFile reads an input file as a sequence of lines.
func ReadInputFile(filename string) []string {
	file, err := os.Open(filename)
	// Check whether file exists
	if errors.Is(err, os.ErrNotExist) {
		return []string{}
	} else if err != nil {
		panic(err)
	}

	reader := bufio.NewReaderSize(file, 1024*128)
	lines := make([]string, 0)
	// Read file line-by-line
	for {
		// Read the next line
		line := readLine(reader)
		// Check whether for EOF
		if line == nil {
			if err = file.Close(); err != nil {
				panic(err)
			}

			return lines
		}

		lines = append(lines, *line)
	}
}

// Read a single line
func readLine(reader *bufio.Reader) *string {
	var (
		bytes []byte
		bit   []byte
		err   error
	)
	//
	cont := true
	//
	for cont {
		bit, cont, err = reader.ReadLine()
		if err == io.EOF {
			return nil
		} else if err != nil {
			panic(err)
		}

		bytes = append(bytes, bit...)
	}
	// Convert to string
	str := string(bytes)
	// Done
	return &str
}

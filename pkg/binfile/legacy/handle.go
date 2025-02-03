package binfile

import (
	"strconv"
	"strings"
)

// ============================================================================
// Column Ref
// ============================================================================

// Handle represents a module / column naming pair.
type Handle struct {
	module string
	column string
}

func asHandle(handle string) Handle {
	split := strings.Split(handle, ".")
	//
	if split[0] == "<prelude>" {
		return Handle{"", split[1]}
	}
	// Easy
	return Handle{split[0], split[1]}
}

func asColumn(handle string) uint {
	split := strings.Split(handle, "#")
	column, err := strconv.Atoi(split[1])
	// Error check
	if err != nil {
		panic(err.Error())
	}

	return uint(column)
}

func asColumns(handles []string) []uint {
	cols := make([]uint, len(handles))
	for i := 0; i < len(cols); i++ {
		cols[i] = asColumn(handles[i])
	}

	return cols
}

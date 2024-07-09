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
	split := strings.Split(handle, ":")
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

func asRegister(handle string, columns []column) uint {
	column := asColumn(handle)
	return columns[column].Register
}

func asRegisters(handles []string, columns []column) []uint {
	regs := make([]uint, len(handles))
	for i := 0; i < len(regs); i++ {
		regs[i] = asRegister(handles[i], columns)
	}

	return regs
}

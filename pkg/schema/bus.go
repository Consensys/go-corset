package schema

import (
	"github.com/consensys/go-corset/pkg/schema/agnostic"
	"github.com/consensys/go-corset/pkg/schema/module"
	"github.com/consensys/go-corset/pkg/schema/register"
)

// Bus describes an I/O bus referred to within a function.  Every function can
// connect with zero or more buses.  For example, making a function call
// requires a bus for the target function.  Each bus consists of some number of
// _address lines_ and some number of _data lines_.  Reading a value from the
// bus requires setting the address lines, then reading the data lines.
// Likewise, put a value onto the bus requires setting both the address and data
// lines.
type Bus interface {
	IsUnlinked() bool
	UnlinkedBus(name module.Name) Bus
	NewBus(name module.Name, id uint, address []register.Id, data []register.Id) Bus
	Address() []register.Id
	Data() []register.Id
	Split(mapping register.LimbsMap, _ agnostic.RegisterAllocator) Bus
	String() string
}


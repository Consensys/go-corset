package main


import (
	"fmt"
	"encoding/json"
	"os"
)

type RegisterID = interface{}
type Value struct {}
type Magma = interface{}
type Kind struct {
	m string
	c string
}
type Base = string
type Handle = string
type Register = interface{}
type FieldRegister = interface{}

type Column struct {
    Register int
    Shift int
    Padding_value Value
    Used bool
    Must_prove bool
    Kind string
    T Magma
    Intrinsic_size_factor string
    Base Base
    Gandle Handle
    Computed bool
}

type ColumnSet struct {
     // Raw array of column data, including virtial those which are
     // virtual and/or overlapping with others.
     Cols []Column `json:"_cols"`
     // Maps column handles to their index in the Cols array.
     ColsMap map[Handle]uint  `json:"cols"`
     // Maps column handles? to their length
     Effective_len map[string]int
     // min_len map[string]uint
     // field_registers []FieldRegister
     // registers []Register
     // spilling map[string]int    
}

type ConstraintSet struct {
    Columns ColumnSet
    // constraints []interface{}
    // constants interface{}
    // computations interface{}
    // perspectives []string
    // transformations uint64
    // auto_constraints uint64
}

func main() {
	file := os.Args[1]
	//
	fmt.Printf("Reading JSON bin file: %s\n",file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error")
	} else {
		var constraints ConstraintSet
		//var constraints map[string]interface{}
		err := json.Unmarshal(bytes, &constraints)
		if err != nil { fmt.Printf("Parsing JSON failed: %s",err) }
		cols := constraints
		fmt.Println(cols)
	}
}

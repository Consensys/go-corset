package main


import (
	"fmt"
	"os"
	"github.com/consensys/go-corset/pkg/binfile"
)


func main() {
	file := os.Args[1]
	//
	fmt.Printf("Reading JSON bin file: %s\n",file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error")
	} else {
		// Parse binary file into JSON form
		cs,_ := binfile.ConstraintSetFromJson(bytes)
		// Translate JSON form into MIR
		for i := 0; i < len(cs.Constraints); i++ {
			ith := cs.Constraints[i]
			hir := ith.Vanishes.Expr.ToHir()
			fmt.Println(hir)
		}
	}
}

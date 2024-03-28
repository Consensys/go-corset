package main


import (
	"fmt"
	"os"
	"github.com/Consensys/go-corset/pkg/binfile"
)


func main() {
	file := os.Args[1]
	//
	fmt.Printf("Reading JSON bin file: %s\n",file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error")
	} else {
		cs,_ := binfile.ConstraintSetFromJson(bytes)
		// do nothing for now
		fmt.Println(cs.Constraints)
	}
}

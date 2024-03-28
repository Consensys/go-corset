package main


import (
	"fmt"
	"os"
	"github.com/Consensys/go-corset/pkg/bin_fmt"
)


func main() {
	file := os.Args[1]
	//
	fmt.Printf("Reading JSON bin file: %s\n",file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error")
	} else {
		cs,_ := bin_fmt.ConstraintSetFromJson(bytes)
		// do nothing for now
		fmt.Println(cs.Constraints)
	}
}

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
		// Parse binary file into HIR schema
		schema,_ := binfile.HirSchemaFromJson(bytes)
		// Printout constraints
		for _,c := range schema.Constraints() {
			fmt.Println(c)
		}
	}
}

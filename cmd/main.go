package main


import (
	"fmt"
	"encoding/json"
	"os"
)

// type ConstraintSet struct {
// 	columns string
// 	constraints string
// 	constants string
// 	computations string
// 	perspectives []string
// 	transformations uint64
// 	auto_constraints uint64
// }

func main() {
	file := os.Args[1]
	//
	fmt.Printf("Reading JSON bin file: %s\n",file)
	bytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error")
	} else {
		//var constraints ConstraintSet
		var constraints map[string]interface{}
		err := json.Unmarshal(bytes, &constraints)
		if err != nil { fmt.Println(err) }
		fmt.Printf("%s",constraints)
	}
}

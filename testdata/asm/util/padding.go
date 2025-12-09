package main

import (
	"fmt"
	"math/big"
	"strings"
)

/**
// generate padding.accepts
go run testdata/asm/util/padding.go > byte_counts_test.accepts
cat byte_counts_test.accepts | shuf | tail -n 5000 > testdata/asm/util/padding.accepts
*/

func main() {
	for i := uint(0); i <= 16; i++ {
		for j := uint(0); j <= i; j++ {
			printInstance(i, j)
		}
	}
}

func printInstance(i uint, j uint) {
	var (
		builder strings.Builder
		input   big.Int
		res     big.Int
	)

	// input = 256^(i-1) + 256^(j-1)
	var a, b big.Int
	if i != 0 {
		a.SetBit(&a, int(8*(i-1)), 1)
	}
	if j != 0 {
		b.SetBit(&b, int(8*(j-1)), 1)
	}
	input.Add(&a, &b)

	// res = input >> (16 - i)
	res.Lsh(&input, 8*uint(16-i))

	builder.WriteString("{\"rpad_128\": { ")
	builder.WriteString(fmt.Sprintf("\"input\": [%s]", input.String()))
	builder.WriteString(fmt.Sprintf(", \"res\": [%s]", res.String()))
	builder.WriteString(" }}")
	fmt.Println(builder.String())

}

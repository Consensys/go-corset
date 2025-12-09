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
	for i := uint(0); i < 17; i++ {
		for j := uint(0); j < i+1; j++ {
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
	// input = 2^i + 2^j
	var a, b big.Int
	a.SetBit(&a, int(i), 1)
	b.SetBit(&b, int(j), 1)
	input.Add(&a, &b)

	// res = input >> (16 - i)
	res.Rsh(&input, uint(16-i))

	builder.WriteString("{\"rpad_128\": { ")
	builder.WriteString(fmt.Sprintf("\"input\": [%s]", input.String()))
	builder.WriteString(fmt.Sprintf(", \"res\": [%s]", res.String()))
	builder.WriteString(" }}")
	fmt.Println(builder.String())
}

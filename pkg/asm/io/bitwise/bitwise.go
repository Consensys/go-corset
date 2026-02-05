package bitwise

import "math/big"

func XOR(arg1, arg2 big.Int) big.Int {
	var out big.Int
	out.Xor(&arg1, &arg2)
	return out
}

func OR(arg1, arg2 big.Int) big.Int {
	var out big.Int
	out.Or(&arg1, &arg2)
	return out
}

func AND(arg1, arg2 big.Int) big.Int {
	var out big.Int
	out.And(&arg1, &arg2)
	return out
}

func NOT(arg1 big.Int) big.Int {
	var out big.Int
	out.Not(&arg1)
	return out
}

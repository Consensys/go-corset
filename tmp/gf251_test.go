package gf251

import (
	"fmt"
	"testing"
)

func TestReduce(t *testing.T) {
	//
	for i := uint8(0); i < uint8(251); i++ {
		lhs := New(i)
		//
		actual := lhs.ToByte()
		fmt.Print("{ \"reduction_u8\": { \"z\": ", lhs, ", \"RESULT\": [", actual, "] }}", "\n")
		//
		if actual != i {
			t.Errorf("matched %d\n", i)
		}
	}
}

func TestReduceSpecific(t *testing.T) {
	//
	lhs := New(5)
	//
	actual := lhs.ToByte()
	fmt.Print("{ \"reduction_u8\": { \"z\": ", lhs, ", \"RESULT\": [", actual, "] }}", "\n")
	//
}

func TestAdd(t *testing.T) {
	for i := uint32(0); i < N; i++ {
		for j := uint32(0); j < N; j++ {
			var (
				expected = uint8((i + j) % N)
				lhs      = New(uint8(i))
				rhs      = New(uint8(j))
			)
			//
			actual := lhs.Add(rhs).ToByte()

			fmt.Print("{ \"add\": { \"x\": ", lhs, ", \"y\": ", rhs, ", \"RESULT\": [", actual, "] }}", "\n")
			//
			if expected != actual {
				t.Errorf("*** %d + %d = %d (but expected %d)", i, j, actual, expected)
			}
		}
	}
}

func TestMul(t *testing.T) {
	for i := uint32(0); i < N; i++ {
		for j := i; j < N; j++ {
			var (
				expected = uint8((i * j) % N)
				lhs      = New(uint8(i))
				rhs      = New(uint8(j))
			)
			//
			actual := lhs.Mul(rhs).ToByte()
			fmt.Print("{ \"reduction_u8\": { \"z\": ", lhs.Mul(rhs), ", \"RESULT\": [", actual, "] }}", "\n")

			//
			if expected != actual {
				t.Errorf("*** %d * %d = %d (but expected %d)", i, j, actual, expected)
				//t.FailNow()
			}
		}
	}
}

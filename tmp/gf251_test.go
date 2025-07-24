package gf251

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestReduce(t *testing.T) {
	//
	for j := uint16(0); j <= uint16(65535); j++ {
		//lhs := New(i)
		//actual := lhs.ToByte()
		res := reduce(j)
		fmt.Print("i: ", j, " and ", res, "\n")
		//fmt.Print("{ \"mul\": { \"x\": [", i, "], \"y\": [", j, "], \"RESULT\": [", res, "] } }", "\n")
		//
		/*if actual != i {
			t.Errorf("matched %d\n", i)
		}*/

	}
}

func TestAdd(t *testing.T) {
	for i := uint32(100); i < 155; i++ {
		for j := uint32(0); j <= i; j++ {
			var (
				expected = uint8((i + j) % N)
				lhs      = New(uint8(i))
				rhs      = New(uint8(j))
			)

			//
			actual := lhs.Add(rhs).ToByte()

			fmt.Print("{ \"add\": { \"x\": [", i, "], \"y\": [", j, "], \"RESULT\": [", actual, "] } }", "\n")
			//
			if expected != actual {
				t.Errorf("*** %d + %d = %d (but expected %d)", i, j, actual, expected)
			}
		}
	}
}

func TestMul(t *testing.T) {
	for i := uint32(0); i < 256; i++ {
		for j := i; j < 256; j++ {
			var (
				expected = uint8((i * j) % N)
				lhs      = New(uint8(i))
				rhs      = New(uint8(j))
			)
			//
			actual := lhs.Mul(rhs).ToByte()

			// If the file doesn't exist, create it, or append to the file
			f, err := os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := f.Write([]byte("{ \"mul\": { \"x\": [" + strconv.Itoa(int(i)) + "], \"y\": [" + strconv.Itoa(int(j)) + "], \"RESULT\": [" + strconv.Itoa(int(actual)) + "] } }" + "\n")); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
			// fmt.Print("{ \"mul\": { \"x\": [", i, "], \"y\": [", j, "], \"RESULT\": [", actual, "] } }", "\n")
			//
			if expected != actual {
				t.Errorf("*** %d * %d = %d (but expected %d)", i, j, actual, expected)
				//t.FailNow()
			}
		}
	}
}

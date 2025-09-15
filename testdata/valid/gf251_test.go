package valid

import (
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/consensys/go-corset/tmp"
)

func TestReduce(t *testing.T) {
	//
	for i := uint8(0); i < uint8(251); i++ {
		lhs := gf251.New(i)
		//
		actual := lhs.ToByte()
		//
		if actual != i {
			t.Errorf("matched %d\n", i)
		}
	}
}

func TestAdd(t *testing.T) {
	for i := uint32(0); i < gf251.N; i++ {
		for j := uint32(0); j < gf251.N; j++ {
			var (
				expected = uint8((i + j) % gf251.N)
				lhs      = gf251.New(uint8(i))
				rhs      = gf251.New(uint8(j))
			)
			//
			actual := lhs.Add(rhs).ToByte()
			//
			if expected != actual {
				t.Errorf("*** %d + %d = %d (but expected %d)", i, j, actual, expected)
			}
		}
	}
}

func TestMul(t *testing.T) {
	for i := uint32(0); i < gf251.N; i++ {
		for j := i; j < gf251.N; j++ {
			var (
				expected = uint8((i * j) % gf251.N)
				lhs      = gf251.New(uint8(i))
				rhs      = gf251.New(uint8(j))
			)
			//
			actual := lhs.Mul(rhs).ToByte()
			// If the file doesn't exist, create it, or append to the file
			f, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := f.Write([]byte("{ \"mul\": { \"x\": [" + strconv.Itoa(int(i)) + "], \"y\": [" + strconv.Itoa(int(j)) + "], \"RESULT\": [" + strconv.Itoa(int(actual)) + "] } }\n")); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
			//
			if expected != actual {
				t.Errorf("*** %d * %d = %d (but expected %d)", i, j, actual, expected)
				//t.FailNow()
			}
		}
	}
}

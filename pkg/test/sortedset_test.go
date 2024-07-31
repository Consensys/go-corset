package test

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

func Test_SortedSet_00(t *testing.T) {
	check_SortedSet_Insert(t, 5, 10)
	check_SortedSet_InsertSorted(t, 5, 10)
}

func Test_SortedSet_01(t *testing.T) {
	// Really hammer it.
	for i := 0; i < 10000; i++ {
		check_SortedSet_Insert(t, 10, 32)
		check_SortedSet_InsertSorted(t, 10, 32)
	}
}

func Test_SortedSet_02(t *testing.T) {
	check_SortedSet_Insert(t, 100, 32)
	check_SortedSet_InsertSorted(t, 50, 32)
}

func Test_SortedSet_03(t *testing.T) {
	check_SortedSet_Insert(t, 1000, 64)
	check_SortedSet_InsertSorted(t, 500, 64)
}

func Test_SortedSet_04(t *testing.T) {
	check_SortedSet_Insert(t, 100000, 1024)
	check_SortedSet_InsertSorted(t, 50000, 1024)
}

func TestSlow_SortedSet_05(t *testing.T) {
	check_SortedSet_Insert(t, 100000, 4096)
	check_SortedSet_InsertSorted(t, 50000, 4096)
}

func TestSlow_SortedSet_06(t *testing.T) {
	check_SortedSet_Insert(t, 100000, 16384)
	check_SortedSet_InsertSorted(t, 50000, 16384)
}

// ===================================================================
// Test Helpers
// ===================================================================

func array_contains(items []uint, element uint) bool {
	for _, e := range items {
		if e == element {
			return true
		}
	}
	// Not present
	return false
}

func check_SortedSet_Insert(t *testing.T, n uint, m uint) {
	items := GenerateRandomInputs(n, m)
	set := toSortedSet(items)

	for i := uint(0); i < m; i++ {
		l := array_contains(items, i)
		r := set.Contains(i)

		if !l && r {
			t.Errorf("unexpected item %d", i)
		} else if l && !r {
			t.Errorf("missing item %d", i)
		}
	}
}

func check_SortedSet_InsertSorted(t *testing.T, n uint, m uint) {
	left := GenerateRandomInputs(n, m)
	right := GenerateRandomInputs(n, m)
	set := toSortedSet(left)
	set.InsertSorted(toSortedSet(right))
	//
	for i := uint(0); i < m; i++ {
		l := array_contains(left, i) || array_contains(right, i)
		r := set.Contains(i)

		if !l && r {
			t.Errorf("unexpected item %d", i)
		} else if l && !r {
			t.Errorf("missing item %d", i)
		}
	}
}

func toSortedSet(items []uint) *util.SortedSet[uint] {
	set := util.NewSortedSet[uint]()
	for _, v := range items {
		set.Insert(v)
	}

	return set
}

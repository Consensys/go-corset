package bit

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

func Test_BitSet_00(t *testing.T) {
	check_BitSet_Insert(t, 5, 10)
}

func Test_BitSet_01(t *testing.T) {
	// Really hammer it.
	for i := 0; i < 10000; i++ {
		check_BitSet_Insert(t, 10, 128)
	}
}

func Test_BitSet_02(t *testing.T) {
	check_BitSet_Insert(t, 100, 256)
}

func Test_BitSet_03(t *testing.T) {
	check_BitSet_Insert(t, 1000, 512)
}

func Test_BitSet_04(t *testing.T) {
	check_BitSet_Insert(t, 100000, 1024)
}

func TestSlow_BitSet_05(t *testing.T) {
	check_BitSet_Insert(t, 100000, 4096)
}

func TestSlow_BitSet_06(t *testing.T) {
	check_BitSet_Insert(t, 100000, 16384)
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

func countUniqueItems(items []uint) uint {
	count := uint(0)
	counts := make(map[uint]bool)
	//
	for _, val := range items {
		if _, ok := counts[val]; !ok {
			count++
			counts[val] = true
		}
	}
	//
	return count
}

func check_BitSet_Insert(t *testing.T, n uint, m uint) {
	var iset Set
	//
	items := util.GenerateRandomUints(n, m)
	count := countUniqueItems(items)
	bset := toBitSet(items)
	iset.InsertAll(bset)
	//
	if bset.Count() != count {
		t.Errorf("unexpected number of items (%d vs %d) (insert)", bset.Count(), count)
	}
	//
	if iset.Count() != count {
		t.Errorf("unexpected number of items (%d vs %d) (insert all)", iset.Count(), count)
	}
	//
	for i := uint(0); i < m; i++ {
		l := array_contains(items, i)
		r := bset.Contains(i)
		s := iset.Contains(i)
		// Check set
		if !l && r {
			t.Errorf("unexpected item %d (insert)", i)
		} else if l && !r {
			t.Errorf("missing item %d (insert)", i)
		}
		// Check iset
		if !l && s {
			t.Errorf("unexpected item %d (insert all)", i)
		} else if l && !s {
			t.Errorf("missing item %d (insert all)", i)
		}
	}
}

func toBitSet(items []uint) Set {
	set := Set{}
	for _, v := range items {
		set.Insert(v)
	}

	return set
}

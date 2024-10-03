package test

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

func Test_Enumerator_1_1(t *testing.T) {
	enumerator := util.EnumerateElements[uint](1, []uint{0})
	checkEnumerator(t, enumerator, [][]uint{{0}}, arrayEquals)
}

func Test_Enumerator_1_2(t *testing.T) {
	enumerator := util.EnumerateElements[uint](1, []uint{0, 1})
	checkEnumerator(t, enumerator, [][]uint{{0}, {1}}, arrayEquals)
}

func Test_Enumerator_1_3(t *testing.T) {
	enumerator := util.EnumerateElements[uint](1, []uint{0, 1, 2})
	checkEnumerator(t, enumerator, [][]uint{{0}, {1}, {2}}, arrayEquals)
}

func Test_Enumerator_2_1(t *testing.T) {
	enumerator := util.EnumerateElements[uint](2, []uint{0})
	checkEnumerator(t, enumerator, [][]uint{{0, 0}}, arrayEquals)
}

func Test_Enumerator_2_2(t *testing.T) {
	enumerator := util.EnumerateElements[uint](2, []uint{0, 1})
	checkEnumerator(t, enumerator, [][]uint{{0, 0}, {1, 0}, {0, 1}, {1, 1}}, arrayEquals)
}

func Test_Enumerator_2_3(t *testing.T) {
	enumerator := util.EnumerateElements[uint](2, []uint{0, 1, 2})
	checkEnumerator(t, enumerator, [][]uint{
		{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}, {2, 1}, {0, 2}, {1, 2}, {2, 2}}, arrayEquals)
}

func Test_Enumerator_3_1(t *testing.T) {
	enumerator := util.EnumerateElements[uint](3, []uint{0})
	checkEnumerator(t, enumerator, [][]uint{{0, 0, 0}}, arrayEquals)
}
func Test_Enumerator_3_2(t *testing.T) {
	enumerator := util.EnumerateElements[uint](3, []uint{0, 1})
	checkEnumerator(t, enumerator, [][]uint{
		{0, 0, 0}, {1, 0, 0}, {0, 1, 0}, {1, 1, 0}, {0, 0, 1}, {1, 0, 1}, {0, 1, 1}, {1, 1, 1}}, arrayEquals)
}

// ===================================================================
// Test Helpers
// ===================================================================

func checkEnumerator[E any](t *testing.T, enumerator util.Enumerator[E], expected []E, eq func(E, E) bool) {
	for i := 0; i < len(expected); i++ {
		ith := enumerator.Next()
		if !eq(ith, expected[i]) {
			t.Errorf("expected %s, got %s", any(expected[i]), any(ith))
		}
	}
	// Sanity check lengths match
	if enumerator.HasNext() {
		t.Errorf("expected %d elements, got more", len(expected))
	}
}

func arrayEquals[T comparable](lhs []T, rhs []T) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	// Check each item in turn
	for i := 0; i < len(lhs); i++ {
		if lhs[i] != rhs[i] {
			return false
		}
	}
	// Done
	return true
}

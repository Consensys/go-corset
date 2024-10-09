package test

import (
	"reflect"
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

func Test_RemoveMatching_01(t *testing.T) {
	check_RemoveMatching(t, []int{1, 2}, 1, []int{2})
}

func Test_RemoveMatching_02(t *testing.T) {
	check_RemoveMatching(t, []int{1, 2}, 2, []int{1})
}

func Test_RemoveMatching_03(t *testing.T) {
	check_RemoveMatching(t, []int{1, 2, 3}, 1, []int{2, 3})
}

func Test_RemoveMatching_04(t *testing.T) {
	check_RemoveMatching(t, []int{2, 1, 3}, 1, []int{2, 3})
}

func Test_RemoveMatching_05(t *testing.T) {
	check_RemoveMatching(t, []int{2, 3, 1}, 1, []int{2, 3})
}

func Test_RemoveMatching_06(t *testing.T) {
	check_RemoveMatching(t, []int{1, 2, 3, 1}, 1, []int{2, 3})
}

func Test_RemoveMatching_07(t *testing.T) {
	check_RemoveMatching(t, []int{2, 1, 3, 1}, 1, []int{2, 3})
}
func Test_RemoveMatching_08(t *testing.T) {
	check_RemoveMatching(t, []int{2, 3, 1, 1}, 1, []int{2, 3})
}

func Test_RemoveMatching_09(t *testing.T) {
	check_RemoveMatching(t, []int{1, 2, 1, 3}, 1, []int{2, 3})
}

func Test_RemoveMatching_10(t *testing.T) {
	check_RemoveMatching(t, []int{1, 1, 2, 3}, 1, []int{2, 3})
}

func check_RemoveMatching(t *testing.T, original []int, item int, expected []int) {
	actual := util.RemoveMatching(original, func(ith int) bool { return ith == item })
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("removing %d from %v got %v", item, original, actual)
	}
}

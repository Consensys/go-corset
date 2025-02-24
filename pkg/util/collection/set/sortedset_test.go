// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package set

import (
	"fmt"
	"strings"
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

const N = 2

type entry struct {
	values [N]uint
}

func (lhs entry) LessEq(rhs entry) bool {
	for i := len(lhs.values); i > 0; i-- {
		//sign := (i == len(lhs.values)) || (i == 1)
		sign := true
		//
		if lhs.values[i-1] < rhs.values[i-1] {
			return sign
		} else if lhs.values[i-1] > rhs.values[i-1] {
			return !sign
		}
	}
	//
	return true
}

func zip(items [N][]uint) []entry {
	var entries []entry

	for i := range items[0] {
		var row [N]uint
		for j := range items {
			row[j] = items[j][i]
		}

		entries = append(entries, entry{row})
	}
	//
	return entries
}

func unzip(entries []entry) [N][]uint {
	var items [N][]uint
	//
	for _, e := range entries {
		for i := 0; i < N; i++ {
			items[i] = append(items[i], e.values[i])
		}
	}
	//
	return items
}

func initWords(n uint, widths []uint) [N][]uint {
	var words [N][]uint
	//
	for k := 0; k < N; k++ {
		if widths[k] != 0 {
			words[k] = util.GenerateRandomUints(n, widths[k])
		} else {
			words[k] = make([]uint, n)
		}
	}
	//
	return words
}

func sortWords(words [N][]uint) [N][]uint {
	// Sort it
	aset := NewAnySortedSet[entry]()
	//
	for _, v := range zip(words) {
		aset.Insert(v)
	}
	// Unzip it
	return unzip(aset.ToArray())
}

func areSorted(words [N][]uint) bool {
	ws := zip(words)
	//
	for i := range ws {
		if i > 0 && !ws[i-1].LessEq(ws[i]) {
			return false
		}
	}
	//
	return true
}

func hasDeltaOverflow(words [N][]uint, bounds []uint) bool {
	nrows := len(words[0])
	for k := 0; k < nrows; k++ {
		if k == 0 {
			continue
		}
		//
		for i := N; i > 0; i-- {
			items := words[i-1]
			// NOTE: assumes positive sign here.
			delta := items[k] - items[k-1]
			//
			if delta >= bounds[i-1] {
				return true
			} else if delta != 0 {
				break
			}
		}
	}
	//
	return false
}

func Test_SortedSet(t *testing.T) {
	// Bounds determined by bitwidth of columns.
	bounds := []uint{262144, 262144}
	init := []uint{300000, 1024}
	//init := []uint{256, 256}
	// //
	for i := 3; i < 10; i++ {
		for j := 0; j < 100000; j++ {
			// Create words
			words := initWords(uint(i), init)
			// Sort words
			words = sortWords(words)
			//
			//if words[N-1][0] != 0 {
			//if !areSorted(words) {
			if hasDeltaOverflow(words, bounds) {
				// Print words
				fmt.Printf("{ ")

				for k := 0; k < N; k++ {
					if k != 0 {
						fmt.Print(", ")
					}

					kth := fmt.Sprintf("%v", words[k])
					kth = strings.ReplaceAll(kth, " ", ",")
					fmt.Printf("\"W%d\": %s", k, kth)
				}

				fmt.Printf("}\n")
			}
		}
	}
}

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
	items := util.GenerateRandomUints(n, m)
	aset := toSortedSet(items)
	anyset := toAnySortedSet(items)

	for i := uint(0); i < m; i++ {
		l := array_contains(items, i)
		r := aset.Contains(i)
		// Check set
		if !l && r {
			t.Errorf("unexpected item %d", i)
		} else if l && !r {
			t.Errorf("missing item %d", i)
		}
		// Check anyset
		r = anyset.Contains(Order[uint]{Item: i})
		if !l && r {
			t.Errorf("unexpected item %d (any)", i)
		} else if l && !r {
			t.Errorf("missing item %d (any)", i)
		}
	}
}

func check_SortedSet_InsertSorted(t *testing.T, n uint, m uint) {
	left := util.GenerateRandomUints(n, m)
	right := util.GenerateRandomUints(n, m)
	aset := toSortedSet(left)
	anyset := toAnySortedSet(left)

	aset.InsertSorted(toSortedSet(right))
	anyset.InsertSorted(toAnySortedSet(right))
	//
	for i := uint(0); i < m; i++ {
		l := array_contains(left, i) || array_contains(right, i)
		r := aset.Contains(i)
		// Check set
		if !l && r {
			t.Errorf("unexpected item %d", i)
		} else if l && !r {
			t.Errorf("missing item %d", i)
		}
		// Check any set
		r = anyset.Contains(Order[uint]{Item: i})
		if !l && r {
			t.Errorf("unexpected item %d (any)", i)
		} else if l && !r {
			t.Errorf("missing item %d (any)", i)
		}
	}
}

func toSortedSet(items []uint) *SortedSet[uint] {
	set := NewSortedSet[uint]()
	for _, v := range items {
		set.Insert(v)
	}

	return set
}

func toAnySortedSet(items []uint) *AnySortedSet[Order[uint]] {
	aset := NewAnySortedSet[Order[uint]]()
	for _, v := range items {
		aset.Insert(Order[uint]{Item: v})
	}

	return aset
}

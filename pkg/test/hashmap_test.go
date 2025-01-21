package test

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util"
)

func Test_HashMap_01(t *testing.T) {
	items := []uint{1, 2, 3, 4, 3, 2, 1}
	check_HashMap(t, items)
}

func Test_HashMap_02(t *testing.T) {
	items := util.GenerateRandomUints(10, 32)
	check_HashMap(t, items)
}

func Test_HashMap_03(t *testing.T) {
	items := util.GenerateRandomUints(100, 32)
	check_HashMap(t, items)
}

func Test_HashMap_04(t *testing.T) {
	items := util.GenerateRandomUints(1000, 32)
	check_HashMap(t, items)
}

func Test_HashMap_05(t *testing.T) {
	items := util.GenerateRandomUints(100000, 32)
	check_HashMap(t, items)
}

func TestSlow_HashMap_08(t *testing.T) {
	items := util.GenerateRandomUints(100000, 64)
	check_HashMap(t, items)
}

func TestSlow_HashMap_09(t *testing.T) {
	items := util.GenerateRandomUints(100000, 128)
	check_HashMap(t, items)
}

// ===================================================================
// Test Helpers
// ===================================================================

func check_HashMap(t *testing.T, items []uint) {
	gmap := initGoMap(items)
	hmap := util.NewHashMap[testKey, uint](0)
	// Insert items
	for key, val := range gmap {
		hmap.Insert(testKey{key}, val)
	}
	// Sanity check number of unique items
	if hmap.Size() != uint(len(gmap)) {
		t.Errorf("expected %d items, got %d: %s", len(gmap), hmap.Size(), hmap.String())
	}
	// Sanity check containership
	for key, val := range gmap {
		if !hmap.ContainsKey(testKey{key}) {
			t.Errorf("missing key %d: %s", key, hmap.String())
		} else if v, ok := hmap.Get(testKey{key}); !ok {
			t.Errorf("missing item %d=>%d: %s", key, val, hmap.String())
		} else if v != val {
			t.Errorf("expecting %d=>%d, got %d=>%d: %s", key, val, key, v, hmap.String())
		}
	}
}

func initGoMap(items []uint) map[uint]uint {
	gmap := make(map[uint]uint)
	//
	for _, v := range items {
		if w, ok := gmap[v]; ok {
			gmap[v] = w + 1
		} else {
			gmap[v] = 1
		}
	}
	//
	return gmap
}

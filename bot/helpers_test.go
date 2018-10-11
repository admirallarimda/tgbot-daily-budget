package bot

import "testing"

func TestUniqueInts(t *testing.T) {
	empty := make([]int, 0)
	empty = uniqueInts(empty)
	if len(empty) != 0 {
		t.Error(empty)
	}

	unique := []int{7, 2, 4, 8}
	unique = uniqueInts(unique)
	if len(unique) != 4 {
		t.Error(unique)
	}

	allSame := []int{5, 5, 5, 5, 5}
	allSame = uniqueInts(allSame)
	if len(allSame) != 1 {
		t.Error(allSame)
	}

	mixed := []int{5, 17, 4, 17, 2, 5, 1, 2, 2, 6}
	mixed = uniqueInts(mixed)
	if len(mixed) != 6 {
		t.Error(mixed)
	}
}

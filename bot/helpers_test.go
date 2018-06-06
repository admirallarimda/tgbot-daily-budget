package bot

import "testing"

func TestUniqueInts(t *testing.T) {
    var null []int = nil
    uniqueInts(null)

    empty := make([]int, 0)
    uniqueInts(empty)
    if len(empty) != 0 {
        t.Log(empty)
    }

    unique := []int {7, 2, 4, 8}
    uniqueInts(unique)
    if len(unique) != 4 {
        t.Log(unique)
    }

    allSame := []int {5, 5, 5, 5, 5}
    uniqueInts(allSame)
    if len(allSame) != 1 {
        t.Log(allSame)
    }

    mixed := []int {5, 17, 4, 17, 2, 5, 1, 2, 2, 6}
    uniqueInts(mixed)
    if len(mixed) != 5 {
        t.Log(mixed)
    }
}

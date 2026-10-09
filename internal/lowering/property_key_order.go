package lowering

import (
	"sort"
	"strconv"

	"github.com/pilotworks/scriptgo/internal/ir"
)

// orderPropertyKeys returns the order of fields as OrdinaryOwnPropertyKeys
// lists them: array-index keys ("0", "16") in ascending numeric order, then
// the other keys in their declaration order. The result maps each output
// position to an input index.
func orderPropertyKeys(fields []ir.Field) []int {
	order := make([]int, len(fields))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		ia, aIndex := arrayIndexKey(fields[order[a]].Name)
		ib, bIndex := arrayIndexKey(fields[order[b]].Name)
		if aIndex && bIndex {
			return ia < ib
		}
		return aIndex && !bIndex
	})
	return order
}

// arrayIndexKey reports a canonical array index key: decimal digits without
// a leading zero, below 2^32 - 1.
func arrayIndexKey(key string) (uint64, bool) {
	if key == "" || (len(key) > 1 && key[0] == '0') {
		return 0, false
	}
	n, err := strconv.ParseUint(key, 10, 64)
	if err != nil || n >= 1<<32-1 {
		return 0, false
	}
	return n, true
}

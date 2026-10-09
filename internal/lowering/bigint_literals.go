package lowering

import (
	"math/big"
	"strings"
)

// bigIntLiteralValue normalizes a BigInt literal (decimal, 0x, 0o, 0b, with
// numeric separators and an optional n suffix) to the decimal text used by
// typed IR. Native bigint is a signed 64-bit integer, so literals outside
// that range are rejected rather than wrapped.
func bigIntLiteralValue(text string) (string, bool) {
	digits := strings.TrimSuffix(strings.TrimSpace(text), "n")
	value, ok := new(big.Int).SetString(digits, 0)
	if !ok || !value.IsInt64() {
		return "", false
	}
	return value.String(), true
}

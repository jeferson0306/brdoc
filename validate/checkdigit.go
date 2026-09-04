package validate

// Most Brazilian document numbers end in check digits computed the same way: a
// weighted sum of the preceding digits, taken modulo 11. Only the weights
// differ. Keeping that arithmetic in one place is what stops each new document
// from arriving with its own copy of the same loop — and its own chance of
// getting the edge case wrong.

// mod11Digit returns the check digit for digits weighted by weights.
//
// The remainder maps to 11 minus itself, except that 10 and 11 both mean zero —
// a check digit is a single character, so the two remainders that would need two
// are folded onto 0. CPF, CNPJ, PIS and título de eleitor all share this rule.
func mod11Digit(digits string, weights []int) int {
	sum := 0
	for i, weight := range weights {
		sum += int(digits[i]-'0') * weight
	}

	if digit := 11 - sum%11; digit < 10 {
		return digit
	}
	return 0
}

// allSameDigit reports a run of one repeated digit.
//
// Every mod-11 document has repeated-digit values that satisfy the arithmetic —
// 111.111.111-11 computes correctly — so they have to be excluded by name
// rather than by the check digits.
func allSameDigit(digits string) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			return false
		}
	}
	return len(digits) > 0
}

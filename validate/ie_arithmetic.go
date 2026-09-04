package validate

import "strings"

// The arithmetic every state registration is built from.
//
// Twenty-seven states, twenty-seven roteiros de crítica — but almost all of
// them are the same three steps with different constants: take a prefix of the
// digits, multiply by descending weights, and fold the remainder into a check
// digit. Keeping those steps here is what lets most states be four lines
// instead of forty.

// descendingWeights produces the weight series the SINTEGRA documents print
// above the digits: counting down from start, and wrapping back to 9 after 2.
//
// descendingWeights(4, 11) is 4 3 2 9 8 7 6 5 4 3 2 — the series Acre's roteiro
// shows, and the reason the wrap exists at all.
func descendingWeights(start, size int) []int {
	weights := make([]int, 0, size)

	for len(weights) < size {
		weights = append(weights, start)
		if start == 2 {
			start = 10
		}
		start--
	}

	return weights
}

// weightedTotal multiplies each digit by its weight and sums, stopping at
// whichever of the two runs out first.
func weightedTotal(digits string, weights []int) int {
	total := 0

	for i := 0; i < len(weights) && i < len(digits); i++ {
		total += int(digits[i]-'0') * weights[i]
	}

	return total
}

// mod11Check is the rule most states share: eleven minus the remainder, with
// remainders of 0 and 1 both yielding zero because the alternative needs two
// characters.
func mod11Check(total int) int {
	if remainder := total % 11; remainder >= 2 {
		return 11 - remainder
	}
	return 0
}

// mod10Check is the same idea in base ten, used by Bahia for the registrations
// whose leading digit puts them in the module-10 group.
func mod10Check(total int) int {
	if remainder := total % 10; remainder != 0 {
		return 10 - remainder
	}
	return 0
}

// mod11TimesTen covers the states whose roteiro multiplies the sum by ten
// before taking the remainder — Alagoas and Rio Grande do Norte — where a
// remainder of 10 yields zero.
func mod11TimesTen(total int) int {
	if remainder := total * 10 % 11; remainder != 10 {
		return remainder
	}
	return 0
}

// hasPrefix reports whether the registration starts with any of the state
// codes given. An empty list means the state does not fix its opening digits.
func hasPrefix(ie string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(ie, prefix) {
			return true
		}
	}

	return false
}

// digitsOnly reports whether every character is a digit, which every state
// requires except São Paulo, whose rural registrations begin with a P.
func digitsOnly(ie string) bool {
	for i := 0; i < len(ie); i++ {
		if ie[i] < '0' || ie[i] > '9' {
			return false
		}
	}
	return len(ie) > 0
}

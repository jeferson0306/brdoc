package utils

import "regexp"

var (
	cnhChars     = regexp.MustCompile(`^[\d.\-\s]*$`)
	renavamChars = regexp.MustCompile(`^[\d.\-\s]*$`)
)

// ValidateCNH checks a driving licence number.
//
// The two check digits are unusual in being coupled: when the first overflows
// past nine it is written as zero, and the second is then computed two lower to
// compensate. Treating them independently — which is what most of the wrong
// implementations of this do — accepts numbers the Detran would reject.
func ValidateCNH(cnh string) (bool, string, string) {
	if !cnhChars.MatchString(cnh) {
		return false, cnh, "Invalid CNH format (unexpected characters)"
	}

	digits := nonDigitsRegex.ReplaceAllString(cnh, "")
	if len(digits) != 11 {
		return false, digits, "Invalid CNH format (incorrect length)"
	}
	if allSameDigit(digits) {
		return false, digits, "Invalid CNH format"
	}

	first, sum := 0, 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (9 - i)
	}

	first = sum % 11
	decrement := 0
	if first >= 10 {
		first, decrement = 0, 2
	}
	if first != int(digits[9]-'0') {
		return false, digits, "Invalid CNH format"
	}

	sum = 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (i + 1)
	}

	second := sum%11 - decrement
	if second < 0 {
		second += 11
	}
	if second >= 10 {
		second = 0
	}
	if second != int(digits[10]-'0') {
		return false, digits, "Invalid CNH format"
	}

	return true, digits, "Valid CNH format"
}

// ValidateRenavam checks a vehicle registration number.
//
// Nine digits was the old format and eleven is the current one; the shorter is
// the longer with leading zeros, so both are accepted and normalised to eleven.
// The weights run from the right and cycle two through nine, unlike the
// left-to-right descending series most Brazilian documents use.
func ValidateRenavam(renavam string) (bool, string, string) {
	if !renavamChars.MatchString(renavam) {
		return false, renavam, "Invalid RENAVAM format (unexpected characters)"
	}

	digits := nonDigitsRegex.ReplaceAllString(renavam, "")
	if len(digits) != 9 && len(digits) != 11 {
		return false, digits, "Invalid RENAVAM format (incorrect length)"
	}

	for len(digits) < 11 {
		digits = "0" + digits
	}

	// Not in the reference this was checked against, but consistent with every
	// other document here: a run of one repeated digit is not a registration.
	if allSameDigit(digits) {
		return false, digits, "Invalid RENAVAM format"
	}

	sum, weight := 0, 2
	for i := 9; i >= 0; i-- {
		sum += int(digits[i]-'0') * weight
		if weight >= 9 {
			weight = 2
		} else {
			weight++
		}
	}

	expected := 11 - sum%11
	if expected >= 10 {
		expected = 0
	}

	if expected != int(digits[10]-'0') {
		return false, digits, "Invalid RENAVAM format"
	}

	return true, digits, "Valid RENAVAM format"
}

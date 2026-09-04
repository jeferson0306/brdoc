package validate

import "regexp"

var boletoChars = regexp.MustCompile(`^[\d.\-\s]*$`)

// A linha digitável is not the barcode. It is the barcode's 44 digits shuffled
// into three groups that each carry their own module-10 check digit, so a
// person typing it from paper is told which group they got wrong rather than
// simply that the whole number is bad.
//
// Validating it therefore means two separate things: check each group, then put
// the barcode back together and check its own digit.

// boletoGroups are the three typed groups: where each starts, where its digits
// end, and where its check digit sits.
var boletoGroups = []struct{ start, end, check int }{
	{0, 9, 9},
	{10, 20, 20},
	{21, 31, 31},
}

// barcodeOrder reassembles the 44-digit barcode out of the typed line. The
// pieces are not in the same order in both, which is the whole reason this
// mapping has to exist.
var barcodeOrder = [][2]int{{0, 4}, {32, 47}, {4, 9}, {10, 20}, {21, 31}}

// barcodeCheckDigitPosition is where the general check digit sits once the
// barcode is reassembled — inside it, not at the end.
const barcodeCheckDigitPosition = 4

// checkBoleto checks the 47-digit linha digitável of a bank slip.
func checkBoleto(boleto string) (bool, string, string) {
	if !boletoChars.MatchString(boleto) {
		return false, boleto, "Invalid boleto (unexpected characters)"
	}

	digits := nonDigitsRegex.ReplaceAllString(boleto, "")

	// Utility and tax slips are 48 digits with a different structure entirely.
	// Saying so is not the same as calling the number invalid: it has not been
	// checked, and reporting it as invalid would be a lie about a slip that may
	// well be good.
	if len(digits) == 48 {
		return false, digits, "Arrecadação slips (48 digits) are not supported yet — this number was not checked"
	}

	if len(digits) != 47 {
		return false, digits, "Invalid boleto (expected 47 digits)"
	}

	for i, group := range boletoGroups {
		expected := mod10Boleto(digits[group.start:group.end])
		if expected != int(digits[group.check]-'0') {
			return false, digits, "Invalid boleto (check digit of group " + string(rune('1'+i)) + " does not match)"
		}
	}

	barcode := ""
	for _, span := range barcodeOrder {
		barcode += digits[span[0]:span[1]]
	}

	withoutCheck := barcode[:barcodeCheckDigitPosition] + barcode[barcodeCheckDigitPosition+1:]
	if mod11Boleto(withoutCheck) != int(barcode[barcodeCheckDigitPosition]-'0') {
		return false, digits, "Invalid boleto (general check digit does not match)"
	}

	return true, digits, "Valid boleto"
}

// mod10Boleto is the Luhn-shaped check the typed groups use: alternating
// weights of two and one from the right, with any product above nine reduced by
// nine.
func mod10Boleto(digits string) int {
	sum := 0

	for i := len(digits) - 1; i >= 0; i-- {
		weight := 2
		if (len(digits)-1-i)&1 == 1 {
			weight = 1
		}

		product := int(digits[i]-'0') * weight
		if product > 9 {
			product -= 9
		}
		sum += product
	}

	if remainder := sum % 10; remainder > 0 {
		return 10 - remainder
	}
	return 0
}

// mod11Boleto is the barcode's own check: weights cycling two through nine from
// the right, where a remainder of 0 or 1 yields 1 rather than the 0 most
// Brazilian documents use.
func mod11Boleto(digits string) int {
	sum, weight := 0, 2

	for i := len(digits) - 1; i >= 0; i-- {
		sum += int(digits[i]-'0') * weight
		if weight < 9 {
			weight++
		} else {
			weight = 2
		}
	}

	if remainder := sum % 11; remainder == 0 || remainder == 1 {
		return 1
	} else {
		return 11 - remainder
	}
}

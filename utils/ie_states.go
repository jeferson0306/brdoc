package utils

import "strconv"

// One entry per state. The common shape — a fixed length, an optional state
// prefix, one trailing check digit over descending weights — is expressed as
// data; the states that genuinely do something else get a function.

type ieValidator func(ie string) bool

// trailingMod11 builds the shape most states use: `length` digits, optionally
// opening with one of `prefixes`, ending in a single mod-11 check digit taken
// over weights descending from `weightStart`.
func trailingMod11(length, weightStart int, prefixes ...string) ieValidator {
	return func(ie string) bool {
		if len(ie) != length || !hasPrefix(ie, prefixes) {
			return false
		}

		base := ie[:length-1]
		total := weightedTotal(base, descendingWeights(weightStart, length-1))
		return int(ie[length-1]-'0') == mod11Check(total)
	}
}

// twoTrailingMod11 builds the thirteen-digit shape used by Acre and the Federal
// District: two check digits, the second computed over the first.
func twoTrailingMod11(prefixes ...string) ieValidator {
	return func(ie string) bool {
		if len(ie) != 13 || !hasPrefix(ie, prefixes) {
			return false
		}

		base := ie[:11]
		first := mod11Check(weightedTotal(base, descendingWeights(4, 11)))
		second := mod11Check(weightedTotal(base+strconv.Itoa(first), descendingWeights(5, 12)))

		return ie == base+strconv.Itoa(first)+strconv.Itoa(second)
	}
}

var ieValidators = map[string]ieValidator{
	// The plain nine-digit, weights nine down to two, mod-11 shape.
	"AM": trailingMod11(9, 9),
	"CE": trailingMod11(9, 9, "06"),
	"ES": trailingMod11(9, 9),
	"MA": trailingMod11(9, 9, "12"),
	"MS": trailingMod11(9, 9, "28"),
	"PA": trailingMod11(9, 9, "15"),
	"PB": trailingMod11(9, 9),
	"PI": trailingMod11(9, 9),
	"SC": trailingMod11(9, 9),
	"SE": trailingMod11(9, 9),

	// Same idea, other lengths.
	"MT": trailingMod11(11, 3),
	"RS": trailingMod11(10, 2),

	"AC": twoTrailingMod11("01"),
	"DF": twoTrailingMod11(),

	"AL": validateAlagoas,
	"AP": validateAmapa,
	"BA": validateBahia,
	"GO": validateGoias,
	"MG": validateMinasGerais,
	"PE": validatePernambuco,
	"PR": validateParana,
	"RJ": validateRioDeJaneiro,
	"RN": validateRioGrandeDoNorte,
	"RO": validateRondonia,
	"RR": validateRoraima,
	"SP": validateSaoPaulo,
	"TO": validateTocantins,
}

// Alagoas multiplies the weighted sum by ten before the remainder, and a
// remainder of ten means zero.
func validateAlagoas(ie string) bool {
	if len(ie) != 9 || !hasPrefix(ie, []string{"24"}) {
		return false
	}

	total := weightedTotal(ie[:8], descendingWeights(9, 8))
	return int(ie[8]-'0') == mod11TimesTen(total)
}

// Amapá adds a constant to the sum, and forces the digit for two registration
// ranges, according to where the number falls in the state's own sequence.
func validateAmapa(ie string) bool {
	if len(ie) != 9 || !hasPrefix(ie, []string{"03"}) {
		return false
	}

	value, err := strconv.Atoi(ie[:8])
	if err != nil {
		return false
	}

	shift, forced := 0, -1
	switch {
	case value >= 3000001 && value <= 3017000:
		shift, forced = 5, 0
	case value >= 3017001 && value <= 3019022:
		shift, forced = 9, 1
	}

	total := shift + weightedTotal(ie[:8], descendingWeights(9, 8))
	digit := 11 - total%11

	switch {
	case digit == 10:
		digit = 0
	case digit == 11:
		digit = forced
		if forced < 0 {
			digit = 0
		}
	}

	return int(ie[8]-'0') == digit
}

// Bahia is the only state where the check digits are stored in the opposite
// order to the one they are computed in: the second digit sits before the
// first. Which module applies depends on the leading digit — the first for
// eight-digit registrations, the second for nine-digit ones.
func validateBahia(ie string) bool {
	if len(ie) != 8 && len(ie) != 9 {
		return false
	}

	baseLength := len(ie) - 2
	base := ie[:baseLength]

	leading := ie[0]
	if len(ie) == 9 {
		leading = ie[1]
	}

	module11 := true
	for _, d := range []byte{'0', '1', '2', '3', '4', '5', '8'} {
		if leading == d {
			module11 = false
			break
		}
	}

	check := mod10Check
	if module11 {
		check = mod11Check
	}

	second := check(weightedTotal(base, descendingWeights(baseLength+1, baseLength)))
	first := check(weightedTotal(base+strconv.Itoa(second), descendingWeights(baseLength+2, baseLength+1)))

	return ie == base+strconv.Itoa(first)+strconv.Itoa(second)
}

// Goiás forces the digit for one historical registration and for a range that
// was issued before the current rule.
func validateGoias(ie string) bool {
	if len(ie) != 9 {
		return false
	}
	if !hasPrefix(ie, []string{"10", "11", "15"}) {
		leading, err := strconv.Atoi(ie[:2])
		if err != nil || leading < 20 || leading > 29 {
			return false
		}
	}

	base := ie[:8]
	if base == "11094402" {
		return ie[8] == '0' || ie[8] == '1'
	}

	total := weightedTotal(base, descendingWeights(9, 8))
	remainder := total % 11

	digit := 11 - remainder
	switch {
	case remainder == 0:
		digit = 0
	case remainder == 1:
		value, err := strconv.Atoi(base)
		if err != nil {
			return false
		}
		digit = 0
		if value >= 10103105 && value <= 10119997 {
			digit = 1
		}
	}

	return int(ie[8]-'0') == digit
}

// Minas Gerais inserts a zero after the municipality code, then takes a digit
// sum rather than a weighted sum for the first check digit.
func validateMinasGerais(ie string) bool {
	if len(ie) != 13 {
		return false
	}

	base := ie[:11]
	expanded := base[:3] + "0" + base[3:]

	sum := 0
	for i := 0; i < len(expanded); i++ {
		product := int(expanded[i]-'0') * (1 + i%2)
		sum += product / 10
		sum += product % 10
	}

	first := (10 - sum%10) % 10
	second := mod11Check(weightedTotal(base+strconv.Itoa(first), []int{3, 2, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2}))

	return ie == base+strconv.Itoa(first)+strconv.Itoa(second)
}

// Pernambuco: seven digits of sequence and two check digits.
func validatePernambuco(ie string) bool {
	if len(ie) != 9 {
		return false
	}

	base := ie[:7]
	first := mod11Check(weightedTotal(base, descendingWeights(8, 7)))
	second := mod11Check(weightedTotal(base+strconv.Itoa(first), descendingWeights(9, 8)))

	return ie == base+strconv.Itoa(first)+strconv.Itoa(second)
}

// Paraná: eight digits and two check digits, with weight series that restart
// rather than descend continuously.
func validateParana(ie string) bool {
	if len(ie) != 10 {
		return false
	}

	base := ie[:8]
	first := mod11Check(weightedTotal(base, append([]int{3, 2}, descendingWeights(7, 6)...)))
	second := mod11Check(weightedTotal(base+strconv.Itoa(first), append([]int{4, 3, 2}, descendingWeights(7, 6)...)))

	return ie == base+strconv.Itoa(first)+strconv.Itoa(second)
}

// Rio de Janeiro is the shortest: eight digits, and the weight series opens
// with 2 before descending from 7.
func validateRioDeJaneiro(ie string) bool {
	if len(ie) != 8 {
		return false
	}

	base := ie[:7]
	return int(ie[7]-'0') == mod11Check(weightedTotal(base, []int{2, 7, 6, 5, 4, 3, 2}))
}

// Rio Grande do Norte accepts nine or ten digits and, like Alagoas, multiplies
// before taking the remainder.
func validateRioGrandeDoNorte(ie string) bool {
	if len(ie) != 9 && len(ie) != 10 || !hasPrefix(ie, []string{"20"}) {
		return false
	}

	base := ie[:len(ie)-1]

	weights := descendingWeights(9, 8)
	if len(ie) == 10 {
		weights = append([]int{10}, weights...)
	}

	return int(ie[len(ie)-1]-'0') == mod11TimesTen(weightedTotal(base, weights))
}

// Rondônia changed format in 2000: the older nine-digit registrations open with
// a three-digit municipality code that is excluded from the calculation, while
// the current fourteen-digit ones dropped it and pad the company number with
// leading zeros instead. The state's own document says the old check digit
// stays valid under the new formula, so both use the same rule.
//
// That rule is not the usual one. Where most states fold a difference of 10 or
// 11 onto zero, Rondônia's roteiro says to *subtract ten* — so a sum divisible
// by eleven yields 1, not 0.
func validateRondonia(ie string) bool {
	switch len(ie) {
	case 9:
		return int(ie[8]-'0') == rondoniaDigit(weightedTotal(ie[3:8], descendingWeights(6, 5)))
	case 14:
		return int(ie[13]-'0') == rondoniaDigit(weightedTotal(ie[:13], descendingWeights(6, 13)))
	default:
		return false
	}
}

func rondoniaDigit(total int) int {
	if digit := 11 - total%11; digit < 10 {
		return digit
	} else {
		return digit - 10
	}
}

// Roraima is the only state on module nine, with ascending weights.
func validateRoraima(ie string) bool {
	if len(ie) != 9 || !hasPrefix(ie, []string{"24"}) {
		return false
	}

	total := weightedTotal(ie[:8], []int{1, 2, 3, 4, 5, 6, 7, 8})
	return int(ie[8]-'0') == total%9
}

// São Paulo has two formats: twelve digits for industry and commerce, and a
// rural producer number opening with P whose single check digit sits in the
// middle rather than at the end.
func validateSaoPaulo(ie string) bool {
	weights := []int{1, 3, 4, 5, 6, 7, 8, 10}

	if len(ie) > 0 && (ie[0] == 'P' || ie[0] == 'p') {
		if len(ie) != 13 || !digitsOnly(ie[1:]) {
			return false
		}

		base := ie[1:9]
		digit := weightedTotal(base, weights) % 11 % 10
		return int(ie[9]-'0') == digit
	}

	if len(ie) != 12 {
		return false
	}

	first := weightedTotal(ie[:8], weights) % 11 % 10
	if int(ie[8]-'0') != first {
		return false
	}

	second := weightedTotal(ie[:11], []int{3, 2, 10, 9, 8, 7, 6, 5, 4, 3, 2}) % 11 % 10
	return int(ie[11]-'0') == second
}

// Tocantins carried a two-digit type code in the middle of its older
// eleven-digit registrations, which is excluded from the calculation.
func validateTocantins(ie string) bool {
	if len(ie) != 9 && len(ie) != 11 {
		return false
	}

	base := ie[:len(ie)-1]
	if len(ie) == 11 {
		switch ie[2:4] {
		case "01", "02", "03", "99":
		default:
			return false
		}
		base = ie[:2] + ie[4:10]
	}

	return int(ie[len(ie)-1]-'0') == mod11Check(weightedTotal(base, descendingWeights(9, 8)))
}

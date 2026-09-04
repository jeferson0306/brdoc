package utils

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	pisChars    = regexp.MustCompile(`^[\d.\-\s]*$`)
	tituloChars = regexp.MustCompile(`^[\d.\-\s]*$`)

	// Old pattern: three letters, four digits. Mercosul: the fifth character
	// became a letter. Both are in circulation and both are legal.
	plateOld      = regexp.MustCompile(`^[A-Z]{3}[0-9]{4}$`)
	plateMercosul = regexp.MustCompile(`^[A-Z]{3}[0-9][A-Z][0-9]{2}$`)
	plateChars    = regexp.MustCompile(`^[A-Za-z0-9\-\s]*$`)

	// A PIX random key is a UUID. Accepted in either case, with or without
	// braces stripped by the caller.
	pixEVP = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// E.164 with the Brazilian country code: +55, two-digit area code, then
	// eight or nine digits.
	pixPhone = regexp.MustCompile(`^\+55\d{10,11}$`)
)

// ValidatePIS checks a PIS/PASEP/NIT/NIS number.
//
// All four names refer to the same eleven-digit number with the same check
// digit, so one validator covers them; the API exposes it as "pis".
func ValidatePIS(pis string) (bool, string, string) {
	if !pisChars.MatchString(pis) {
		return false, pis, "Invalid PIS format (unexpected characters)"
	}

	sanitized := nonDigitsRegex.ReplaceAllString(pis, "")
	if len(sanitized) != 11 {
		return false, sanitized, "Invalid PIS format (incorrect length)"
	}
	if allSameDigit(sanitized) {
		return false, sanitized, "Invalid PIS format"
	}

	weights := []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	if mod11Digit(sanitized[:10], weights) != int(sanitized[10]-'0') {
		return false, sanitized, "Invalid PIS format"
	}

	return true, sanitized, "Valid PIS format"
}

// ValidateTituloEleitor checks a voter registration number.
//
// The number is an eight-digit sequence, two digits identifying the issuing
// state, and two check digits. Three details make it unlike the other
// documents here:
//
//   - The check digits are the remainder itself, not eleven minus it.
//   - São Paulo (01) and Minas Gerais (02) outgrew the eight-digit sequence, so
//     their numbers carry thirteen digits rather than twelve. The state code is
//     therefore located from the end, never at a fixed offset.
//   - For those same two states a remainder of zero yields 1, not 0.
func ValidateTituloEleitor(titulo string) (bool, string, string) {
	if !tituloChars.MatchString(titulo) {
		return false, titulo, "Invalid voter ID format (unexpected characters)"
	}

	digits := nonDigitsRegex.ReplaceAllString(titulo, "")
	if len(digits) < 12 {
		return false, digits, "Invalid voter ID format (incorrect length)"
	}

	union := digits[len(digits)-4 : len(digits)-2]
	extended := union == "01" || union == "02"

	if len(digits) != 12 && !(len(digits) == 13 && extended) {
		return false, digits, "Invalid voter ID format (incorrect length)"
	}

	state, err := strconv.Atoi(union)
	if err != nil || state < 1 || state > 28 {
		return false, digits, "Invalid voter ID format (unknown issuing state)"
	}

	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(digits[i]-'0') * (i + 2)
	}
	first := voterDigit(sum, extended)

	sum = int(union[0]-'0')*7 + int(union[1]-'0')*8 + first*9
	second := voterDigit(sum, extended)

	if digits[len(digits)-2:] != strconv.Itoa(first)+strconv.Itoa(second) {
		return false, digits, "Invalid voter ID format"
	}

	return true, digits, "Valid voter ID format"
}

// voterDigit maps a weighted sum to a voter-ID check digit: the remainder
// itself, with 10 folded onto 0, and zero raised to 1 for the two states whose
// numbering was extended.
func voterDigit(sum int, extendedState bool) int {
	remainder := sum % 11

	if remainder == 0 && extendedState {
		return 1
	}
	if remainder == 10 {
		return 0
	}
	return remainder
}

// ValidatePlate checks a vehicle plate in either the old or the Mercosul
// pattern. There is no check digit — a plate is a format, not a computation.
func ValidatePlate(plate string) (bool, string, string) {
	if !plateChars.MatchString(plate) {
		return false, plate, "Invalid plate format (unexpected characters)"
	}

	sanitized := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(plate))

	switch {
	case plateMercosul.MatchString(sanitized):
		return true, sanitized, "Valid plate format (Mercosul)"
	case plateOld.MatchString(sanitized):
		return true, sanitized, "Valid plate format (pre-Mercosul)"
	default:
		return false, sanitized, "Invalid plate format"
	}
}

// ValidateDocument accepts whichever of CPF or CNPJ the digits describe.
//
// Forms routinely offer a single "document" field, and asking the caller to
// decide which endpoint to call defeats the point of that field.
func ValidateDocument(document string) (bool, string, string) {
	digits := nonDigitsRegex.ReplaceAllString(document, "")

	switch len(digits) {
	case 11:
		isValid, sanitized, _ := ValidateCPF(document)
		if isValid {
			return true, sanitized, "Valid document (CPF)"
		}
		return false, sanitized, "Invalid document (CPF check failed)"
	case 14:
		isValid, sanitized, _ := ValidateCNPJ(document)
		if isValid {
			return true, sanitized, "Valid document (CNPJ)"
		}
		return false, sanitized, "Invalid document (CNPJ check failed)"
	default:
		return false, digits, "Invalid document (expected 11 digits for CPF or 14 for CNPJ)"
	}
}

// ValidatePixKey checks a PIX key in any of the five forms the arrangement
// defines: CPF, CNPJ, email, phone in E.164, or a random UUID.
//
// The form is inferred rather than asked for, because that is how a key arrives
// — pasted into one field, with nothing saying which kind it is.
func ValidatePixKey(key string) (bool, string, string) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return false, key, "Invalid PIX key (empty)"
	}

	if pixEVP.MatchString(trimmed) {
		return true, strings.ToLower(trimmed), "Valid PIX key (random)"
	}
	if pixPhone.MatchString(trimmed) {
		return true, trimmed, "Valid PIX key (phone)"
	}
	if strings.Contains(trimmed, "@") {
		if isValid, sanitized, _ := ValidateEmail(trimmed); isValid {
			return true, sanitized, "Valid PIX key (email)"
		}
		return false, trimmed, "Invalid PIX key (malformed email)"
	}

	digits := nonDigitsRegex.ReplaceAllString(trimmed, "")
	switch len(digits) {
	case 11:
		if isValid, sanitized, _ := ValidateCPF(trimmed); isValid {
			return true, sanitized, "Valid PIX key (CPF)"
		}
		return false, digits, "Invalid PIX key (CPF check failed)"
	case 14:
		if isValid, sanitized, _ := ValidateCNPJ(trimmed); isValid {
			return true, sanitized, "Valid PIX key (CNPJ)"
		}
		return false, digits, "Invalid PIX key (CNPJ check failed)"
	}

	return false, trimmed, "Invalid PIX key (not a CPF, CNPJ, email, phone or random key)"
}

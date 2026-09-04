package validate

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	nonDigitsRegex = regexp.MustCompile(`\D`)
	emailRegex     = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	nameRegex      = regexp.MustCompile(`[^\p{L}\s\-']`)
)

// Characters a person legitimately types *around* the value. Anything outside
// these sets means the caller sent something that is not the document at all.
//
// Stripping every non-digit and then validating what remains accepts
// "abc529.982.247-25" and "529.982.247-25jasasas" as valid CPFs, because the
// letters disappear before the check digits are computed. An API that answers
// "valid" to those lets junk into the caller's database.
//
// These guard the *decision*. Sanitisation still happens afterwards for values
// that pass, so "529.982.247-25" is still normalised to "52998224725".
var (
	cpfChars   = regexp.MustCompile(`^[\d.\-\s]*$`)
	cnpjChars  = regexp.MustCompile(`^[\d.\-/\s]*$`)
	rgChars    = regexp.MustCompile(`^[\d.\-\s]*[xX]?$`)
	cepChars   = regexp.MustCompile(`^[\d\-\s]*$`)
	phoneChars = regexp.MustCompile(`^[\d+().\-\s]*$`)
	cardChars  = regexp.MustCompile(`^[\d\-\s]*$`)
	nameChars  = regexp.MustCompile(`^[\p{L}\s\-'.]*$`)
)

// checkCPF checks the format of a CPF number.
func checkCPF(cpf string) (bool, string, string) {
	if !cpfChars.MatchString(cpf) {
		return false, cpf, "Invalid CPF format (unexpected characters)"
	}

	sanitizedCPF := nonDigitsRegex.ReplaceAllString(cpf, "")
	if len(sanitizedCPF) != 11 || !isValidCPF(sanitizedCPF) {
		return false, sanitizedCPF, "Invalid CPF format"
	}
	return true, sanitizedCPF, "Valid CPF format"
}

// isValidCPF validates a CPF number according to CPF rules.
func isValidCPF(cpf string) bool {
	if cpf == "00000000000" || cpf == "11111111111" ||
		cpf == "22222222222" || cpf == "33333333333" ||
		cpf == "44444444444" || cpf == "55555555555" ||
		cpf == "66666666666" || cpf == "77777777777" ||
		cpf == "88888888888" || cpf == "99999999999" {
		return false
	}

	for i := 9; i < 11; i++ {
		sum := 0
		for j := 0; j < i; j++ {
			num := int(cpf[j] - '0')
			sum += num * (i + 1 - j)
		}
		result := (sum * 10) % 11
		if result == 10 {
			result = 0
		}
		if result != int(cpf[i]-'0') {
			return false
		}
	}
	return true
}

// checkCNPJ checks the format and both check digits of a CNPJ.
//
// Ported from the shared TypeScript validators so the two cannot disagree; the
// weights and the "remainder below 2 means zero" rule come from the Receita
// Federal specification.
func checkCNPJ(cnpj string) (bool, string, string) {
	if !cnpjChars.MatchString(cnpj) {
		return false, cnpj, "Invalid CNPJ format (unexpected characters)"
	}

	sanitizedCNPJ := nonDigitsRegex.ReplaceAllString(cnpj, "")
	if len(sanitizedCNPJ) != 14 {
		return false, sanitizedCNPJ, "Invalid CNPJ format (incorrect length)"
	}

	// A CNPJ of one repeated digit passes the check-digit arithmetic, so it has
	// to be excluded explicitly — 00.000.000/0000-00 is the classic example.
	if allSameDigit(sanitizedCNPJ) {
		return false, sanitizedCNPJ, "Invalid CNPJ format"
	}

	firstWeights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	secondWeights := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	if mod11Digit(sanitizedCNPJ[:12], firstWeights) != int(sanitizedCNPJ[12]-'0') ||
		mod11Digit(sanitizedCNPJ[:13], secondWeights) != int(sanitizedCNPJ[13]-'0') {
		return false, sanitizedCNPJ, "Invalid CNPJ format"
	}

	return true, sanitizedCNPJ, "Valid CNPJ format"
}

// checkEmail validates if the email format is correct.
func checkEmail(email string) (bool, string, string) {
	sanitizedEmail := strings.TrimSpace(strings.ToLower(email))
	if emailRegex.MatchString(sanitizedEmail) {
		return true, sanitizedEmail, "Valid email format"
	}
	return false, sanitizedEmail, "Invalid email format"
}

// checkRG validates the format of a Brazilian RG (Registro Geral).
func checkRG(rg string) (bool, string, string) {
	if !rgChars.MatchString(rg) {
		return false, rg, "Invalid RG format (unexpected characters)"
	}

	sanitizedRG := nonDigitsRegex.ReplaceAllString(rg, "")

	if len(sanitizedRG) < 7 || len(sanitizedRG) > 9 {
		return false, sanitizedRG, "Invalid RG format (incorrect length)"
	}

	return true, sanitizedRG, "Valid RG format"
}

// checkCEP validates a Brazilian postal code (CEP).
func checkCEP(cep string) (bool, string, string) {
	if !cepChars.MatchString(cep) {
		return false, cep, "Invalid CEP format (unexpected characters)"
	}

	sanitizedCEP := nonDigitsRegex.ReplaceAllString(cep, "")

	if len(sanitizedCEP) != 8 {
		return false, sanitizedCEP, "Invalid CEP format (incorrect length)"
	}

	if sanitizedCEP == "00000000" {
		return false, sanitizedCEP, "Invalid CEP format"
	}

	return true, sanitizedCEP, "Valid CEP format"
}

// checkName sanitizes and validates the name format.
func checkName(name string) (bool, string, string) {
	rawName := name
	if !nameChars.MatchString(name) {
		return false, rawName, "Invalid name format (unexpected characters)"
	}

	sanitizedName := removeAccents(name)

	sanitizedName = nameRegex.ReplaceAllString(sanitizedName, "")
	sanitizedName = strings.Join(strings.Fields(sanitizedName), " ")
	sanitizedName = strings.ToUpper(strings.TrimSpace(sanitizedName))

	if len(sanitizedName) < 3 {
		return false, rawName, "Invalid name format (too short)"
	}
	if len(sanitizedName) > 60 {
		return false, rawName, "Invalid name format (too long)"
	}

	if strings.Contains(sanitizedName, "  ") {
		return false, rawName, "Invalid name format (contains multiple spaces)"
	}

	return true, sanitizedName, "Valid name format"
}

// removeAccents removes accents from characters.
func removeAccents(input string) string {
	decomposed := norm.NFD.String(input)
	var output strings.Builder
	for _, char := range decomposed {
		if unicode.Is(unicode.Mn, char) {
			continue
		}
		output.WriteRune(char)
	}
	return output.String()
}

// checkPhone validates and sanitizes Brazilian phone numbers.
func checkPhone(phone string) (bool, string, string) {
	if !phoneChars.MatchString(phone) {
		return false, phone, "Invalid phone format (unexpected characters)"
	}

	sanitizedPhone := nonDigitsRegex.ReplaceAllString(phone, "")

	if strings.HasPrefix(sanitizedPhone, "55") && (len(sanitizedPhone) == 12 || len(sanitizedPhone) == 13) {
		sanitizedPhone = sanitizedPhone[2:]
	}

	if len(sanitizedPhone) != 10 && len(sanitizedPhone) != 11 {
		return false, sanitizedPhone, "Invalid phone format"
	}

	ddd := sanitizedPhone[:2]
	if ddd < "11" || ddd > "99" {
		return false, sanitizedPhone, "Invalid phone format (invalid DDD)"
	}

	if len(sanitizedPhone) == 11 && sanitizedPhone[2] != '9' {
		return false, sanitizedPhone, "Invalid phone format (mobile numbers must start with 9)"
	}

	if len(sanitizedPhone) == 10 && sanitizedPhone[2] == '0' {
		return false, sanitizedPhone, "Invalid phone format"
	}

	return true, sanitizedPhone, "Valid phone format"
}

// checkPlastic validates a credit card number using the Luhn Algorithm and determines its brand.
func checkPlastic(cardNumber string) (bool, string, string) {
	if !cardChars.MatchString(cardNumber) {
		return false, cardNumber, "Invalid credit card number (unexpected characters)"
	}

	sanitizedCardNumber := nonDigitsRegex.ReplaceAllString(cardNumber, "")

	if len(sanitizedCardNumber) < 13 || len(sanitizedCardNumber) > 19 {
		return false, sanitizedCardNumber, "Invalid credit card number (incorrect length)"
	}

	if !isValidLuhn(sanitizedCardNumber) {
		return false, sanitizedCardNumber, "Invalid credit card number"
	}

	brand := cardBrandOf(sanitizedCardNumber)
	return true, sanitizedCardNumber, "Valid credit card number (" + brand + ")"
}

// cardBrandOf identifies the card brand based on the BIN.
func cardBrandOf(cardNumber string) string {
	cardNumber = strings.ReplaceAll(cardNumber, " ", "") // Remove spaces

	if len(cardNumber) >= 2 {
		switch {
		case strings.HasPrefix(cardNumber, "34") || strings.HasPrefix(cardNumber, "37"):
			return "American Express"
		case strings.HasPrefix(cardNumber, "36"):
			return "Diners Club"
		case strings.HasPrefix(cardNumber, "54") || strings.HasPrefix(cardNumber, "55") ||
			(cardNumber[:2] >= "51" && cardNumber[:2] <= "55"):
			return "MasterCard"
		case strings.HasPrefix(cardNumber, "4"):
			return "Visa"
		case len(cardNumber) >= 4 && (strings.HasPrefix(cardNumber, "6011") ||
			(cardNumber[:3] >= "644" && cardNumber[:3] <= "649") || strings.HasPrefix(cardNumber, "65")):
			return "Discover"
		case len(cardNumber) >= 4 && (strings.HasPrefix(cardNumber, "5067") || strings.HasPrefix(cardNumber, "4576") ||
			strings.HasPrefix(cardNumber, "4011")):
			return "Elo"
		case len(cardNumber) >= 6 && (strings.HasPrefix(cardNumber, "384100") || strings.HasPrefix(cardNumber, "384140") ||
			strings.HasPrefix(cardNumber, "384160") || strings.HasPrefix(cardNumber, "606282") || strings.HasPrefix(cardNumber, "637095")):
			return "Hipercard"
		}
	}

	return "Unknown"
}

// isValidLuhn implements the Luhn Algorithm to verify credit card numbers.
func isValidLuhn(cardNumber string) bool {
	var sum int
	alt := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n := int(cardNumber[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}

	return sum%10 == 0
}

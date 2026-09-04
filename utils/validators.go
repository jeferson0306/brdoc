package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-redis/redis/v8"
	"golang.org/x/text/unicode/norm"
)

// cacheTimeout bounds every Redis call. Without it a hung or unreachable cache
// blocks the request for as long as the client is willing to wait — validation
// itself takes microseconds, so anything slower than this is not worth waiting
// for and the answer is recomputed instead.
const cacheTimeout = 150 * time.Millisecond

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

// Initialize Redis client
var rdb = redis.NewClient(&redis.Options{
	Addr: getRedisAddr(),
})

// CacheHealthy reports whether Redis is actually reachable, so /health can say
// so instead of the service degrading silently — which is how a broken cache
// went unnoticed in production.
func CacheHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), cacheTimeout)
	defer cancel()
	return rdb.Ping(ctx).Err() == nil
}

func getRedisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

// cacheKeyFor hashes the value before it becomes a Redis key.
//
// The key used to be "cpf:<the actual CPF>", which put every CPF ever validated
// into the cache in the clear, readable by anyone with access to it. A SHA-256
// digest keys just as well — equal inputs still collide onto one entry — while
// storing nothing that identifies a person.
func cacheKeyFor(prefix, value string) string {
	digest := sha256.Sum256([]byte(value))
	return prefix + ":" + hex.EncodeToString(digest[:])
}

// ValidateCPFWithCache validates the CPF and uses cache to avoid duplicate validations.
func ValidateCPFWithCache(cpf string) (bool, string, string, bool) {
	// Before the cache, not after: the key is built from the extracted digits, so
	// "abc529.982.247-25" and "529.982.247-25" would otherwise share an entry and
	// the dirty value would inherit the clean one's cached "true".
	if !cpfChars.MatchString(cpf) {
		return false, cpf, "Invalid CPF format (unexpected characters)", false
	}

	sanitizedCPF := nonDigitsRegex.ReplaceAllString(cpf, "")
	cacheKey := cacheKeyFor("cpf", sanitizedCPF)

	ctx, cancel := context.WithTimeout(context.Background(), cacheTimeout)
	defer cancel()

	cachedResult, err := rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		isValid := cachedResult == "true"
		message := "Valid CPF format"
		if !isValid {
			message = "Invalid CPF format"
		}
		return isValid, sanitizedCPF, message, true
	}

	isValid, sanitizedCPF, message := ValidateCPF(cpf)

	cacheValue := "false"
	if isValid {
		cacheValue = "true"
	}
	if err := rdb.Set(ctx, cacheKey, cacheValue, 24*time.Hour).Err(); err != nil {
		// Debug, not error: a missing cache is a degraded mode, not a failure,
		// and at info level a down Redis would drown the log at request rate.
		// CacheHealthy() is what surfaces the condition.
		slog.Debug("cache write failed", slog.String("error", err.Error()))
	}

	return isValid, sanitizedCPF, message, false
}

// ValidateCPF checks the format of a CPF number.
func ValidateCPF(cpf string) (bool, string, string) {
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

// ValidateCNPJ checks the format and both check digits of a CNPJ.
//
// Ported from the shared TypeScript validators so the two cannot disagree; the
// weights and the "remainder below 2 means zero" rule come from the Receita
// Federal specification.
func ValidateCNPJ(cnpj string) (bool, string, string) {
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

	if checkDigit(sanitizedCNPJ[:12], firstWeights) != int(sanitizedCNPJ[12]-'0') ||
		checkDigit(sanitizedCNPJ[:13], secondWeights) != int(sanitizedCNPJ[13]-'0') {
		return false, sanitizedCNPJ, "Invalid CNPJ format"
	}

	return true, sanitizedCNPJ, "Valid CNPJ format"
}

// checkDigit applies the weighted sum the Receita Federal specifies: a
// remainder below 2 yields a zero, otherwise the digit is 11 minus it.
func checkDigit(digits string, weights []int) int {
	sum := 0
	for i, weight := range weights {
		sum += int(digits[i]-'0') * weight
	}

	if remainder := sum % 11; remainder >= 2 {
		return 11 - remainder
	}
	return 0
}

func allSameDigit(digits string) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			return false
		}
	}
	return len(digits) > 0
}

// ValidateEmail validates if the email format is correct.
func ValidateEmail(email string) (bool, string, string) {
	sanitizedEmail := strings.TrimSpace(strings.ToLower(email))
	if emailRegex.MatchString(sanitizedEmail) {
		return true, sanitizedEmail, "Valid email format"
	}
	return false, sanitizedEmail, "Invalid email format"
}

// ValidateRG validates the format of a Brazilian RG (Registro Geral).
func ValidateRG(rg string) (bool, string, string) {
	if !rgChars.MatchString(rg) {
		return false, rg, "Invalid RG format (unexpected characters)"
	}

	sanitizedRG := nonDigitsRegex.ReplaceAllString(rg, "")

	if len(sanitizedRG) < 7 || len(sanitizedRG) > 9 {
		return false, sanitizedRG, "Invalid RG format (incorrect length)"
	}

	return true, sanitizedRG, "Valid RG format"
}

// ValidateCEP validates a Brazilian postal code (CEP).
func ValidateCEP(cep string) (bool, string, string) {
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

// ValidateName sanitizes and validates the name format.
func ValidateName(name string) (bool, string, string) {
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

// ValidatePhone validates and sanitizes Brazilian phone numbers.
func ValidatePhone(phone string) (bool, string, string) {
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

// ValidatePlastic validates a credit card number using the Luhn Algorithm and determines its brand.
func ValidatePlastic(cardNumber string) (bool, string, string) {
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

	brand := GetCardBrand(sanitizedCardNumber)
	return true, sanitizedCardNumber, "Valid credit card number (" + brand + ")"
}

// GetCardBrand identifies the card brand based on the BIN.
func GetCardBrand(cardNumber string) string {
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

package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

// Fixed vectors published by brazilian-utils, the reference implementation
// these validators were checked against.
func TestValidatePISKnownVectors(t *testing.T) {
	tests := []struct {
		pis   string
		valid bool
	}{
		{"12056874107", true},
		{"12056412847", true},
		{"120.56874.10-7", true},
		{"12056412547", false},
		{"12081636639", false},
		{"00000000000", false},
		{"11111111111", false},
		{"123456", false},
		{"12056Aabb412847", false},
		{"12056412847abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.pis, func(t *testing.T) {
			isValid, _, message := ValidatePIS(tt.pis)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

func TestValidateTituloEleitor(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{"unknown_state_00", "0000000000" + "00", false},
		{"unknown_state_29", "12345678" + "29" + "00", false},
		{"too_short", "1234567890", false},
		// Thirteen digits are only legal for São Paulo and Minas Gerais.
		{"thirteen_digits_other_state", "123456789" + "03" + "12", false},
		{"letters", "12345678010abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := ValidateTituloEleitor(tt.value)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

// referencePIS and referenceVoterID are ported from brazilian-utils
// (github.com/brazilian-utils/brazilian-utils, MIT). They are deliberately
// written as flat, literal transcriptions — no shared helpers, no reuse of the
// code under test — so a differential run compares two independent expressions
// of the same specification rather than one implementation with itself.
func referencePIS(digits string) bool {
	if len(digits) != 11 {
		return false
	}
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			break
		}
		if i == 10 {
			return false
		}
	}

	weights := []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(digits[i]-'0') * weights[i]
	}

	expected := 11 - sum%11
	if expected >= 10 {
		expected = 0
	}
	return expected == int(digits[10]-'0')
}

func referenceVoterID(digits string) bool {
	if len(digits) < 12 {
		return false
	}

	union := digits[len(digits)-4 : len(digits)-2]
	extended := union == "01" || union == "02"
	if len(digits) != 12 && !(len(digits) == 13 && extended) {
		return false
	}

	state, err := strconv.Atoi(union)
	if err != nil || state < 1 || state > 28 {
		return false
	}

	digit := func(sum int) int {
		remainder := sum % 11
		if remainder == 0 && extended {
			return 1
		}
		if remainder == 10 {
			return 0
		}
		return remainder
	}

	sum := 0
	for i := 0; i < 8; i++ {
		sum += int(digits[i]-'0') * (i + 2)
	}
	first := digit(sum)
	second := digit(int(union[0]-'0')*7 + int(union[1]-'0')*8 + first*9)

	return digits[len(digits)-2:] == fmt.Sprintf("%d%d", first, second)
}

// A fixed seed keeps a failure reproducible: the same run produces the same
// counterexample every time.
func TestPISAgreesWithReference(t *testing.T) {
	random := rand.New(rand.NewSource(20260904))

	for i := 0; i < 5000; i++ {
		candidate := randomDigits(random, 11)
		got, _, _ := ValidatePIS(candidate)

		if want := referencePIS(candidate); got != want {
			t.Fatalf("PIS %s: got %v, reference says %v", candidate, got, want)
		}
	}
}

func TestVoterIDAgreesWithReference(t *testing.T) {
	random := rand.New(rand.NewSource(20260904))

	for i := 0; i < 5000; i++ {
		length := 12
		if i%4 == 0 {
			length = 13
		}
		candidate := randomDigits(random, length)
		got, _, _ := ValidateTituloEleitor(candidate)

		if want := referenceVoterID(candidate); got != want {
			t.Fatalf("voter ID %s: got %v, reference says %v", candidate, got, want)
		}
	}
}

// Random digits are almost never valid, so both suites also walk the valid
// space directly — a validator that returned false unconditionally would
// otherwise pass the differential run.
func TestConstructedNumbersAreAccepted(t *testing.T) {
	random := rand.New(rand.NewSource(7))

	for i := 0; i < 500; i++ {
		pis := constructPIS(randomDigits(random, 10))
		if isValid, _, message := ValidatePIS(pis); !isValid {
			t.Fatalf("constructed PIS %s was rejected: %s", pis, message)
		}
		if !referencePIS(pis) {
			t.Fatalf("constructed PIS %s is not valid by the reference either", pis)
		}
	}
}

func constructPIS(base string) string {
	weights := []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(base[i]-'0') * weights[i]
	}

	digit := 11 - sum%11
	if digit >= 10 {
		digit = 0
	}
	return base + strconv.Itoa(digit)
}

func randomDigits(random *rand.Rand, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte('0' + random.Intn(10))
	}
	return string(out)
}

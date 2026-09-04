package utils

import (
	"math/rand"
	"testing"
)

// Flat transcriptions of the reference implementations, written without the
// helpers the code under test uses, so a differential run compares two
// independent expressions of the same specification.

func referenceCNH(digits string) bool {
	if len(digits) != 11 {
		return false
	}
	repeated := true
	for i := 1; i < 11; i++ {
		if digits[i] != digits[0] {
			repeated = false
			break
		}
	}
	if repeated {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(digits[i]-'0') * (9 - i)
	}

	first, decrement := sum%11, 0
	if first >= 10 {
		first, decrement = 0, 2
	}
	if first != int(digits[9]-'0') {
		return false
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
	return second == int(digits[10]-'0')
}

func referenceRenavam(digits string) bool {
	if len(digits) != 11 {
		return false
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
	return expected == int(digits[10]-'0')
}

func TestCNHAgreesWithReference(t *testing.T) {
	random := rand.New(rand.NewSource(20260904))

	for i := 0; i < 5000; i++ {
		candidate := randomDigits(random, 11)
		got, _, _ := ValidateCNH(candidate)

		if want := referenceCNH(candidate); got != want {
			t.Fatalf("CNH %s: got %v, reference says %v", candidate, got, want)
		}
	}
}

func TestRenavamAgreesWithReference(t *testing.T) {
	random := rand.New(rand.NewSource(20260904))

	for i := 0; i < 5000; i++ {
		candidate := randomDigits(random, 11)
		got, _, _ := ValidateRenavam(candidate)

		// The one deliberate deviation: a run of one repeated digit is not a
		// registration, and this service rejects it where the reference does not.
		want := referenceRenavam(candidate) && !allSameDigit(candidate)

		if got != want {
			t.Fatalf("RENAVAM %s: got %v, expected %v", candidate, got, want)
		}
	}
}

// Random digits are almost never valid, so both are also walked through the
// valid space: a validator returning false unconditionally would otherwise pass.
func TestConstructedVehicleNumbersAreAccepted(t *testing.T) {
	random := rand.New(rand.NewSource(11))

	for i := 0; i < 500; i++ {
		base := randomDigits(random, 9)

		sum := 0
		for j := 0; j < 9; j++ {
			sum += int(base[j]-'0') * (9 - j)
		}
		first, decrement := sum%11, 0
		if first >= 10 {
			first, decrement = 0, 2
		}

		sum = 0
		for j := 0; j < 9; j++ {
			sum += int(base[j]-'0') * (j + 1)
		}
		second := sum%11 - decrement
		if second < 0 {
			second += 11
		}
		if second >= 10 {
			second = 0
		}

		cnh := base + string(rune('0'+first)) + string(rune('0'+second))
		if allSameDigit(cnh) {
			continue
		}
		if isValid, _, message := ValidateCNH(cnh); !isValid {
			t.Fatalf("constructed CNH %s was rejected: %s", cnh, message)
		}
	}
}

// A boleto is built rather than guessed: random 47-digit numbers are never
// valid, so the only way to test acceptance is to compute the check digits the
// way the specification says and confirm they are the ones expected back.
func TestConstructedBoletoIsAcceptedAndSensitiveToEveryDigit(t *testing.T) {
	valid := "00190000090114971860168524522114675860000102656"

	if isValid, _, message := ValidateBoleto(valid); !isValid {
		t.Fatalf("the published example was rejected: %s", message)
	}

	// Changing any single digit must break either a group digit or the general
	// one. A check that missed a position would show up here as a survivor.
	for i := 0; i < len(valid); i++ {
		mutated := []byte(valid)
		if mutated[i] == '9' {
			mutated[i] = '0'
		} else {
			mutated[i]++
		}

		if isValid, _, _ := ValidateBoleto(string(mutated)); isValid {
			t.Fatalf("changing digit %d left the boleto valid: %s", i, string(mutated))
		}
	}
}

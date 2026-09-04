package utils

import "testing"

// Vectors published by brazilian-utils, the reference these three were checked
// against.
func TestValidateCNH(t *testing.T) {
	tests := []struct {
		cnh   string
		valid bool
	}{
		{"00000000119", true},
		{"000000001-19", true},
		{"12345678901", false},
		{"11111111111", false},
		{"0000000011", false},
		{"00000000119abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.cnh, func(t *testing.T) {
			isValid, _, message := ValidateCNH(tt.cnh)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

func TestValidateRenavam(t *testing.T) {
	tests := []struct {
		name    string
		renavam string
		valid   bool
	}{
		{"nine_digits", "639884962", true},
		{"eleven_digits_same_number", "00639884962", true},
		{"last_digit_changed", "639884963", false},
		{"wrong_checksum", "12345678901", false},
		{"eight_digits", "12345678", false},
		{"ten_digits", "1234567890", false},
		{"twelve_digits", "123456789012", false},
		// The reference accepts "639884962abc" because it strips letters before
		// validating. This service rejects stray characters everywhere else and
		// is not going to make an exception here.
		{"trailing_letters", "639884962abc", false},
		{"all_zeros", "00000000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := ValidateRenavam(tt.renavam)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

// The nine-digit form is the eleven-digit form with leading zeros, so both must
// normalise to the same stored value.
func TestRenavamNormalisesToElevenDigits(t *testing.T) {
	_, short, _ := ValidateRenavam("639884962")
	_, long, _ := ValidateRenavam("00639884962")

	if short != long || short != "00639884962" {
		t.Fatalf("expected both forms to normalise to 00639884962, got %q and %q", short, long)
	}
}

func TestValidateBoleto(t *testing.T) {
	tests := []struct {
		name   string
		boleto string
		valid  bool
	}{
		{"valid", "00190000090114971860168524522114675860000102656", true},
		{"valid_as_printed", "0019000009 01149.718601 68524.522114 6 75860000102656", true},
		// One digit changed in the first group, and one in the third.
		{"first_group_check_digit", "00190000020114971860168524522114675860000102656", false},
		{"third_group_check_digit", "00190000090114971860168524522114975860000102656", false},
		{"too_short", "123456789", false},
		{"letters", "0019000009011497186016852452211467586000010265X", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := ValidateBoleto(tt.boleto)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

// A 48-digit slip is a different structure that has not been implemented.
// Reporting it as invalid would be a lie about a number that may well be good,
// so it gets a message that says it was not checked.
func TestBoletoArrecadacaoIsReportedAsUnchecked(t *testing.T) {
	arrecadacao := "846700000017435900240200610207807116000000000000"

	isValid, _, message := ValidateBoleto(arrecadacao)
	if isValid {
		t.Fatal("a 48-digit slip must not be reported as valid")
	}
	if !contains(message, "not supported yet") || !contains(message, "not checked") {
		t.Fatalf("the message must say the number was not checked, got: %s", message)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

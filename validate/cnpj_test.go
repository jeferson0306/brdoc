package validate

import "testing"

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name      string
		cnpj      string
		valid     bool
		sanitized string
	}{
		// Petrobras and the Banco do Brasil — real, published CNPJs.
		{"formatted", "33.000.167/0001-01", true, "33000167000101"},
		{"bare", "00000000000191", true, "00000000000191"},
		{"spaced", "33 000 167 0001 01", true, "33000167000101"},

		{"wrong_first_check_digit", "33.000.167/0001-11", false, "33000167000111"},
		{"wrong_second_check_digit", "33.000.167/0001-02", false, "33000167000102"},
		{"too_short", "33.000.167/0001", false, "330001670001"},
		{"too_long", "33.000.167/0001-010", false, "330001670001010"},

		// Passes the check-digit arithmetic, so it needs excluding explicitly.
		{"all_zeros", "00.000.000/0000-00", false, "00000000000000"},
		{"all_ones", "11.111.111/1111-11", false, "11111111111111"},

		// The guard from the sanitisation fix must cover CNPJ too.
		{"trailing_letters", "33.000.167/0001-01abc", false, "33.000.167/0001-01abc"},
		{"leading_letters", "abc33.000.167/0001-01", false, "abc33.000.167/0001-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, sanitized, message := checkCNPJ(tt.cnpj)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
			if sanitized != tt.sanitized {
				t.Fatalf("expected sanitized %q, got %q", tt.sanitized, sanitized)
			}
		})
	}
}

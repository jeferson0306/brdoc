package validate

import "testing"

func TestValidatePlate(t *testing.T) {
	tests := []struct {
		name  string
		plate string
		valid bool
	}{
		{"mercosul", "ABC1D23", true},
		{"mercosul_lowercase", "abc1d23", true},
		{"pre_mercosul", "ABC1234", true},
		{"pre_mercosul_hyphenated", "ABC-1234", true},
		{"pre_mercosul_spaced", "ABC 1234", true},

		{"letter_in_the_wrong_slot", "AB1CD23", false},
		{"all_digits", "1234567", false},
		{"all_letters", "ABCDEFG", false},
		{"too_short", "ABC123", false},
		{"too_long", "ABC12345", false},
		// A plate carries no check digit, so the format is the whole validation
		// — which makes stray punctuation the only thing left to reject.
		{"foreign_symbol", "ABC1D23!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := checkPlate(tt.plate)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

func TestValidateDocument(t *testing.T) {
	tests := []struct {
		name     string
		document string
		valid    bool
	}{
		{"cpf", "529.982.247-25", true},
		{"cnpj", "33.000.167/0001-01", true},
		{"cpf_bad_check_digit", "529.982.247-26", false},
		{"cnpj_bad_check_digit", "33.000.167/0001-11", false},
		// Eleven or fourteen digits, nothing in between: the length is what
		// decides which validator answers.
		{"twelve_digits", "123456789012", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := checkDocument(tt.document)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
		})
	}
}

func TestValidatePixKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
		form  string
	}{
		{"cpf", "529.982.247-25", true, "Valid PIX key (CPF)"},
		{"cnpj", "33.000.167/0001-01", true, "Valid PIX key (CNPJ)"},
		{"email", "jeferson@example.com", true, "Valid PIX key (email)"},
		{"phone_mobile", "+5561991946758", true, "Valid PIX key (phone)"},
		{"phone_landline", "+551133334444", true, "Valid PIX key (phone)"},
		{"random_uuid", "123e4567-e89b-12d3-a456-426614174000", true, "Valid PIX key (random)"},
		{"random_uuid_uppercase", "123E4567-E89B-12D3-A456-426614174000", true, "Valid PIX key (random)"},

		{"cpf_bad_check_digit", "529.982.247-26", false, ""},
		{"malformed_email", "jeferson@@example.com", false, ""},
		// A phone key must be in E.164 with the country code; a bare national
		// number is ambiguous and the arrangement does not accept it.
		{"phone_without_country_code", "61991946758", false, ""},
		{"uuid_missing_a_section", "123e4567-e89b-12d3-426614174000", false, ""},
		{"empty", "   ", false, ""},
		{"arbitrary_text", "minha-chave-pix", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := checkPixKey(tt.key)
			if isValid != tt.valid {
				t.Fatalf("expected valid=%v, got %v (%s)", tt.valid, isValid, message)
			}
			if tt.valid && message != tt.form {
				t.Fatalf("expected the key to be recognised as %q, got %q", tt.form, message)
			}
		})
	}
}

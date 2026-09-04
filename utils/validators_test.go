package utils

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		valid    bool
		sanitize string
	}{
		{"valid_email", "USER@Test.COM", true, "user@test.com"},
		{"valid_email_with_spaces", "  user@test.com  ", true, "user@test.com"},
		{"invalid_email", "invalid@", false, "invalid@"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, sanitized, _ := ValidateEmail(tt.email)
			if isValid != tt.valid {
				t.Fatalf("expected %v, got %v", tt.valid, isValid)
			}
			if sanitized != tt.sanitize {
				t.Fatalf("expected sanitized %q, got %q", tt.sanitize, sanitized)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		valid    bool
		sanitize string
	}{
		{"valid_mobile", "(11) 91234-5678", true, "11912345678"},
		{"valid_landline", "1132345678", true, "1132345678"},
		{"valid_with_country_code", "+55 (11) 91234-5678", true, "11912345678"},
		{"invalid_mobile_prefix", "11812345678", false, "11812345678"},
		{"invalid_ddd", "00123456789", false, "00123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, sanitized, _ := ValidatePhone(tt.phone)
			if isValid != tt.valid {
				t.Fatalf("expected %v, got %v", tt.valid, isValid)
			}
			if sanitized != tt.sanitize {
				t.Fatalf("expected sanitized %q, got %q", tt.sanitize, sanitized)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	isValid, sanitized, _ := ValidateName("  João   da   Silva ")
	if !isValid {
		t.Fatal("expected valid name")
	}
	if sanitized != "JOAO DA SILVA" {
		t.Fatalf("expected JOAO DA SILVA, got %q", sanitized)
	}
}

func TestValidateCPF(t *testing.T) {
	valid, _, _ := ValidateCPF("529.982.247-25")
	if !valid {
		t.Fatal("expected valid cpf")
	}

	invalid, _, _ := ValidateCPF("111.111.111-11")
	if invalid {
		t.Fatal("expected invalid cpf")
	}
}

func TestValidateCEP(t *testing.T) {
	valid, sanitized, _ := ValidateCEP("01310-100")
	if !valid {
		t.Fatal("expected valid cep")
	}
	if sanitized != "01310100" {
		t.Fatalf("expected sanitized 01310100, got %q", sanitized)
	}

	invalid, _, _ := ValidateCEP("00000-000")
	if invalid {
		t.Fatal("expected invalid cep for all zeroes")
	}
}

// Regression: every case below used to return valid. Non-digits were removed
// before the check digits were computed, so anything wrapped around a real
// value passed.
func TestStrayCharactersAreRejected(t *testing.T) {
	tests := []struct {
		name  string
		check func() (bool, string, string)
	}{
		{"cpf_trailing_letters", func() (bool, string, string) { return ValidateCPF("529.982.247-25jasasas") }},
		{"cpf_leading_letters", func() (bool, string, string) { return ValidateCPF("abc529.982.247-25") }},
		{"cpf_foreign_symbol", func() (bool, string, string) { return ValidateCPF("529.982.247-25/") }},
		{"rg_trailing_letters", func() (bool, string, string) { return ValidateRG("12.345.678abc") }},
		{"cep_trailing_letters", func() (bool, string, string) { return ValidateCEP("70040-010abc") }},
		{"phone_trailing_letters", func() (bool, string, string) { return ValidatePhone("(11) 98765-4321abc") }},
		{"card_trailing_letters", func() (bool, string, string) { return ValidatePlastic("4111111111111111zz") }},
		{"name_with_digits", func() (bool, string, string) { return ValidateName("Jeferson123 Siqueira") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := tt.check()
			if isValid {
				t.Fatalf("expected rejection, got valid (%s)", message)
			}
		})
	}
}

// The guard must not reject the formatting people actually type.
func TestLegitimateFormattingStillPasses(t *testing.T) {
	tests := []struct {
		name  string
		check func() (bool, string, string)
	}{
		{"cpf_dotted", func() (bool, string, string) { return ValidateCPF("529.982.247-25") }},
		{"cpf_spaced", func() (bool, string, string) { return ValidateCPF("529 982 247 25") }},
		{"cpf_bare", func() (bool, string, string) { return ValidateCPF("52998224725") }},
		{"rg_dotted", func() (bool, string, string) { return ValidateRG("12.345.678") }},
		{"cep_dashed", func() (bool, string, string) { return ValidateCEP("70040-010") }},
		{"phone_parenthesised", func() (bool, string, string) { return ValidatePhone("(11) 98765-4321") }},
		{"phone_country_code", func() (bool, string, string) { return ValidatePhone("+55 11 98765-4321") }},
		{"card_spaced", func() (bool, string, string) { return ValidatePlastic("4111 1111 1111 1111") }},
		{"name_accented", func() (bool, string, string) { return ValidateName("Jeférson D'Ávila Siqueira") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _, message := tt.check()
			if !isValid {
				t.Fatalf("expected valid, got rejected (%s)", message)
			}
		})
	}
}

// A rejected value is echoed back untouched: it was never accepted, so
// presenting a "sanitised" version of it would be misleading.
func TestRejectedValueIsEchoedRaw(t *testing.T) {
	_, value, _ := ValidateCPF("abc529.982.247-25")
	if value != "abc529.982.247-25" {
		t.Fatalf("expected the raw input back, got %q", value)
	}
}

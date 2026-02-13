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

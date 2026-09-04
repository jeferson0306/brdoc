package utils

import "testing"

// Every number here is taken from the "roteiro de crítica" the state itself
// publishes through SINTEGRA — the worked example in the state's own document,
// completed with the check digit that document calculates.
//
// They are the reason this file exists: an algorithm transcribed from a
// specification and never run against the specification's own example is a
// guess with good grammar.
func TestInscricaoEstadualOfficialExamples(t *testing.T) {
	official := map[string]string{
		"AC": "0100482300112",
		"AL": "240000048",
		"AP": "030123459",
		"BA": "12345663",
		"CE": "060000015",
		"GO": "109876547",
		"MA": "120000385",
		"MT": "00130000019",
		"PA": "159999995",
		"PB": "060000015",
		"PI": "012345679",
		"PR": "1234567850",
		"RO": "101625213",
		"RR": "240061536",
		"RS": "2243658792",
		"SC": "251040852",
		"SE": "271234563",
		"SP": "110042490114",
		"TO": "29010227836",
	}

	for uf, ie := range official {
		t.Run(uf, func(t *testing.T) {
			isValid, _, message := ValidateInscricaoEstadual(ie, uf)
			if !isValid {
				t.Fatalf("%s rejected its own published example %s: %s", uf, ie, message)
			}
		})
	}
}

// Bahia publishes examples for both of its lengths and both of its modules.
// The nine-digit module-10 case is the one the reference implementation this
// was checked against gets wrong, which is why it is called out separately.
func TestInscricaoEstadualBahiaBothLengths(t *testing.T) {
	for _, ie := range []string{"12345663", "100000306"} {
		if isValid, _, message := ValidateInscricaoEstadual(ie, "BA"); !isValid {
			t.Fatalf("BA rejected its published example %s: %s", ie, message)
		}
	}
}

func TestInscricaoEstadualRejections(t *testing.T) {
	tests := []struct {
		name string
		ie   string
		uf   string
	}{
		{"wrong_check_digit", "0100482300113", "AC"},
		{"wrong_length", "010048230011", "AC"},
		{"wrong_prefix", "0200482300112", "AC"},
		{"letters", "010048230011X", "AC"},
		{"valid_elsewhere", "0100482300112", "SP"},
		{"unknown_state", "0100482300112", "XX"},
		{"missing_state", "0100482300112", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isValid, _, message := ValidateInscricaoEstadual(tt.ie, tt.uf); isValid {
				t.Fatalf("expected rejection, got valid (%s)", message)
			}
		})
	}
}

func TestInscricaoEstadualExempt(t *testing.T) {
	// A business can be registered and exempt from holding a number; a
	// validator that rejected "ISENTO" would be rejecting a legal answer.
	for _, value := range []string{"ISENTO", "isento", "  Isento  "} {
		if isValid, sanitized, _ := ValidateInscricaoEstadual(value, "SP"); !isValid || sanitized != "ISENTO" {
			t.Fatalf("expected %q to be accepted as exempt, got valid=%v value=%q", value, isValid, sanitized)
		}
	}
}

func TestSupportedStatesCoverTheFederation(t *testing.T) {
	if got := len(SupportedIEStates()); got != 27 {
		t.Fatalf("expected all 27 states, got %d: %v", got, SupportedIEStates())
	}
}

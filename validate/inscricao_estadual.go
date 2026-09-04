package validate

import (
	"regexp"
	"sort"
	"strings"
)

// A state registration is only meaningful alongside the state that issued it:
// the same digits can be valid in one and nonsense in another. The API
// therefore requires both, and says so rather than guessing.

var ieChars = regexp.MustCompile(`^[\dPp.\-/\s]*$`)

// supportedStates lists the states this service can check, sorted, for the
// error message when an unknown one arrives.
func supportedStates() []string {
	states := make([]string, 0, len(ieValidators))
	for uf := range ieValidators {
		states = append(states, uf)
	}
	sort.Strings(states)
	return states
}

// checkInscricaoEstadual checks a state registration against the roteiro de
// crítica published by the issuing state.
//
// The state is a required argument, not an inference. A registration carries no
// marker of where it came from, so a service that guessed would be answering a
// question it was not asked.
func checkInscricaoEstadual(ie, uf string) (bool, string, string) {
	// "Isento" — exempt — is a legitimate value in every state for a business
	// that is registered but has no number assigned. It has to be recognised
	// before the character guard, which allows only digits and formatting.
	if strings.EqualFold(strings.TrimSpace(ie), "ISENTO") {
		return true, "ISENTO", "Valid state registration (exempt)"
	}

	if !ieChars.MatchString(ie) {
		return false, ie, "Invalid state registration (unexpected characters)"
	}

	state := strings.ToUpper(strings.TrimSpace(uf))
	if state == "" {
		return false, ie, "Missing state: a registration cannot be checked without the state that issued it"
	}

	validate, supported := ieValidators[state]
	if !supported {
		return false, ie, "Unknown state " + state + " (expected one of " + strings.Join(supportedStates(), ", ") + ")"
	}

	// São Paulo's rural registrations open with a P; every other character that
	// is not a digit is formatting and comes out.
	sanitized := strings.ToUpper(strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == 'P' || r == 'p' {
			return r
		}
		return -1
	}, ie))

	if !validate(sanitized) {
		return false, sanitized, "Invalid state registration for " + state
	}

	return true, sanitized, "Valid state registration (" + state + ")"
}

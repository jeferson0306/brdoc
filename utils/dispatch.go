package utils

// One table maps a parameter name to the validator behind it. Both the
// single-value endpoint and the batch endpoint read from it, so a new document
// becomes available on both the moment it is added here — and neither can
// silently fall behind the other.

// Outcome is what a validator answers, independent of how the request arrived.
type Outcome struct {
	// Sanitized is the value normalised for storage, or the raw input when it
	// was rejected outright.
	Sanitized string
	Valid     bool
	Message   string
	// FromCache is only ever true for CPF, the one check expensive enough to
	// be worth remembering.
	FromCache bool
}

// Qualifier is the second argument some validations need. Only inscrição
// estadual uses one today — the issuing state — but the signature exists so a
// document that needs context does not have to break the table to get it.
type validatorFunc func(value, qualifier string) Outcome

func plain(f func(string) (bool, string, string)) validatorFunc {
	return func(value, _ string) Outcome {
		valid, sanitized, message := f(value)
		return Outcome{Sanitized: sanitized, Valid: valid, Message: message}
	}
}

var validators = map[string]validatorFunc{
	"email": plain(ValidateEmail),
	"cpf": func(value, _ string) Outcome {
		valid, sanitized, message, fromCache := ValidateCPFWithCache(value)
		return Outcome{Sanitized: sanitized, Valid: valid, Message: message, FromCache: fromCache}
	},
	"cnpj":      plain(ValidateCNPJ),
	"documento": plain(ValidateDocument),
	"ie": func(value, qualifier string) Outcome {
		valid, sanitized, message := ValidateInscricaoEstadual(value, qualifier)
		return Outcome{Sanitized: sanitized, Valid: valid, Message: message}
	},
	"pis":       plain(ValidatePIS),
	"cnh":       plain(ValidateCNH),
	"renavam":   plain(ValidateRenavam),
	"boleto":    plain(ValidateBoleto),
	"titulo":    plain(ValidateTituloEleitor),
	"placa":     plain(ValidatePlate),
	"pix":       plain(ValidatePixKey),
	"name":      plain(ValidateName),
	"telephone": plain(ValidatePhone),
	"plastic":   plain(ValidatePlastic),
	"rg":        plain(ValidateRG),
	"cep":       plain(ValidateCEP),
}

// ValidationKeys is the order the single-value endpoint checks parameters in.
// It is explicit rather than derived from the map, because map iteration order
// is random in Go and the endpoint's behaviour must not be.
var ValidationKeys = []string{
	"email", "cpf", "cnpj", "documento", "ie", "pis", "titulo",
	"cnh", "renavam", "placa", "boleto", "pix",
	"name", "telephone", "plastic", "rg", "cep",
}

// SupportedKeys reports what can be validated, for error messages.
func SupportedKeys() []string {
	keys := make([]string, len(ValidationKeys))
	copy(keys, ValidationKeys)
	return keys
}

// Validate runs the validator registered for key. The second return reports
// whether the key is one this service knows at all, which is a different
// failure from a value that did not pass.
func Validate(key, value, qualifier string) (Outcome, bool) {
	validate, known := validators[key]
	if !known {
		return Outcome{}, false
	}
	return validate(value, qualifier), true
}

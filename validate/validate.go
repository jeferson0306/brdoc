// Package validate checks Brazilian documents.
//
// Everything here is pure: a function takes a string, returns an answer, and
// touches nothing else. There is no network, no cache, no configuration and no
// state — importing this package opens no connections and reads no environment.
// That is deliberate. Validation is arithmetic, and arithmetic that needs a
// server is arithmetic you cannot trust when the server is down.
//
// Values are checked, not laundered. Formatting a person legitimately types —
// dots, dashes, spaces, the slash in a CNPJ — is accepted and removed; anything
// else is a rejection. Stripping unexpected characters and validating what
// remains would accept "abc529.982.247-25" as a CPF, which is how a validator
// ends up letting junk into a database.
//
//	result := validate.CPF("529.982.247-25")
//	result.Valid      // true
//	result.Normalized // "52998224725"
//
// For callers that do not know the document type at compile time — a form
// driven by configuration, say — [ByKey] dispatches on a string, and [Keys]
// lists what it accepts.
package validate

// Result is the answer to one check.
type Result struct {
	// Valid reports whether the value passed.
	Valid bool

	// Normalized is the value reduced to what a database should store:
	// digits only for most documents, lower-cased for an email, upper-cased and
	// unpunctuated for a plate. When a value is rejected outright it is
	// returned untouched, because presenting a cleaned-up version of something
	// that was never accepted is misleading.
	Normalized string

	// Reason explains the outcome in English, whether it passed or failed. It
	// is meant for a log or a developer, not for an end user — a message shown
	// to a person should be written by the application, in its own voice and
	// language.
	Reason string
}

func result(valid bool, normalized, reason string) Result {
	return Result{Valid: valid, Normalized: normalized, Reason: reason}
}

// CPF checks an individual taxpayer number, including both check digits.
func CPF(value string) Result { return result(checkCPF(value)) }

// CNPJ checks a company taxpayer number, including both check digits.
func CNPJ(value string) Result { return result(checkCNPJ(value)) }

// Document accepts whichever of CPF or CNPJ the digit count describes, for the
// single "documento" field forms usually offer.
func Document(value string) Result { return result(checkDocument(value)) }

// PIS checks a PIS, PASEP, NIT or NIS number — four names for one number.
func PIS(value string) Result { return result(checkPIS(value)) }

// VoterID checks a título de eleitor, including the issuing state encoded in it.
func VoterID(value string) Result { return result(checkTituloEleitor(value)) }

// CNH checks a driving licence number. Its two check digits are coupled: when
// the first overflows it is written as zero and the second compensates.
func CNH(value string) Result { return result(checkCNH(value)) }

// Renavam checks a vehicle registration number in either the nine-digit or the
// eleven-digit form, normalising both to eleven.
func Renavam(value string) Result { return result(checkRenavam(value)) }

// Plate checks a vehicle plate in the pre-Mercosul or Mercosul pattern.
func Plate(value string) Result { return result(checkPlate(value)) }

// Boleto checks the 47-digit linha digitável of a bank slip: the three typed
// groups and then the barcode's own check digit.
func Boleto(value string) Result { return result(checkBoleto(value)) }

// PixKey checks a PIX key in any of its five forms — CPF, CNPJ, email, phone in
// E.164, or a random UUID — inferring which it is.
func PixKey(value string) Result { return result(checkPixKey(value)) }

// StateRegistration checks an inscrição estadual against the rules of the state
// that issued it. The state is required: the same digits can be valid in one
// state and nonsense in another.
func StateRegistration(value, uf string) Result {
	return result(checkInscricaoEstadual(value, uf))
}

// Email checks the address format and normalises its case.
func Email(value string) Result { return result(checkEmail(value)) }

// FullName checks that a name is a plausible full name and normalises it.
func FullName(value string) Result { return result(checkName(value)) }

// Phone checks a Brazilian landline or mobile number, area code included.
func Phone(value string) Result { return result(checkPhone(value)) }

// Card checks a card number with the Luhn checksum and identifies the brand.
func Card(value string) Result { return result(checkPlastic(value)) }

// CardBrand names the brand of a card number, or an empty string when no
// pattern matches. It does not validate — use [Card] for that.
func CardBrand(value string) string {
	if brand := cardBrandOf(value); brand != "Unknown" {
		return brand
	}
	return ""
}

// RG checks the format of a state identity number. There is no national
// checksum, so this is a shape check and nothing more.
func RG(value string) Result { return result(checkRG(value)) }

// CEP checks a postal code.
func CEP(value string) Result { return result(checkCEP(value)) }

// States lists the state codes [StateRegistration] can check, sorted.
func States() []string { return supportedStates() }

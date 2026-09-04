package validate

// Dispatch by string, for callers that do not know the document type at compile
// time — a form driven by configuration, or an HTTP endpoint taking the type
// from a query parameter.

type checker func(value, qualifier string) Result

func ignoringQualifier(f func(string) Result) checker {
	return func(value, _ string) Result { return f(value) }
}

var checkers = map[string]checker{
	"cpf":       ignoringQualifier(CPF),
	"cnpj":      ignoringQualifier(CNPJ),
	"documento": ignoringQualifier(Document),
	"pis":       ignoringQualifier(PIS),
	"titulo":    ignoringQualifier(VoterID),
	"cnh":       ignoringQualifier(CNH),
	"renavam":   ignoringQualifier(Renavam),
	"placa":     ignoringQualifier(Plate),
	"boleto":    ignoringQualifier(Boleto),
	"pix":       ignoringQualifier(PixKey),
	"email":     ignoringQualifier(Email),
	"name":      ignoringQualifier(FullName),
	"telephone": ignoringQualifier(Phone),
	"plastic":   ignoringQualifier(Card),
	"rg":        ignoringQualifier(RG),
	"cep":       ignoringQualifier(CEP),
	"ie":        StateRegistration,
}

// keyOrder is explicit rather than derived from the map. Go randomises map
// iteration, and a caller walking the supported types deserves a stable answer.
var keyOrder = []string{
	"email", "cpf", "cnpj", "documento", "ie", "pis", "titulo",
	"cnh", "renavam", "placa", "boleto", "pix",
	"name", "telephone", "plastic", "rg", "cep",
}

// Keys lists the document types [ByKey] accepts, in a stable order.
func Keys() []string {
	keys := make([]string, len(keyOrder))
	copy(keys, keyOrder)
	return keys
}

// ByKey runs the check registered for key. The qualifier carries the extra
// context a type may need — today only "ie", which takes the issuing state.
//
// The second return distinguishes an unknown key from a value that failed:
// they are different mistakes and deserve different handling.
func ByKey(key, value, qualifier string) (Result, bool) {
	check, known := checkers[key]
	if !known {
		return Result{}, false
	}
	return check(value, qualifier), true
}

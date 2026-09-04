package validate_test

import (
	"testing"

	"github.com/jeferson0306/api-data-validator/validate"
)

// These exist to answer one question the service's design depends on: is any of
// this expensive enough to be worth remembering between requests?
func BenchmarkCPF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validate.CPF("529.982.247-25")
	}
}

func BenchmarkCNPJ(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validate.CNPJ("33.000.167/0001-01")
	}
}

func BenchmarkBoleto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validate.Boleto("00190000090114971860168524522114675860000102656")
	}
}

func BenchmarkStateRegistration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validate.StateRegistration("0100482300112", "AC")
	}
}

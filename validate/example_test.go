package validate_test

import (
	"fmt"

	"github.com/jeferson0306/brdoc/validate"
)

func ExampleCPF() {
	result := validate.CPF("529.982.247-25")

	fmt.Println(result.Valid)
	fmt.Println(result.Normalized)
	// Output:
	// true
	// 52998224725
}

// Formatting a person types is accepted and removed. Anything else is a
// rejection — the letters here are not stripped and ignored.
func ExampleCPF_strayCharacters() {
	fmt.Println(validate.CPF("529.982.247-25").Valid)
	fmt.Println(validate.CPF("abc529.982.247-25").Valid)
	// Output:
	// true
	// false
}

// A rejected value comes back untouched: it was never accepted, so presenting a
// cleaned-up version of it would be misleading.
func ExampleResult_normalized() {
	fmt.Printf("%q\n", validate.CEP("70040-010").Normalized)
	fmt.Printf("%q\n", validate.CEP("70040-010abc").Normalized)
	// Output:
	// "70040010"
	// "70040-010abc"
}

// The single "document" field most forms offer, taking either a CPF or a CNPJ.
func ExampleDocument() {
	fmt.Println(validate.Document("529.982.247-25").Reason)
	fmt.Println(validate.Document("33.000.167/0001-01").Reason)
	// Output:
	// Valid document (CPF)
	// Valid document (CNPJ)
}

// A PIX key arrives pasted into one box with nothing saying which of the five
// forms it is, so the form is inferred.
func ExamplePixKey() {
	for _, key := range []string{
		"529.982.247-25",
		"jeferson@example.com",
		"+5511987654321",
		"123e4567-e89b-12d3-a456-426614174000",
	} {
		fmt.Println(validate.PixKey(key).Reason)
	}
	// Output:
	// Valid PIX key (CPF)
	// Valid PIX key (email)
	// Valid PIX key (phone)
	// Valid PIX key (random)
}

// The state is required: the same digits can be valid in one state and nonsense
// in another.
func ExampleStateRegistration() {
	fmt.Println(validate.StateRegistration("0100482300112", "AC").Valid)
	fmt.Println(validate.StateRegistration("0100482300112", "SP").Valid)
	// Output:
	// true
	// false
}

// For callers that learn the document type at runtime.
func ExampleByKey() {
	result, known := validate.ByKey("cnpj", "33.000.167/0001-01", "")

	fmt.Println(known, result.Valid)
	// Output: true true
}

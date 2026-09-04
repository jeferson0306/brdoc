// Command brdoc checks Brazilian documents from the shell.
//
// It exits 0 when the value is valid and 1 when it is not, so it composes with
// everything else a shell can do:
//
//	brdoc cpf 529.982.247-25 && echo accepted
//	brdoc ie 0100482300112 AC
//	brdoc --json cnpj 33.000.167/0001-01 | jq .normalized
//
// Nothing leaves the machine. The library it wraps performs arithmetic and
// makes no network calls, which is the point of having a CLI at all.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jeferson0306/api-data-validator/validate"
)

const (
	exitValid   = 0
	exitInvalid = 1
	exitUsage   = 2
)

func main() {
	asJSON := flag.Bool("json", false, "print the result as JSON")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 1 && args[0] == "list" {
		fmt.Println(strings.Join(validate.Keys(), "\n"))
		os.Exit(exitValid)
	}

	if len(args) < 2 {
		usage()
		os.Exit(exitUsage)
	}

	key, value := args[0], args[1]
	qualifier := ""
	if len(args) > 2 {
		qualifier = args[2]
	}

	result, known := validate.ByKey(key, value, qualifier)
	if !known {
		fmt.Fprintf(os.Stderr, "unknown document type %q\nknown types: %s\n",
			key, strings.Join(validate.Keys(), ", "))
		os.Exit(exitUsage)
	}

	if *asJSON {
		// Field names match the library's, so a script reading this and a
		// program importing the package see the same shape.
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"valid":      result.Valid,
			"normalized": result.Normalized,
			"reason":     result.Reason,
		})
	} else {
		mark := "invalid"
		if result.Valid {
			mark = "valid"
		}
		fmt.Printf("%-8s %s  %s\n", mark, result.Normalized, result.Reason)
	}

	if !result.Valid {
		os.Exit(exitInvalid)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `brdoc — check Brazilian documents

usage:
  brdoc [--json] <type> <value> [qualifier]
  brdoc list

  The qualifier is only used by "ie", which takes the issuing state:
    brdoc ie 0100482300112 AC

exit status:
  0  valid
  1  invalid
  2  usage error

`)
	flag.PrintDefaults()
}

Data Validator API

![Data Validator API - Thumbnail](assets/api-validation-thumbnail.png)

> Upwork portfolio thumbnail — Data validation API (email, CPF, phone, etc.)


A Go library for validating Brazilian documents, and an HTTP service built on it.

```bash
go get github.com/jeferson0306/api-data-validator/validate
```

```go
result := validate.CPF("529.982.247-25")
result.Valid      // true
result.Normalized // "52998224725"
```

The library is pure: no network, no cache, no configuration, no state. Importing
it opens no connections and reads no environment. Validation is arithmetic, and
arithmetic that needs a server is arithmetic you cannot trust when the server is
down.

Values are **checked, not laundered**. Formatting a person legitimately types —
dots, dashes, spaces, the slash in a CNPJ — is accepted and removed. Anything
else is a rejection. Stripping unexpected characters and validating what remains
accepts `abc529.982.247-25` as a CPF, which is how a validator ends up letting
junk into a database.

## What it checks

| | |
|---|---|
| People | CPF, PIS/PASEP/NIT/NIS, RG, título de eleitor, CNH, full name |
| Companies | CNPJ, inscrição estadual (**all 27 states**) |
| Either | `documento` — whichever of CPF or CNPJ the digits describe |
| Vehicles | RENAVAM, plate (pre-Mercosul and Mercosul) |
| Money | PIX key (all five forms), card number with brand, boleto linha digitável |
| Contact | email, phone, CEP |

Every algorithm was checked against a published reference or the issuing state's
own worked example before it shipped. Two of them disagree with the reference
implementation they were compared against, in both cases because the reference
is wrong — see `validate/inscricao_estadual_test.go`.

## Command line

```bash
go install github.com/jeferson0306/api-data-validator/cmd/brdoc@latest

brdoc cpf 529.982.247-25 && echo accepted
brdoc ie 0100482300112 AC
brdoc --json cnpj 33.000.167/0001-01 | jq .normalized
brdoc list
```

Exit status is 0 for valid, 1 for invalid and 2 for a usage error, so it
composes with everything else a shell can do.

## HTTP service

```
GET  /validate?cpf=529.982.247-25
GET  /validate?ie=0100482300112&uf=AC
POST /validate/batch     one form, one request
GET  /health
GET  /swagger/index.html
```

`POST /validate/batch` takes a list of items, each with its own type, and
answers one result per item plus a summary. A form with eight fields is one
request rather than eight.

## Performance

```
BenchmarkCPF-12                  522 ns/op
BenchmarkCNPJ-12                 499 ns/op
BenchmarkBoleto-12               864 ns/op
BenchmarkStateRegistration-12    248 ns/op
```

Worth knowing before reaching for a cache: a CPF check costs about half a
microsecond, against a Redis round trip measured at 150–220 ms in production.
The cache is opt-in through `REDIS_ADDR` and off by default.

## Configuration

Every value is optional and the service runs with none of them set.

| Variable | Effect |
|---|---|
| `PORT` | Listen address. Defaults to 8080. |
| `TRUSTED_PLATFORM` | `cloudflare` reads the caller's address from `CF-Connecting-IP`. **Set this on Render**, which fronts services with Cloudflare — without it every caller shares one rate-limit bucket. |
| `TRUSTED_PROXIES` | CIDRs allowed to set `X-Forwarded-For`, for other hosts. |
| `RATE_LIMIT_RPS` | Requests per second per caller. Defaults to 20. |
| `RATE_LIMIT_BURST` | Burst allowance. Defaults to 60. |
| `REDIS_ADDR` | Enables the optional CPF cache. Off by default, and the benchmarks above explain why. |
| `LOG_LEVEL` | `debug` for cache diagnostics. |

## Layout

```
validate/        the library — pure, importable, no dependencies beyond x/text
internal/cache/  the service's optional CPF cache, kept out of the library
handlers/        HTTP handlers
middleware/      CORS
observability/   request logging that never records the value being validated
cmd/brdoc/       the command line tool
main.go          the service
```

## Development

```bash
go test ./... -race
go test ./validate -bench .
go run ./cmd/brdoc cpf 529.982.247-25
```

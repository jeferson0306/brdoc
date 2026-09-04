// Package cache remembers CPF results between requests.
//
// It lives here rather than in the validate package on purpose: validation is
// arithmetic and must work with nothing else running, while a cache is
// infrastructure with a connection, a timeout and a failure mode. Keeping them
// apart is what lets the library be imported without opening a socket.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/jeferson0306/api-data-validator/validate"
)

const (
	// timeout bounds every call. Validation itself is measured in microseconds,
	// so an unreachable cache must never hold a request open — the answer is
	// recomputed instead, which costs less than waiting.
	timeout = 150 * time.Millisecond
	// retention is generous because the answer cannot change: a CPF that is
	// valid today was valid yesterday.
	retention = 24 * time.Hour
)

var client = redis.NewClient(&redis.Options{Addr: address()})

func address() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

// keyFor hashes the value before it becomes a key.
//
// The key used to be "cpf:" followed by the CPF itself, which put every CPF
// ever checked into the cache in the clear. A digest keys just as well — equal
// inputs still collide onto one entry — while storing nothing that identifies
// a person.
func keyFor(value string) string {
	digest := sha256.Sum256([]byte(value))
	return "cpf:" + hex.EncodeToString(digest[:])
}

// CPF validates through the cache, reporting whether the answer was remembered.
//
// Note what this costs. Building the key needs the normalised value, and
// normalising is most of the validation, so the answer is already computed
// before the cache is consulted — see the benchmarks in the validate package,
// where a CPF check runs in about 500 nanoseconds against a Redis round trip
// measured at 150 to 220 milliseconds in production. The cache is retained
// because from_cache is part of the published response; the numbers argue for
// removing both.
func CPF(value string) (validate.Result, bool) {
	result := validate.CPF(value)

	if !Enabled() {
		return result, false
	}

	key := keyFor(result.Normalized)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if cached, err := client.Get(ctx, key).Result(); err == nil {
		result.Valid = cached == "true"
		return result, true
	}

	stored := "false"
	if result.Valid {
		stored = "true"
	}

	if err := client.Set(ctx, key, stored, retention).Err(); err != nil {
		// Debug, not error: a missing cache is a degraded mode, not a failure,
		// and at request rate an unreachable Redis would drown the log.
		// Healthy is what surfaces the condition.
		slog.Debug("cache write failed", slog.String("error", err.Error()))
	}

	return result, false
}

// Enabled reports whether a cache was configured at all. Unset, the service
// skips the round trip rather than attempting one and timing out per request.
func Enabled() bool {
	return os.Getenv("REDIS_ADDR") != ""
}

// Healthy reports whether the cache is reachable, so a health endpoint can say
// so rather than the service degrading in silence.
func Healthy() bool {
	if !Enabled() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return client.Ping(ctx).Err() == nil
}

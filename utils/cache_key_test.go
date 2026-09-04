package utils

import (
	"strings"
	"testing"
)

// The key used to be "cpf:<the actual CPF>", so every CPF ever validated sat in
// Redis in the clear.
func TestCacheKeyDoesNotContainTheValue(t *testing.T) {
	key := cacheKeyFor("cpf", "52998224725")

	if strings.Contains(key, "52998224725") {
		t.Fatalf("the cache key still carries the CPF: %s", key)
	}
	if !strings.HasPrefix(key, "cpf:") {
		t.Fatalf("expected the namespace prefix to survive, got %s", key)
	}
}

func TestCacheKeyIsStableAndDistinct(t *testing.T) {
	// Equal inputs must still collide onto one entry, or the cache stops working.
	if cacheKeyFor("cpf", "52998224725") != cacheKeyFor("cpf", "52998224725") {
		t.Fatal("the same value produced two different keys")
	}
	if cacheKeyFor("cpf", "52998224725") == cacheKeyFor("cpf", "16899535009") {
		t.Fatal("two different values collided")
	}
}

package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Rate limiting here is per client address, which carries a known and
// deliberate cost: everyone behind one NAT — an office, a mobile carrier, a
// university — shares a bucket and is throttled as though they were a single
// abusive caller.
//
// That is not a theoretical objection. It is the finding a load test against a
// gateway of mine produced: unit and integration tests never generate
// concurrent traffic from distinct clients, so the behaviour is invisible until
// real load arrives. Keying on an authenticated identity is the fix, and this
// service has no authentication, so the limit is set generously instead — high
// enough that a shared address is not punished, low enough that a script is.
//
// The buckets live in memory rather than in Redis. Validation costs about half
// a microsecond; adding a network round trip to protect it would cost five
// orders of magnitude more than the work being protected.

const (
	defaultRequestsPerSecond = 20
	defaultBurst             = 60
	// idleBefore is how long a client is remembered after its last request.
	// Without eviction the map grows for the lifetime of the process.
	idleBefore = 10 * time.Minute
	sweepEvery = time.Minute
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type limiterSet struct {
	mu      sync.Mutex
	clients map[string]*client
	rps     rate.Limit
	burst   int
}

func newLimiterSet() *limiterSet {
	set := &limiterSet{
		clients: make(map[string]*client),
		rps:     rate.Limit(envInt("RATE_LIMIT_RPS", defaultRequestsPerSecond)),
		burst:   envInt("RATE_LIMIT_BURST", defaultBurst),
	}

	go set.sweep()
	return set
}

func (s *limiterSet) allow(address string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, seen := s.clients[address]
	if !seen {
		existing = &client{limiter: rate.NewLimiter(s.rps, s.burst)}
		s.clients[address] = existing
	}
	existing.lastSeen = time.Now()

	return existing.limiter.Allow()
}

func (s *limiterSet) sweep() {
	for range time.Tick(sweepEvery) {
		s.mu.Lock()
		for address, c := range s.clients {
			if time.Since(c.lastSeen) > idleBefore {
				delete(s.clients, address)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit throttles by client address.
//
// /health is exempt: a monitor that gets throttled reports an outage that is
// not happening, and health checks are the one caller whose rate is known and
// harmless.
func RateLimit() gin.HandlerFunc {
	set := newLimiterSet()

	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// ClientIP honours the trusted-proxy configuration set in main; behind
		// a platform proxy the socket address is the proxy's, not the caller's.
		if !set.allow(c.ClientIP()) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status_code": http.StatusTooManyRequests,
				"error_code":  "RATE_LIMITED",
				"message":     "Too many requests from this address",
				"is_valid":    false,
			})
			return
		}

		c.Next()
	}
}

func envInt(name string, fallback int) int {
	if raw := os.Getenv(name); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			return value
		}
	}
	return fallback
}

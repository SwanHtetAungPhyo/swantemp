package middleware

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SwanHtetAungPhyo/swantemp/closure"
)

type rateCounter struct {
	count     int32
	lastReset int64
}

var rateMap sync.Map

func RateLimitMiddleware(limit int, window time.Duration) *closure.Middleware {
	return &closure.Middleware{
		Name: "RateLimit",
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				ip := ctx.RemoteIP().String()

				val, _ := rateMap.LoadOrStore(ip, &rateCounter{})
				counter := val.(*rateCounter)

				now := time.Now().UnixNano()
				resetTime := atomic.LoadInt64(&counter.lastReset)
				if now-resetTime > window.Nanoseconds() {
					atomic.StoreInt32(&counter.count, 0)
					atomic.StoreInt64(&counter.lastReset, now)
				}
				if atomic.AddInt32(&counter.count, 1) > int32(limit) {
					return errors.New("too many requests")
				}

				return next(ctx)
			}
		},
	}
}

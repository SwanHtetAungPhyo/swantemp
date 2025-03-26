package middleware

import (
	"fmt"
	"github.com/SwanHtetAungPhyo/swantemp/closure"
	logging "github.com/SwanHtetAungPhyo/swantemp/log"
	"sync"
	"time"
)

type rateMapper struct {
	Url      string
	RemoteIp string
	Count    int
}

var statement = fmt.Sprintf("%s[CLOCACHE] %s", cyan, reset)

type cacheEntry struct {
	data    []byte
	expires int64
}

func CacheMiddleware(ttl time.Duration) *closure.Middleware {
	var (
		cache sync.Map
		_     sync.RWMutex
	)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now().Unix()
			cache.Range(func(key, value interface{}) bool {
				if entry := value.(*cacheEntry); entry.expires < now {
					cache.Delete(key)
				}
				return true
			})
		}
	}()

	return &closure.Middleware{
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				cacheKey := ctx.URI().String()

				if val, ok := cache.Load(cacheKey); ok {
					entry := val.(*cacheEntry)
					if time.Now().Unix() < entry.expires {
						ctx.SetBody(entry.data)
						return nil
					}
				}

				err := next(ctx)
				if err != nil {
					return err
				}

				cache.Store(cacheKey, &cacheEntry{
					data:    ctx.Response.Body(),
					expires: time.Now().Add(ttl).Unix(),
				})

				return nil
			}
		},
	}
}
func LoggerMiddleware() *closure.Middleware {
	return &closure.Middleware{
		Name: "Logger",
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				start := time.Now()
				req := &ctx.Request
				response := next(ctx)

				duration := time.Since(start)
				statusCode := ctx.Response.StatusCode()
				method := string(ctx.Method())
				clientIP := ctx.RemoteIP()

				logging.Info(
					"Method:", method,
					"Routes:", req.URI().String(),
					"Status:", statusCode,
					"Duration:", duration,
					"ClientIP:", clientIP,
				)

				return response
			}
		},
	}
}

package middleware

import (
	"github.com/SwanHtetAungPhyo/swantemp/closure"
	"github.com/valyala/fasthttp"
)

func RecoveryMiddleware() *closure.Middleware {
	return &closure.Middleware{
		Name: "Recovery",
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				defer func() {
					if err := recover(); err != nil {
						err = closure.JSONMe(ctx, fasthttp.StatusInternalServerError, "internal server error", nil)
					}
				}()
				return next(ctx)
			}
		},
	}
}

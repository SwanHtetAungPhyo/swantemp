package middleware

import (
	"github.com/SwanHtetAungPhyo/swantemp/closure"
)

func JWTMiddleware(secret string) *closure.Middleware {
	//manager := swan_lib.NewJWTMiddleware(secret)
	return &closure.Middleware{
		Name: "JWT",
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				//manager.FastAuthorize(func(ctx *closure.Context) {
				//	err := next(ctx)
				//	if err != nil {
				//		return
				//	}
				//})(ctx)		//manager.FastAuthorize(func(ctx *closure.Context) {
				//	err := next(ctx)
				//	if err != nil {
				//		return
				//	}
				//})(ctx)
				return nil
			}
		},
	}
}

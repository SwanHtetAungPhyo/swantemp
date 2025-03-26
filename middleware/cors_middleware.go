package middleware

import (
	"github.com/SwanHtetAungPhyo/swantemp/closure"
	"github.com/valyala/fasthttp"
	"strings"
)

const (
	ORIGIN_CONTROL      = "Access-Control-Allow-Origin"
	METHOD_CONTROL      = "Access-Control-Allow-Methods"
	HEADER_CONTROL      = "Access-Control-Allow-Headers"
	CREDENTIALS_CONTROL = "Access-Control-Allow-Credentials"
	GET                 = "GET"
	HEAD                = "HEAD"
	POST                = "POST"
	OPTIONS             = "OPTIONS"
	DELETE              = "DELETE"
	PUT                 = "PUT"
)

type CORSMiddleware struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{GET, POST, PUT, DELETE, OPTIONS},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	}
}

func (c *CORSMiddleware) AllowOrigins(origins ...string) *CORSMiddleware {
	c.AllowedOrigins = origins
	return c
}

func (c *CORSMiddleware) AllowMethods(methods ...string) *CORSMiddleware {
	c.AllowedMethods = methods
	return c
}

func (c *CORSMiddleware) AllowHeaders(headers ...string) *CORSMiddleware {
	c.AllowedHeaders = headers
	return c
}

func (c *CORSMiddleware) ToMiddleware() *closure.Middleware {
	return &closure.Middleware{
		Handler: func(next closure.Handler) closure.Handler {
			return func(ctx *closure.Context) error {
				origin := string(ctx.Request.Header.Peek("Origin"))

				if !c.isOriginAllowed(origin) {
					ctx.Error("Forbidden", fasthttp.StatusForbidden)
					return nil
				}

				ctx.Response.Header.Set(ORIGIN_CONTROL, origin)
				ctx.Response.Header.Set(METHOD_CONTROL, strings.Join(c.AllowedMethods, ","))
				ctx.Response.Header.Set(HEADER_CONTROL, strings.Join(c.AllowedHeaders, ","))

				if c.AllowCredentials {
					ctx.Response.Header.Set(CREDENTIALS_CONTROL, "true")
					if origin == "*" {
						origin = ctx.RemoteIP().String()
					}
				}

				if string(ctx.Method()) == "OPTIONS" {
					ctx.Response.SetStatusCode(fasthttp.StatusNoContent)
					return nil
				}

				return next(ctx)
			}
		},
	}
}
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range c.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}

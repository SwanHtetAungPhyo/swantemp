package main

import (
	"github.com/SwanHtetAungPhyo/swantemp/closure"
	"github.com/SwanHtetAungPhyo/swantemp/middleware"
	"github.com/valyala/fasthttp"
	_ "net/http/pprof"
)

type User struct {
	Name string `json:"name"`
}

func main() {
	app := closure.New(closure.WithMaxConnsPerIP(100))

	router := closure.NewRouter()

	app.ApplyMiddleware(
		*middleware.LoggerMiddleware(),
		*middleware.RecoveryMiddleware(),
	)

	mainCluster := closure.NewCluster("/", router)
	mainCluster.Group("/public", func(g *closure.Cluster) {
		g.Get("/info", GetHandler)
	})

	authCluster := closure.NewCluster("/auth", router, *middleware.LoggerMiddleware())
	authCluster.Group("/admin", func(g *closure.Cluster) {
		g.Get("/info", func(ctx *closure.Context) error {
			return closure.JSONMe(ctx, fasthttp.StatusAccepted, "Hello Admin", nil)
		})
	})

	app.Mount(mainCluster)
	app.Mount(authCluster)
	app.Start(":8080")

}

func GetHandler(ctx *closure.Context) error {
	user := User{Name: "Swan"}
	return closure.JSONMe(ctx, fasthttp.StatusAccepted, "user retrieved", user)
}

func PostHandler(ctx *closure.Context) error {
	var user User
	if err := closure.Binder(ctx, &user); err != nil {
		closure.JSONError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return err
	}
	return closure.JSONMe(ctx, fasthttp.StatusAccepted, "user created", user)
}

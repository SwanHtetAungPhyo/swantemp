package closure

import (
	"github.com/SwanHtetAungPhyo/swantemp/utils"
)

type Cluster struct {
	prefix     string
	router     *Router
	middleware []Middleware
	parent     *Cluster
}

func NewCluster(prefix string, router *Router, mw ...Middleware) *Cluster {
	return &Cluster{
		prefix:     utils.NormalizePath(prefix),
		router:     router,
		middleware: mw,
	}
}

func (c *Cluster) Group(subPrefix string, block func(*Cluster)) *Cluster {
	fullPrefix := utils.JoinPaths(c.prefix, subPrefix)
	child := &Cluster{
		prefix:     fullPrefix,
		router:     c.router,
		parent:     c,
		middleware: make([]Middleware, len(c.middleware)),
	}
	_ = copy(child.middleware, c.middleware)
	block(child)
	return child
}

func (c *Cluster) applyMiddleware(handler Handler) Handler {
	middlewares := c.collectMiddleware()

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Apply(handler)
	}
	return handler
}

func (c *Cluster) collectMiddleware() []Middleware {
	var mws []Middleware
	current := c

	for current != nil {
		mws = append(mws, current.middleware...)
		current = current.parent
	}

	return mws
}

func (c *Cluster) registerRoute(method, path string, handler Handler) {
	fullPath := utils.JoinPaths(c.prefix, path)
	wrappedHandler := c.applyMiddleware(handler)
	c.router.Register(method, fullPath, wrappedHandler)
}
func (c *Cluster) Get(path string, handler Handler)  { c.registerRoute("GET", path, handler) }
func (c *Cluster) Post(path string, handler Handler) { c.registerRoute("POST", path, handler) }
func (c *Cluster) Put(path string, handler Handler) {
	c.registerRoute("PUT", path, handler)
}

func (c *Cluster) Patch(path string, handler Handler) {
	c.registerRoute("PATCH", path, handler)
}

func (c *Cluster) Delete(path string, handler Handler) {
	c.registerRoute("DELETE", path, handler)
}

func (c *Cluster) Head(path string, handler Handler) {
	c.registerRoute("HEAD", path, handler)
}

func (c *Cluster) Options(path string, handler Handler) {
	c.registerRoute("OPTIONS", path, handler)
}

func (c *Cluster) Trace(path string, handler Handler) {
	c.registerRoute("TRACE", path, handler)
}

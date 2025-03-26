package closure

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/valyala/fasthttp"
)

//go:embed swagger-ui/*
var swaggerUIAssets embed.FS

// Context wraps fasthttp.RequestCtx to provide additional utilities
type Context struct {
	*fasthttp.RequestCtx
	Params map[string]string
}

// Handler defines the request handler function signature
type Handler func(ctx *Context) error

// routeNode represents a node in the routing tree
type routeNode struct {
	segment    string
	handler    Handler
	children   map[string]*routeNode
	paramChild *routeNode
	wildcard   *routeNode
}

// Router manages HTTP routes using a trie structure
type Router struct {
	methods map[string]*routeNode
}

// NewRouter initializes a new Router instance with Swagger routes
func NewRouter() *Router {
	r := &Router{
		methods: make(map[string]*routeNode),
	}

	r.Register("GET", "/swagger.json", r.serveSwaggerSpec)
	r.Register("GET", "/docs/", r.serveSwaggerUI)
	r.Register("GET", "/docs/*filepath", r.serveSwaggerUI)
	r.Register("GET", "/docs", func(ctx *Context) error {
		ctx.Redirect("/docs/", fasthttp.StatusMovedPermanently)
		return nil
	})

	return r
}

// Register adds a new route and its handler to the router
func (r *Router) Register(method, path string, handler Handler) {
	method = strings.ToUpper(method)
	if r.methods[method] == nil {
		r.methods[method] = &routeNode{children: make(map[string]*routeNode)}
	}

	parts := splitPath(path)
	current := r.methods[method]

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			if current.paramChild == nil {
				current.paramChild = &routeNode{segment: part, children: make(map[string]*routeNode)}
			}
			current = current.paramChild
		} else if part == "*" {
			if current.wildcard == nil {
				current.wildcard = &routeNode{segment: "*", children: make(map[string]*routeNode)}
			}
			current = current.wildcard
		} else {
			if current.children[part] == nil {
				current.children[part] = &routeNode{segment: part, children: make(map[string]*routeNode)}
			}
			current = current.children[part]
		}
	}

	current.handler = handler
}

func (r *Router) ServeHTTP(ctx *fasthttp.RequestCtx) {
	ctxON := &Context{
		RequestCtx: ctx,
	}
	r.serveHTTPHandleFunc(ctxON)
}

// ServeHTTPHandleFunc processes an HTTP request using tree-based route matching
func (r *Router) serveHTTPHandleFunc(ctx *Context) {
	method := string(ctx.Method())
	path := string(ctx.Path())

	root, exists := r.methods[method]
	if !exists {
		JSONError(ctx, fasthttp.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	params := make(map[string]string)
	handler, found := r.matchRoute(root, splitPath(path), params)

	if !found {
		JSONError(ctx, fasthttp.StatusNotFound, "Not Found")
		return
	}

	customCtx := &Context{RequestCtx: ctx.RequestCtx, Params: params}
	if err := handler(customCtx); err != nil {
		JSONError(ctx, fasthttp.StatusInternalServerError, err.Error())
	}
}

// matchRoute recursively traverses the trie to find a matching handler
func (r *Router) matchRoute(node *routeNode, parts []string, params map[string]string) (Handler, bool) {
	if len(parts) == 0 {
		return node.handler, node.handler != nil
	}

	part := parts[0]
	if child, exists := node.children[part]; exists {
		if handler, found := r.matchRoute(child, parts[1:], params); found {
			return handler, true
		}
	}

	if node.paramChild != nil {
		params[node.paramChild.segment[1:]] = part
		if handler, found := r.matchRoute(node.paramChild, parts[1:], params); found {
			return handler, true
		}
	}

	if node.wildcard != nil {
		params["wildcard"] = strings.Join(parts, "/")
		return node.wildcard.handler, node.wildcard.handler != nil
	}

	return nil, false
}

// JSONError sends a JSON error response
func JSONError(ctx *Context, code int, message string) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	ctx.SetBody([]byte(message))
}

// splitPath splits a path into parts while ignoring empty segments
func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 1 && parts[0] == "" {
		return []string{}
	}
	return parts
}

// ServeSwaggerSpec serves the Swagger JSON specification
func (r *Router) serveSwaggerSpec(ctx *Context) error {
	content, err := fs.ReadFile(swaggerUIAssets, "../docs/swagger.json")
	if err != nil {
		ctx.Error("Swagger spec not found", fasthttp.StatusInternalServerError)
		return err
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(content)
	return nil
}

// ServeSwaggerUI serves the Swagger UI files
func (r *Router) serveSwaggerUI(ctx *Context) error {
	filepathVal := ctx.UserValue("filepath")
	filepath, ok := filepathVal.(string)
	if !ok {
		JSONError(ctx, fasthttp.StatusBadRequest, "Invalid path parameter")
		return nil
	}

	if filepath == "" || filepath == "/" {
		filepath = "index.html"
	}
	filepath = strings.TrimPrefix(filepath, "/")

	if strings.Contains(filepath, "..") {
		JSONError(ctx, fasthttp.StatusBadRequest, "Invalid path")
		return nil
	}

	fullPath := path.Join("swagger-ui", filepath)
	content, err := swaggerUIAssets.ReadFile(fullPath)
	if err != nil {
		ctx.NotFound()
		return nil
	}

	switch ext := path.Ext(filepath); ext {
	case ".js":
		ctx.SetContentType("application/javascript")
	case ".css":
		ctx.SetContentType("text/css")
	case ".html":
		ctx.SetContentType("text/html")
	case ".png":
		ctx.SetContentType("image/png")
	default:
		ctx.SetContentType("text/plain")
	}

	_, err = ctx.Write(content)
	if err != nil {
		return err
	}
	return nil
}

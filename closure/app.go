package closure

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	logging "github.com/SwanHtetAungPhyo/swantemp/log"
	"github.com/common-nighthawk/go-figure"
	"github.com/valyala/fasthttp"
)

const CLOSURE = "CLOSURE"

var (
	green = "\033[32m"
	cyan  = "\033[36m"
	reset = "\033[0m"
	bold  = "\033[1m"
)

type App struct {
	router  *Router
	config  *Config
	server  *fasthttp.Server
	options []Option
}

type Config struct {
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxConnsPerIP      int
	MaxRequestsPerConn int
	MaxRequestBodySize int
	ReduceMemoryUsage  bool
	DisableKeepalive   bool
	CloseOnShutdown    bool
}

type Option func(*Config)

// Default server configuration
func defaultConfig() *Config {
	return &Config{
		ReadTimeout:        5 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleTimeout:        30 * time.Second,
		MaxConnsPerIP:      100,
		MaxRequestsPerConn: 10,
		MaxRequestBodySize: 1024 * 1024,
		ReduceMemoryUsage:  true,
		DisableKeepalive:   false,
		CloseOnShutdown:    true,
	}
}

// Functional options for customization
func WithReadTimeout(d time.Duration) Option {
	return func(c *Config) { c.ReadTimeout = d }
}

func WithWriteTimeout(d time.Duration) Option {
	return func(c *Config) { c.WriteTimeout = d }
}

func WithMaxConnsPerIP(max int) Option {
	return func(c *Config) { c.MaxConnsPerIP = max }
}

func WithMaxRequestBodySize(size int) Option {
	return func(c *Config) { c.MaxRequestBodySize = size }
}

func New(opts ...Option) *App {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	return &App{
		router:  NewRouter(),
		config:  config,
		options: opts,
	}
}

func (a *App) ApplyMiddleware(mw ...Middleware) *App {
	return a
}

func (a *App) Cluster(prefix string, block func(*Cluster)) *App {
	cluster := NewCluster(prefix, a.router)
	block(cluster)
	return a
}
func (a *App) Mount(cluster *Cluster) *App {
	for method, rootNode := range cluster.router.methods {
		traverseAndRegister(a.router, method, "", rootNode)
	}
	return a
}

// traverseAndRegister recursively registers all routes from the source router into the target router
func traverseAndRegister(targetRouter *Router, method, currentPath string, node *routeNode) {
	if node.handler != nil {
		targetRouter.Register(method, currentPath, node.handler)
	}

	for segment, childNode := range node.children {
		newPath := path.Join(currentPath, segment)
		traverseAndRegister(targetRouter, method, newPath, childNode)
	}

	if node.paramChild != nil {
		paramPath := path.Join(currentPath, node.paramChild.segment)
		traverseAndRegister(targetRouter, method, paramPath, node.paramChild)
	}

	if node.wildcard != nil {
		wildcardPath := path.Join(currentPath, "*")
		traverseAndRegister(targetRouter, method, wildcardPath, node.wildcard)
	}
}

func (a *App) Info(addr string) {
	fmt.Printf("\n%s╭%s───────────────────────────────────────────────────────╮\n", cyan, bold)
	fmt.Printf("%s│%s", cyan, reset)
	figure.NewColorFigure("Closure", "slant", "cyan", true).Print()
	fmt.Printf("%s│\n%s╰───────────────────────────────────────────────────────╯%s\n\n", cyan, cyan, reset)

	fmt.Printf("%s┌%s CLOSURE WEB FRAMEWORK ───────────────────────────────────────────┐%s\n", cyan, bold, reset)
	fmt.Printf("%s│ Version:   0.0.1%s                                                 │%s\n", cyan, reset, reset)
	fmt.Printf("%s│ PID:       %d%s                                                   │%s\n", cyan, os.Getpid(), reset, reset)
	fmt.Printf("%s│ Go:        %s%s                                                   │%s\n", cyan, runtime.Version()[2:], reset, reset)
	fmt.Printf("%s│ Author:    SWAN HTET AUNG PHYO%s                                   │%s\n", cyan, reset, reset)
	fmt.Printf("%s└───────────────────────────────────────────────────────────────────┘%s\n\n", cyan, reset)

	logging.Info("🚀 %sLaunching server...%s", bold, reset)
	logging.Info("🌐 %sListening on:%s %shttp://localhost:%s%s%s", bold, reset, cyan, reset, green, addr)
	logging.Info("📡 %sNetwork:%s %slocalhost | %s0.0.0.0%s", bold, reset, green, green, reset)
	logging.Info("📂 %sPID:%s %d | %sGo Routines:%s %d", bold, reset, os.Getpid(), bold, reset, runtime.NumGoroutine())
}

func (a *App) Start(addr string) {
	a.Info(addr)
	shutDownChannel := make(chan os.Signal, 1)
	signal.Notify(shutDownChannel, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.server = &fasthttp.Server{
		Name:                  CLOSURE,
		Handler:               a.router.ServeHTTP,
		ReadTimeout:           a.config.ReadTimeout,
		WriteTimeout:          a.config.WriteTimeout,
		IdleTimeout:           a.config.IdleTimeout,
		MaxConnsPerIP:         a.config.MaxConnsPerIP,
		MaxRequestsPerConn:    a.config.MaxRequestsPerConn,
		MaxRequestBodySize:    a.config.MaxRequestBodySize,
		ReduceMemoryUsage:     a.config.ReduceMemoryUsage,
		DisableKeepalive:      a.config.DisableKeepalive,
		CloseOnShutdown:       a.config.CloseOnShutdown,
		NoDefaultServerHeader: true,
		NoDefaultDate:         true,
		NoDefaultContentType:  true,
	}

	serverError := make(chan error, 1)
	go func() {
		err := a.server.ListenAndServe(addr)
		if err != nil {
			logging.Error(err.Error())
			serverError <- err
		}
	}()

	select {
	case stop := <-shutDownChannel:
		logging.Info("Received shutdown signal: %v", stop)
	case err := <-serverError:
		logging.Fatal("Server error %s", err.Error())
		return
	}

	logging.Info("Shutting down the server .....")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 2*time.Second)
	defer shutdownCancel()

	if err := a.server.ShutdownWithContext(shutdownCtx); err != nil {
		logging.Fatal("Error occurred during shutdown %s", err.Error())
	} else {
		logging.Fatal("Server is successfully shut down")
	}
}

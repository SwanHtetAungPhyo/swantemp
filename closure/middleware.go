package closure



type Middleware struct{
	Name string
	Handler func(next Handler) Handler
}

func (m *Middleware) Apply(nextHandler Handler) Handler{
	return m.Handler(nextHandler)
}


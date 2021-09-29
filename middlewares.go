package gate

/*
	Middlewares will be applied as 'Use'd
	i.e. The first middleware to be used will be the
	first one that is called on request, then the second
	and so on - until finally the handler is called.
*/

type Middleware struct {
	ID      string
	Handler func(Handler) Handler
}

var (
	mwares []*Middleware
	mindex = map[string]int{}
)

func (m Middleware) valid() bool {
	if _, ok := mindex[m.ID]; ok || m.ID == "" {
		return false
	}

	return true
}

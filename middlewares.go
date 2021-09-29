package gate

/*
	Middlewares will be called as 'Apply'ed
	i.e. The first middleware to be added via a call to App.Apply()
	will be the first one that is called on request, then the second
	and so on - until finally the handler is called.
*/

type Middleware struct {
	ID      string
	Handler func(Handler) Handler
}

func (m Middleware) valid(app *App) bool {
	if _, ok := app.mwareIndex[m.ID]; ok || m.ID == "" {
		return false
	}

	return true
}

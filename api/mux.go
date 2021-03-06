package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/vincentcr/huecontrol/services"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type Mux struct {
	svc *services.Services
}
type HCContext struct {
	web.C
	Services *services.Services
}

func (c *HCContext) GetUser() (services.User, bool) {
	val, ok := c.Env["user"]
	var user services.User
	if ok {
		user = val.(services.User)
	}

	return user, ok
}

func (c *HCContext) MustGetUser() services.User {
	user, ok := c.GetUser()
	if !ok {
		panic("no user but must get user")
	}
	return user
}

type middleware func(c *HCContext, w http.ResponseWriter, r *http.Request) error
type handler func(c *HCContext, w http.ResponseWriter, r *http.Request) error

func NewMux(svc *services.Services) *Mux {
	return &Mux{svc}
}

func (mux *Mux) Serve() {
	goji.Serve()
	mux.Use(panicRecovery)
}

func (mux *Mux) Use(m middleware) {
	gojiMiddleware := func(c *web.C, h http.Handler) http.Handler {

		handlerFn := func(w http.ResponseWriter, r *http.Request) {
			sc := &HCContext{*c, mux.svc}
			err := m(sc, w, r)
			if err != nil {
				handleError(err, w)
			} else {
				h.ServeHTTP(w, r)
			}
		}

		return http.HandlerFunc(handlerFn)
	}

	goji.Use(gojiMiddleware)
}

func (mux *Mux) Delete(pattern web.PatternType, h handler) {
	goji.Delete(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Head(pattern web.PatternType, h handler) {
	goji.Head(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Get(pattern web.PatternType, h handler) {
	goji.Get(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Options(pattern web.PatternType, h handler) {
	goji.Options(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Patch(pattern web.PatternType, h handler) {
	goji.Patch(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Post(pattern web.PatternType, h handler) {
	goji.Post(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Put(pattern web.PatternType, h handler) {
	goji.Put(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) Trace(pattern web.PatternType, h handler) {
	goji.Trace(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		mux.handleRequest(c, w, r, h)
	})
}

func (mux *Mux) handleRequest(c web.C, w http.ResponseWriter, r *http.Request, h handler) {
	sc := &HCContext{c, mux.svc}
	err := h(sc, w, r)
	if err != nil {
		handleError(err, w)
	}
}

type HttpError struct {
	StatusCode int
	StatusText string
	Data       interface{}
}

func NewHttpError(statusCode int) HttpError {
	statusText := http.StatusText(statusCode)
	return NewHttpErrorWithText(statusCode, statusText)
}

func NewHttpErrorWithText(statusCode int, statusText string) HttpError {
	return HttpError{StatusCode: statusCode, StatusText: statusText}
}

func (err HttpError) Error() string {
	return fmt.Sprintf("%v:%s", err.StatusCode, err.StatusText)
}

func (err HttpError) String() string {
	return err.Error()
}

func handleError(err error, w http.ResponseWriter) {
	var code int
	var text string
	if httpErr, ok := err.(HttpError); ok {
		code = httpErr.StatusCode
		text = httpErr.StatusText
	} else if err == services.ErrNotFound {
		code = http.StatusNotFound
	} else if err == services.ErrUniqueViolation {
		code = http.StatusBadRequest
	} else {
		code = http.StatusInternalServerError
		stack := debug.Stack()
		log.Printf("Internal error: %s\n%s\n", err, stack)
	}

	if text == "" {
		text = http.StatusText(code)
	}

	http.Error(w, text, code)
}

func panicRecovery(c *HCContext, w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if reason := recover(); reason != nil {
			err, ok := reason.(error)
			if !ok {
				err = fmt.Errorf("panic! %v", err)
			}
			handleError(err, w)
		}
	}()

	return nil
}

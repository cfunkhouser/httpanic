// Package httpanic contains a few utilities to streamline HTTP handlers by
// abusing the panic mechanism.
package httpanic

import (
	"errors"
	"net/http"
)

// Reason to panic from inside a HTTP handler.
type Reason struct {
	error

	// Status of the HTTP response which should be served as a result of this
	// Reason to panic.
	Status int

	// Explanation about why we decided to panic.
	Explanation string
}

func (r Reason) Unwrap() error {
	return r.error
}

// Detail about a Reason for panicking.
type Detail func(*Reason)

// WithStatus sets an explicit HTTP status code on the Reason to panic.
func WithStatus(status int) Detail {
	return func(r *Reason) {
		r.Status = status
	}
}

// WithExplanation sets an explicit HTTP status code on the Reason to panic.
func WithExplanation(explanation string) Detail {
	return func(r *Reason) {
		r.Explanation = explanation
	}
}

// Because describes the reason we are deciding to panic. Unless a specific
// status is set using WithStatus, 500 Internal Server Error is assumed.
func Because(e error, deets ...Detail) Reason {
	r := Reason{
		error:  e,
		Status: http.StatusInternalServerError,
	}
	for _, d := range deets {
		d(&r)
	}
	return r
}

// Renderer of Reasons to the client. Used to present the reason for panicking
// to the client in a custom way.
type Renderer func(http.ResponseWriter, Reason)

var defaultRenderer = func(w http.ResponseWriter, reason Reason) {
	// Send the Reason status to the client, and nothing else.
	w.WriteHeader(reason.Status)
}

// reasoner is the interface which describes how to convert an error to a
// Reason. Because is a reasoner.
type reasoner func(error, ...Detail) Reason

// attemptToRecover invokes a Renderer to provide some useful HTTP response to a
// panic in a HTTP handler, but only if the argument to panic is something this
// package knows what to do with.
func attemptToRecover(w http.ResponseWriter, render Renderer, cuz reasoner) {
	switch reason := recover().(type) {
	case Reason:
		render(w, reason)
	case error:
		render(w, cuz(reason))
	case string:
		render(w, cuz(errors.New(reason)))
	default:
		panic(reason)
	}
}

// Gracefully handle any Reason to panic. If the panic is because of an unclear
// reason, it is treated as an Internal Server Error. If anything besides a
// string, error or Reason was given as an argument to panic, the assumption is
// that it was done for a pretty good reason, and this function propagates the
// panic.
func Gracefully(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer attemptToRecover(w, defaultRenderer, Because)
		next.ServeHTTP(w, r)
	})
}

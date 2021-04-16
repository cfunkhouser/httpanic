// Package httpanic contains a few utilities to streamline HTTP handlers by
// abusing the panic mechanism.
package httpanic

import (
	"encoding/json"
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

// MarshalJSON implements custom JSON marshaling for Reason.
func (r Reason) MarshalJSON() ([]byte, error) {
	jr := struct {
		Error       string `json:"error"`
		Explanation string `json:"explanation,omitempty"`
	}{
		Error:       r.Error(),
		Explanation: r.Explanation,
	}
	return json.Marshal(jr)
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
	r := recover()
	// recover returns nil when:
	//   1. It is called outside of a deferred function
	//   2. When the goroutine is not panicking
	//   3. When panic() was called with nil as an argument
	// Since it is impossible to distinguish between these cases, don't even try.
	if r == nil {
		return
	}

	switch reason := r.(type) {
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

// AsJSON renders a Reason for panicking. If any errors are encountered during
// render, this function will panic.
func AsJSON(w http.ResponseWriter, reason Reason) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(reason.Status)
	if err := json.NewEncoder(w).Encode(reason); err != nil {
		panic(err)
	}
}

// GracefullyRender any Reason to panic with the provided Renderer. If the panic
// is because of an unclear reason, it is treated as an Internal Server Error.
// If anything besides a string, error or Reason was given as an argument to
// panic, the assumption is that it was done for a pretty good reason, and this
// function propagates the panic. If anything panics while attempting to handle
// a panic, no attempt will be made to recover from that panic.
func GracefullyRender(next http.Handler, render Renderer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer attemptToRecover(w, render, Because)
		next.ServeHTTP(w, r)
	})
}

// Gracefully handle any Reason to panic by returning an appropriate status
// code, with no response body. See GracefullyRender for additional detail.
func Gracefully(next http.Handler) http.Handler {
	return GracefullyRender(next, defaultRenderer)
}

// Package httpanic contains a few utilities to streamline HTTP handlers by
// abusing the panic mechanism.
package httpanic

import (
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

// Because describes the reason we are deciding to panic.
func Because(e error, and ...Detail) Reason {
	r := Reason{
		error:  e,
		Status: http.StatusInternalServerError,
	}
	for _, a := range and {
		a(&r)
	}
	return r
}

// Gracefully handle any Reason to panic. If the panic is because of an unclear
// reason / error, treat it as an Internal Server Error.
func Gracefully(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)
		next.ServeHTTP(w, r)
	})
}

func recoverFromPanic(w http.ResponseWriter) {
	reason, ok := recover().(Reason)
	if !ok {
		// If we're panicking without Reason, don't try to send anything
		// meaningful.  Further, even though panicking with nil doesn't make
		// much sense, httpanic assumes a panic() is intentional and treats a
		// nil reason the same as a non-Reason error.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send the Reason status to the client. If anything has been returned to
	// the client already, this will do nothing.
	w.WriteHeader(reason.Status)

	// TODO(cfunkhouser): Handle how to return meaningful data to the client.
}

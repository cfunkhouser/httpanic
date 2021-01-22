package httpanic_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/cfunkhouser/httpanic"
)

func validRequest(r *http.Request) bool {
	return false
}

var errBadRequest = errors.New("user made a bad request")

func panickyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	if !validRequest(r) {
		panic(httpanic.Because(errBadRequest, httpanic.WithStatus(http.StatusBadRequest)))
	}
	fmt.Fprintln(w, "Looks good!")
}

func Example() {
	srv := &http.Server{
		// Return any Reasons to panic as JSON to the client, with an
		// appropriate HTTP status code.
		Handler: httpanic.GracefullyRender(http.HandlerFunc(panickyHTTPHandler), httpanic.AsJSON),
	}
	log.Println(srv.ListenAndServe())
}

# `httpanic`

`httpanic` is a Go package which allows for streamlined HTTP handlers by abusing
Go `panic` and `recover`.

## Usage

This package is designed for use as HTTP middleware. Handlers wrapped with `httpanic.Gracefully` can `panic` whenever non-OK responses are to be sent to the client.

```go

var errInvalidRequest = errors.New("invalid request")

func panickyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	if !validRequest(r) {
		// Request validation has failed, so send a JSON-wrapped error payload
		// to the client with a 400 status.
		panic(httpanic.Because(
			errInvalidRequest,
			httpanic.WithStatus(http.StatusBadRequest),
			httpanic.WithExplanation("Request failed validation for example purposes.")))
	}
	fmt.Fprintln(w, "Looks good!")
}

func main() {
	srv := &http.Server{
		Handler: httpanic.GracefullyRender(http.HandlerFunc(panickyHTTPHandler), httpanic.AsJSON),
	}
	log.Println(srv.ListenAndServe())
}
```
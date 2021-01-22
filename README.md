# `httpanic`

`httpanic` is a Go package which allows for streamlined HTTP handlers by abusing
Go `panic` and `recover`.

## Usage

This package is designed for use as HTTP middleware. Handlers wrapped with `httpanic.Gracefully` can `panic` whenever non-OK responses are to be sent to the client.

```go
func panickyHTTPHandler(w http.ResponseWriter, r *http.Request) {
	if !validRequest(r) {
        // Request validation has failed, so send a 400 Bad Request to the
        // client by panicking.
		panic(httpanic.Because(errors.New("invalid request"), httpanic.WithStatus(http.StatusBadRequest)))
	}
	fmt.Fprintln(w, "Looks good!")
}

func main() {
	srv := &http.Server{
		Handler: httpanic.Gracefully(http.HandlerFunc(panickyHTTPHandler)),
	}
	log.Println(srv.ListenAndServe())
}
```
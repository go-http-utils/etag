package etag_test

import (
	"net/http"

	"github.com/go-http-utils/etag"
)

func Example() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Hello World"))
	})

	http.ListenAndServe(":8080", etag.Handler(mux, false))
}

package main

import (
	"net/http"
	"net/http/cgi"
)

func main() {
	cgi.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")

		switch action {
			case "admin":
				manageHandler(w, r)
			default:
				loginHandler(w, r)
		}
	}))
}
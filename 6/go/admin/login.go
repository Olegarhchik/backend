package main

import (
	"fmt"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Вы успешно вошли как админ!")
}
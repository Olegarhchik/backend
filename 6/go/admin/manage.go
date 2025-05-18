package main

import (
	"fmt"
	"net/http"
)

func manageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Измените данные пользователей")
}
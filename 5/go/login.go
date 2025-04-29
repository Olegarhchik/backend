package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("login.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	tmpl.Execute(w, nil)
}
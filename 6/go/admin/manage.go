package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type (
	Application struct {
		Login string
		FullName string
		Phone string
		Email string
		Birthdate string
		Gender string
		ProgLang []string
		Bio string
	}

	Statistics struct {
		ApplQuantity int
		ProgLang []int
	}

	displayResponse struct {
		Applications []Application
		Statistics Statistics
	}
)

func isAuthorized(r *http.Request) bool {
	cookie, err := r.Cookie("accessToken")

	if err != nil {
		return false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("access-token-secret-key"), nil
	})

	if err != nil || !token.Valid {
		return false
	}

	payload, _ := token.Claims.(*jwt.RegisteredClaims)

	return payload.Subject == "admin"
}

func getApplications() ([]Application, error) {
	return []Application{}, nil
}

func getStatistics() (Statistics, error) {
	return Statistics{}, nil
}

func parseApplications(r *http.Request) []Application {
	return []Application{}
}

func updateApplication(appl Application) {
	
}

func removeUser(login string) error {
	return nil
}

func displayHandler (w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("admin/info.html")

	if err != nil {
		fmt.Fprintf(w, "При работе с шаблоном: %v", err)
		return
	}

	appls, err := getApplications()

	if err != nil {
		fmt.Fprintf(w, "При при работе с базой данных: %v", err)
		return
	}

	statistics, err := getStatistics()

	if err != nil {
		fmt.Fprintf(w, "При при работе с базой данных: %v", err)
		return
	}

	statistics.ApplQuantity = len(appls)

	response := displayResponse{
		Applications: appls,
		Statistics: statistics,
	}

	tmpl.Execute(w, response)
}

func updateHandler (w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		tmpl, err := template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "При работе с шаблоном: %v", err)
			return
		}
		
		response := LoginResp{
			Type: "Error",
			Error: "У вас нет прав на выполнение этого действия!",
		}

		tmpl.Execute(w, response)
		return
	}

	appls := parseApplications(r)

	for _, appl := range appls {
		updateApplication(appl)
	}

	displayHandler(w, r)
}

func removeHandler (w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
  		tmpl, err := template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "При работе с шаблоном: %v", err)
			return
		}

		response := LoginResp{
			Type: "Error",
			Error: "У вас нет прав на выполнение этого действия!",
		}

		tmpl.Execute(w, response)
		return
	}

	login := r.FormValue("login")

	err := removeUser(login)

	if err != nil {
		fmt.Fprintf(w, "При работе с базой данных: %v", err)
		return
	}

	displayHandler(w, r)
}

func manageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		typee := r.URL.Query().Get("type")

		switch typee {
			case "update":
				updateHandler(w, r)
			case "remove":
				removeHandler(w, r)
			default:
				displayHandler(w, r)
		}

		return
	}

	tmpl, err := template.ParseFiles("login.html")

	if err != nil {
		fmt.Fprintf(w, "При работе с шаблоном: %v", err)
		return
	}

	tmpl.Execute(w, nil)
}
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"shared"

	"github.com/golang-jwt/jwt/v5"
)

type (
	Application struct {
		shared.Application
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

func (appl *Application) writePL(db *sql.DB) error {
	sel, err := db.Query(`
		SELECT Name
		FROM Abilities abs
		JOIN ProgLang pl
		ON pl.ProgLangID = abs.ProgLangID
		WHERE ApplicationID = ?;
	`, appl.Login)

	if err != nil {
		return err
	}

	defer sel.Close()

	for sel.Next() {
		pl := ""

		err := sel.Scan(&pl)

		if err != nil {
			return err
		}

		appl.ProgLang = append(appl.ProgLang, pl)
	}

	return nil
}

func getApplications() ([]Application, error) {
	appls := []Application{}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return nil, err
	}

	defer db.Close()

	sel, err := db.Query(`
		SELECT *
		FROM Application
	`)

	if err != nil {
		return nil, err
	}

	defer sel.Close()

	for sel.Next() {
		appl := Application{}

		err := sel.Scan(&appl.Login, &appl.FullName, &appl.Phone, &appl.Email, &appl.Birthdate, &appl.Gender, &appl.Bio)

		if err != nil {
			return nil, err
		}

		err = appl.writePL(db)

		if err != nil {
			return appls, err
		}

		appls = append(appls, appl)
	}

	return appls, nil
}

func getStatistics() (Statistics, error) {
	return Statistics{}, nil
}

func parseApplications(r *http.Request) []Application {
	appls := []Application{}

	r.ParseForm()

	for i := 0;; i++ {
		ind := fmt.Sprintf("[%d]", i)

		appl := Application{
			shared.Application {
				Login: r.Form.Get("login" + ind),
				FullName: r.Form.Get("fullname" + ind),
				Phone: r.Form.Get("phone" + ind),
				Email: r.Form.Get("email" + ind),
				Birthdate: r.Form.Get("birthdate" + ind),
				Gender: r.Form.Get("gender" + ind),
				ProgLang: strings.Split(r.Form.Get("proglang" + ind), ", "),
				Bio: r.Form.Get("bio" + ind),
			},
		}

		if appl.Login == "" {
			break
		}

		appls = append(appls, appl)
	}

	return appls
}

func updateApplication(newAppl Application) error {
	oldAppl, err := shared.GetUser(newAppl.Login)

	if err != nil {
		return err
	}

	err = shared.UpdateCols(oldAppl, newAppl.Application, newAppl.Login)

	if err != nil {
		return err
	}

	return nil
}

func removeRow(table string, key string, value string) error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE %s = ?;
	`, table, key), value)

	if err != nil {
		return err
	}

	return nil
}

func removeUser(login string) error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	data := [][]string{
		{"User", "Login"},
		{"Abilities", "ApplicationID"},
		{"Application", "ApplicationID"},
	}

	for _, tr := range data {
		table := tr[0]
		key := tr[1]

		err := removeRow(table, key, login)

		if err != nil {
			return err
		}
	}

	return nil
}

func displayHandler (w http.ResponseWriter, r *http.Request) {
	tmpl := template.New("info.html")

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := tmpl.Funcs(funcMap).ParseFiles("admin/info.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	appls, err := getApplications()

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
		return
	}

	statistics, err := getStatistics()

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
		return
	}

	statistics.ApplQuantity = len(appls)

	response := displayResponse{
		Applications: appls,
		Statistics: statistics,
	}

	err = tmpl.Execute(w, response)

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
	}
}

func updateHandler (w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		tmpl, err := template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
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
		err := updateApplication(appl)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
			return
		}
	}

	displayHandler(w, r)
}

func removeHandler (w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
  		tmpl, err := template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
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
		fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
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
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	tmpl.Execute(w, nil)
}
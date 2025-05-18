package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	jwt "github.com/golang-jwt/jwt/v5"
)

type (
	Admin struct {
		Login string
		Password string
	}

	LoginResp struct {
		Type string
		Error string
	}
)

func dataCorrect(admin Admin) (bool, error) {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return false, err
	}

	defer db.Close()

	sel, err := db.Query(`
		SELECT *
		FROM Admin
	`)

	if err != nil {
		return false, err
	}

	defer sel.Close()

	var login, password string

	for sel.Next() {
		err := sel.Scan(&login, &password)

		if err != nil {
			return false, err
		}
	}

	admin.Password = fmt.Sprintf("%x", sha256.Sum256([]byte(admin.Password)))

	if (admin.Login != login || admin.Password != password) {
		return false, nil
	} else {
		return true, nil 
	}
}

func grantAccessToken(w http.ResponseWriter) error {
	payload := jwt.RegisteredClaims{
		Subject: "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	key := []byte("access-token-secret-key")

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := accessToken.SignedString(key)

	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name: "accessToken",
		Value: t,
	}

	http.SetCookie(w, cookie)

	return nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("admin/login.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
	}

	response := Admin{}

	if r.Method == http.MethodPost {
		response.Login = r.FormValue("login")
		response.Password = r.FormValue("password")

		correct, err := dataCorrect(response)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
		}

		if correct {
			err := grantAccessToken(w)

			if err != nil {
				fmt.Fprintf(w, "Ошибка при создании токена: %v", err)
				return
			}

			manageHandler(w, r)
			return
		} else {
			tmpl, err := template.ParseFiles("login.html")

			if err != nil {
				fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
			}

			response := LoginResp{
				Type: "Error",
				Error: "Пользователь не администратор! Введите ваши данные еще раз",
			}

			tmpl.Execute(w, response)
			return
		}
	}

	tmpl.Execute(w, response)
}
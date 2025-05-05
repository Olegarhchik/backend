package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type (
	InfoResp struct {
		Info Info
		Errors Errors
		Saved bool
	}

	Errors struct {
		FullName string
		Phone string
		Email string
		Birthdate string
		Gender string
		ProgLang string
		Bio string
	}

	Info struct {
		FullName string
		Phone string
		Email string
		Birthdate string
		Gender string
		ProgLang []string
		Bio string
	}

	Payload struct {
		Email string
		jwt.RegisteredClaims
	}
)

func (response InfoResp) IsChecked(lang string) bool {
	info := response.Info

	for _, val := range info.ProgLang {
		if val == lang {
			return true
		}
	}

	return false
}

func extractID(login string) string {
	pattern := `^u\d{6}$`
	re := regexp.MustCompile(pattern)

	if re.MatchString(login) {
		p := 1

		for login[p] == '0' {
			p++
		}
	
		return login[p:]
	} else {
		return "0"	
	}
}

func checkData(user User) (string, error) {
	id := extractID(user.Login)

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return "", err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf("SELECT Password FROM User WHERE Login = '%s'", id))

	if err != nil {
		return "", err
	}

	defer sel.Close()

	var password string

	for sel.Next() {
		err := sel.Scan(&password)

		if err != nil {
			return "", err
		}
	}

	if password == "" || password != fmt.Sprintf("%x",sha256.Sum256([]byte(user.Password))) {
		return "Неверный логин или пароль", nil
	} else {
		return "", nil
	}
}

func getUser(id string) (Info, error) {
	info := Info{}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return info, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf(`
		SELECT FullName, PhoneNumber, APPL.Email, Birthdate, Gender, Biography
		FROM Application APPL
		JOIN User U ON ApplicationID = Login
		WHERE APPL.ApplicationID = '%s';
	`, id))

	if err != nil {
		return info, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&info.FullName, &info.Phone, &info.Email, &info.Birthdate, &info.Gender, &info.Bio)

		if err != nil {
			return info, err
		}
	}

	sel, err = db.Query(fmt.Sprintf(`
		SELECT Name
		FROM ProgLang PL
		JOIN Abilities A
		ON PL.ProgLangID = A.ProgLangID
		WHERE A.ApplicationID = '%s';
	`, id));

	if err != nil {
		return info, err
	}

	defer sel.Close()

	for sel.Next() {
		var pl string
		err := sel.Scan(&pl)

		if err != nil {
			return info, err
		}

		info.ProgLang = append(info.ProgLang, pl)
	}

	return info, nil
}

func getEmail(login string) (string, error) {
	id := extractID(login)
	email := ""

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return email, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf(`
		SELECT Email
		FROM User U
		WHERE Login = '%s';
	`, id));

	if err != nil {
		return email, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&email)

		if err != nil {
			return email, err
		}
	}

	return email, nil
}

func updateUser(response *InfoResp, login string) error {
	info, err := getUser(extractID(login))
	errors := Errors{}

	if err != nil {
		return err
	}

	if info.FullName == "" {
		info.Email, err = getEmail(login)
		errors = Errors{
			FullName: "Заполните поле",
			Phone: "Заполните поле",
			Email: "",
			Birthdate: "Заполните поле",
			Gender: "Выберите пол",
			ProgLang: "Выберите хотя бы один язык",
			Bio: "Заполните поле",
		}

		if err != nil {
			return err
		}
	}

	response.Info = info
	response.Errors = errors

	return nil
}

func grantAccessToken(w http.ResponseWriter, email string) {
	payload := Payload{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	key := []byte("access-token-secret-key")

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := accessToken.SignedString(key)

	if err != nil {
		fmt.Fprintf(w, "Ошибка при создании токена: %v", err)
		return
	}

	cookie := &http.Cookie{
		Name: "accessToken",
		Value: t,
	}

	http.SetCookie(w, cookie)
}

func deleteCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("accessToken")

	if err != nil {
		return
	}

	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	deleteCookie(w, r)

	tmpl, err := template.ParseFiles("login.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	response := LoginResp{}

	if r.Method == http.MethodPost {
		user := User{
			Login: r.FormValue("login"),
			Password: r.FormValue("password"),
			Email: "",
		}

		response.User = user
		response.Error, err = checkData(response.User)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
			return
		}

		if response.Error != "" {
			response.Type = "onLogin"

			tmpl.Execute(w, response)
			return
		}

		response := InfoResp{}

		tmpl, err = template.ParseFiles("info.html")
		
		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
			return
		}

		err = updateUser(&response, user.Login)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
			return
		}

		grantAccessToken(w, response.Info.Email)

		tmpl.Execute(w, response)
		return
	}

	tmpl.Execute(w, response)
}
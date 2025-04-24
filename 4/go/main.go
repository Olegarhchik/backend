package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/url"
	"regexp"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type Form struct {
	FullName string `json:"fullName"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Birthdate string `json:"birthdate"`
	Gender string `json:"gender"`
	ProgLang []string `json:"progLang"`
	Bio string `json:"bio"`
}

type Errors struct {
	FullName string `json:"fullName"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Birthdate string `json:"birthdate"`
	Gender string `json:"gender"`
	ProgLang string `json:"progLang"`
	Bio string `json:"bio"`
}

func (e Errors) ToArray() []string {
	var ans []string

	if e.FullName != "" {
		ans = append(ans, e.FullName)
	}

	if e.Phone != "" {
		ans = append(ans, e.Phone)
	}

	if e.Email != "" {
        ans = append(ans, e.Email)
    }

	if e.Birthdate != "" {
        ans = append(ans, e.Birthdate)
    }
	
    if e.Gender != "" {
		ans = append(ans, e.Gender)
	}

	if e.ProgLang != "" {
		ans = append(ans, e.ProgLang)
	}

	if e.Bio != "" {
       ans = append(ans, e.Bio)
	}

	return ans
}

func (e Errors) hasErrors() bool {
	return !(e.FullName == "" && e.Phone == "" && e.Email == "" && e.Birthdate == "" && e.Gender == "" && e.ProgLang == "" && e.Bio == "")
}

type Response struct {
	Data Form `json:"data"`
	Errors Errors `json:"errors"`
	FormIsValid bool `json:"formIsValid"`
}

func checkErrors(user Form) Errors {
	var pattern string
	var re *regexp.Regexp

	var ans Errors

	pattern = `^([А-Я][а-я]+ ){2}[А-Я][а-я]+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.FullName) {
		ans.FullName = "ФИО должно содержать только буквы русского алфавита"
	}

	pattern = `^(\+7|8)\d{10}$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.Phone) {
		ans.Phone = "Телефон должен содержать 11 цифр, начинающихся с +7 или 8"
	}

	pattern = `^[A-Za-z][\w\.]+@\w+\.[a-z]+$`
	re = regexp.MustCompile(pattern)

	if !re.MatchString(user.Email) {
		ans.Email = "Email должно содержать только буквы латинского алфавита, цифры, знаки . и один знак @"
	}

	if user.Birthdate == "" {
		ans.Birthdate = "Поле Дата рождения не должно быть пустым"
	}
	
	if user.Gender == "" {
		ans.Gender = "Поле Пол не должно быть пустым"
	}

	if len(user.ProgLang) == 0 {
		ans.ProgLang = "Выбор Любимого языка программирования обязателен"
	}

	if user.Bio == "" {
		ans.Bio = "Поле Биография не должно быть пустым"
	}

	return ans
}

func addToDataBase(user Form, w http.ResponseWriter) {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		fmt.Fprintf(w, "Ошибка подключения: %v", err)
		return
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO Application (FullName, PhoneNumber, Email, Birthdate, Gender, Biography) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", user.FullName, user.Phone, user.Email, user.Birthdate, user.Gender, user.Bio))

	if err != nil {
		fmt.Fprintf(w, "Ошибка добавления: %v", err)
		return
	}

	defer insert.Close()

	sel, err := db.Query("SELECT ApplicationID FROM Application ORDER BY ApplicationID DESC LIMIT 1")

	if err != nil {
		fmt.Fprintf(w, "Ошибка извлечения: %v", err)
		return
	}

	defer sel.Close()

	var id int
	for sel.Next() {
		err = sel.Scan(&id)
	}

	if err != nil {
		fmt.Fprintf(w, "Ошибка считывания: %v", err)
		return
	}

	for _, name := range user.ProgLang {
		sel, err := db.Query(fmt.Sprintf("SELECT ProgLangID FROM ProgLang WHERE Name='%s'", name))

		if err != nil {
			fmt.Fprintf(w, "Ошибка извлечения: %v", err)
			return
		}

		defer sel.Close()

		var plId int
		for sel.Next() {
			err = sel.Scan(&plId)
		}

		if err != nil {
			fmt.Fprintf(w, "Ошибка считывания: %v", err)
			return
		}

		insert, err := db.Query(fmt.Sprintf("INSERT INTO Abilities (ApplicationID, ProgLangID) VALUES ('%d', '%d')", id, plId))

		if err != nil {
			fmt.Fprintf(w, "Ошибка добавления: %v", err)
			return
		}

		defer insert.Close()
	}
}

func setCookies(w http.ResponseWriter, response Response) (*http.Cookie, error) {
	responseJSON, err := json.Marshal(response)
	responseEncoded := url.QueryEscape(string(responseJSON))

	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
        Name:     "user",
        Value:    responseEncoded,
    }

	http.SetCookie(w, cookie)

	return cookie, nil
}

func getResponse(r *http.Request) (Response, error) {
	cookie, err := r.Cookie("user")

	if err != nil {
		return Response{}, nil
	}

	var response Response

	responseDecoded, err := url.QueryUnescape(cookie.Value)

	if err != nil {
		return Response{}, err
	}

	err = json.Unmarshal([]byte(responseDecoded), &response)

	if err != nil {
		return Response{}, err
	}

	return response, nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("form.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	response, err := getResponse(r)

	if err != nil {
		fmt.Fprintf(w, "Ошибка декодирования JSON: %v", err)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()

		if err != nil {
			fmt.Fprintf(w, "Произошла ошибка: %v", err)
			return
		}

		user := Form{r.FormValue("full_name"),
						r.FormValue("phone"),
						r.FormValue("email"),
						r.FormValue("birthdate"),
						r.FormValue("gender"),
						r.PostForm["prog_lang[]"],
						r.FormValue("bio")}

		formErr := checkErrors(user)

		response = Response{user, formErr, !formErr.hasErrors()}
		cookie, err := setCookies(w, response)

		if err != nil {
			fmt.Fprintf(w, "Ошибка кодирования JSON: %v", err)
			return
		}

		if response.FormIsValid {
			addToDataBase(user, w)

			cookie.MaxAge = 3600 * 24 * 365
			http.SetCookie(w, cookie)
		}
	}

	tmpl.Execute(w, response)
}

func main() {
	cgi.Serve(http.HandlerFunc(postHandler))
}
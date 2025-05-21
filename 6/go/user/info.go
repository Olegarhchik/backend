package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"

	"shared"

	"github.com/golang-jwt/jwt/v5"
)

func checkInfo(info Info) Errors {
	errors := Errors{}

	pattern := `^([А-Я][а-я]+ ){2}[А-Я][а-я]+$`
	re := regexp.MustCompile(pattern)

	if info.FullName == "" {
		errors.FullName = "Заполните поле"
	} else if !re.MatchString(info.FullName) {
		errors.FullName = "ФИО должно содержать только буквы русского алфавита"
	}

	pattern = `^(\+7|8)\d{10}$`
	re = regexp.MustCompile(pattern)

	if info.Phone == "" {
		errors.Phone = "Заполните поле"
	} else if !re.MatchString(info.Phone) {
		errors.Phone = "Телефон должен содержать 11 цифр, начинающихся с +7 или 8"
	}

	pattern = `^[A-Za-z][\w\.]+@\w+\.[a-z]+$`
	re = regexp.MustCompile(pattern)

	if info.Email == "" {
		errors.Email = "Заполните поле"
	} else if !re.MatchString(info.Email) {
		errors.Email = "Email может содержать только буквы латинского алфавита, цифры, знаки . и один знак @"
	}

	if info.Birthdate == "" {
		errors.Birthdate = "Заполните поле"
	}

	if info.Gender == "" {
		errors.Gender = "Выберите пол"	
	}

	if len(info.ProgLang) == 0 {
		errors.ProgLang = "Выберите хотя бы один язык программирования"
	}

	if info.Bio == "" {
		errors.Bio = "Заполните поле"
	} 
	
	return errors
}

func (response InfoResp) IsValid() bool {
	errors := response.Errors

	return errors.FullName == "" && errors.Phone == "" && errors.Email == "" && errors.Birthdate == "" && errors.Gender == "" && errors.ProgLang == "" && errors.Bio == ""
}

func isAuthorized(r *http.Request, email string) bool {
	cookie, err := r.Cookie("accessToken")

	if err != nil {
		return false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("access-token-secret-key"), nil
	})

	if err != nil || !token.Valid {
		return false
	}

	payload, _ := token.Claims.(*Payload)

	return payload.Email == email
}

func getIdByEmail(email string) (string, error) {
	id := ""

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return id, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf(`
		SELECT Login
		FROM User
		WHERE Email = '%s'
	`, email))

	if err != nil {
		return id, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&id)

		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func updateDB(oldInfo Info, newInfo Info, id string) error {
	if oldInfo.FullName == "" {
		err := shared.InsValues("Application", "ApplicationID, FullName, PhoneNumber, Email, Birthdate, Gender, Biography", id, newInfo.FullName, newInfo.Phone, newInfo.Email, newInfo.Birthdate, newInfo.Gender, newInfo.Bio)

		if err != nil {
			return err
		}

		for _, pl := range newInfo.ProgLang {
			err := shared.InsertPL(id, pl)

			if err != nil {
				return err
			}
		}

		return nil
	}

	shared.UpdateCols(oldInfo, newInfo, id)

	return nil
}

func applyChanges(info Info, id string) error {
	oldInfo, err := shared.GetUser(id)

	if err != nil {
		return err
	}

	err = updateDB(oldInfo, info, id)

	if err != nil {
		return err
	}

	return nil
}

func saveInfoHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("info.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	response := InfoResp{}

	if r.Method == http.MethodPost {
		err := r.ParseForm()

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с формой: %v", err)
			return
		}

		info := Info{
			FullName: r.FormValue("full_name"),
			Phone: r.FormValue("phone"),
			Email: r.FormValue("email"),
			Birthdate: r.FormValue("birthdate"),
			Gender: r.FormValue("gender"),
			ProgLang: r.PostForm["prog_lang[]"],
			Bio: r.FormValue("bio"),
		}
		id, err := getIdByEmail(info.Email)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с базой данных: %v", err)
			return
		}

		errors := checkInfo(info)
		response = InfoResp{info, errors, false}

		if !isAuthorized(r, info.Email) {
			tmpl, err := template.ParseFiles("login.html")

			if err != nil {
				fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
				return
			}

			response := LoginResp{
				User: User{},
				Error: "Обнаружен неавторизованный пользователь! Пожалуйста, войдите в систему",
				Type: "Error",
			}

			tmpl.Execute(w, response)
			return
		}

		if !response.IsValid() {
			tmpl.Execute(w, response)
			return
		}
		
		applyChanges(info, id)
		response.Saved = true
	}

	tmpl.Execute(w, response)
}
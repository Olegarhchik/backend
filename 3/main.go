package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cgi"

	_ "github.com/go-sql-driver/mysql"
)

type FormUser struct {
	fullName, phone, email, birthdate, gender string
	progLang []string
	bio string
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		fmt.Fprintf(w, "Произошла ошибка: %v", err)
		return
	}

	user := FormUser{r.FormValue("full_name"),
					 r.FormValue("phone"),
					 r.FormValue("email"),
					 r.FormValue("birthdate"),
					 r.FormValue("gender"),
					 r.PostForm["prog_lang[]"],
					 r.FormValue("bio")}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		fmt.Fprintf(w, "Ошибка подключения: %v", err)
		return
	}

	defer db.Close()

	insert, err := db.Query(fmt.Sprintf("INSERT INTO Application (FullName, PhoneNumber, Email, Birthdate, Gender, Biography) VALUES ('%s', '%s', '%s', '%s', '%s', '%s')", user.fullName, user.phone, user.email, user.birthdate, user.gender, user.bio))

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

	for _, name := range user.progLang {
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

	fmt.Fprintf(w, "Данные добавлены успешно")
}

func main() {
	cgi.Serve(http.HandlerFunc(postHandler))
}
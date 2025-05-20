package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type (
	RegisterResp struct {
		Error string
	}

	LoginResp struct {
		User User
		Error string
		Type string
	}

	User struct {
		Login string
		Password string
		Email string
	}
)

func insValues(table string, cols string, values ...string) error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	for i, value := range values {
		values[i] = "'" + value + "'"
	}

	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s(%s)
		VALUES (%s)
	`, table, cols, strings.Join(values, ", ")))

	if err != nil {
		return err
	}

	return nil
}

func validateEmail(email string) string {
	pattern := `^[A-Za-z][\w\.]+@\w+\.[a-z]+$`
	re := regexp.MustCompile(pattern)

	if re.MatchString(email) {
		return ""
	} else {
		return "Неверный формат email"
	}
}

func findUserByEmail(email string) (User, error) {
	user := User{}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return user, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf("SELECT Login, Password, Email FROM User WHERE Email='%s'", email))

	if err != nil {
		return user, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&user.Login, &user.Password, &user.Email)

		if err != nil {
			return user, err
		}
	}

	return user, nil
}

func userExists(email string) (string, error) {
	user, err := findUserByEmail(email)

	if err != nil {
		return "", err
	}

	if user == (User{}) {
		return "", nil
	} else {
		return "Пользователь с таким email уже существует", nil
	}
}

func generatePassword(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;':\",./<>?`~"
	password := ""

	for i := 0; i < length; i++ {
		password += string(chars[rand.Intn(len(chars))])
	}

	return password
}

func generateLogin() (string, error) {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return "", err
	}

	defer db.Close()

	lastLogin := "0"

	sel, err := db.Query(`
		SELECT AUTO_INCREMENT
		FROM information_schema.TABLES
		WHERE TABLE_NAME = 'Application';
	`)

	if err != nil {
		return "", err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&lastLogin)

		if err != nil {
			return "", err
		}
	}

	return "u" + strings.Repeat("0", 6 - len(lastLogin)) + lastLogin, nil
}

func createUser(email string) (User, error) {
	login, err := generateLogin()

	if err != nil {
		return User{}, err
	}

	password := generatePassword(8)

	return User{login, password, email}, nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("register.html")

	if err != nil {
		fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
		return
	}

	response := RegisterResp{}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")

		// проверка email
		response.Error = validateEmail(email)

		if response.Error == "" {
			response.Error, err = userExists(email)
		
			if err != nil {
				fmt.Fprintf(w, "Ошибка при заполнении формы регистрации: %v", err)
				return
			}
		}

		if response.Error != "" {
			tmpl.Execute(w, response)
			return
		}

		response := LoginResp{}

		// создание пользователя
		response.User, err = createUser(email)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при создании пользователя: %v", err)
			return
		}

		// добавление пользователя в базу данных
		err = insValues("User", "Password, Email", fmt.Sprintf("%x", sha256.Sum256([]byte(response.User.Password))), response.User.Email)
		response.Type = "postRegister"

		if err != nil {
			fmt.Fprintf(w, "Ошибка при добавлении пользователя в базу данных: %v", err)
			return
		}

		tmpl, err = template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
			return
		}

		tmpl.Execute(w, response)
		return
	}

	tmpl.Execute(w, response)
}
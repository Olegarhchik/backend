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
	Response struct {
		User User
		Error string
	}

	User struct {
		Login string
		Password string
		Email string
	}
)

func (user User) addToDB() error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	ins, err := db.Query(fmt.Sprintf("INSERT INTO User(Password, Email) VALUES ('%x', '%s')", sha256.Sum256([]byte(user.Password)), user.Email))

	if err != nil {
		return err
	}

	defer ins.Close()

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

func increaseByOne(str string) string {
	digits := "0123456789"
	res := ""

	p := len(str) - 1

	for str[p] == '9' {
		res += "0"
		p--
	}

	q := strings.Index(digits, string(str[p])) + 1
	res = string(digits[q]) + res

	return str[:p] + res
}

func generateLogin() (string, error) {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return "", err
	}

	defer db.Close()

	lastLogin := "0"
	sel, err := db.Query("SELECT Login FROM User ORDER BY Login DESC LIMIT 1")

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

	return increaseByOne("u" + strings.Repeat("0", 6 - len(lastLogin)) + lastLogin), nil
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

	response := Response{}

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

		// создание пользователя
		response.User, err = createUser(email)

		if err != nil {
			fmt.Fprintf(w, "Ошибка при создании пользователя: %v", err)
			return
		}

		// добавление пользователя в базу данных
		err = response.User.addToDB()

		if err != nil {
			fmt.Fprintf(w, "Ошибка при добавлении пользователя в базу данных: %v", err)
			return
		}

		tmpl, err = template.ParseFiles("login.html")

		if err != nil {
			fmt.Fprintf(w, "Ошибка при работе с шаблоном: %v", err)
			return
		}
	}

	tmpl.Execute(w, response)
}
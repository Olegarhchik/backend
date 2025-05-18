package main

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
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

func manageHandler(w http.ResponseWriter, r *http.Request) {
	
}
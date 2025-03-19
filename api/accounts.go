package api

import (
	"fmt"
	"net/http"

	"aiolimas/accounts"
)

func CreateAccount(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	err := accounts.CreateAccount(pp["username"].(string), pp["password"].(string))
	if err != nil {
		fmt.Printf("/account/create %s", err.Error())
		wError(w, 500, "Failed to create account: %s", err.Error())
		return
	}

	success(w)
}

func Login(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	username := pp.Get("username", "").(string)
	password := pp.Get("password", "").(string)
	if username == "" || password == "" {
		wError(w, 401, "Please enter credentials")
		return
	}

	_, err := accounts.CkLogin(username, password)
	if err != nil{
		wError(w, 400, "Could not login: %s", err.Error())
		return
	}

	success(w)
}

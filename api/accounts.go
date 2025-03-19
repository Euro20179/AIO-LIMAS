package api

import (
	"aiolimas/accounts"
	"fmt"
	"net/http"
)

func CreateAccount(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	err := accounts.CreateAccount(pp["username"].(string), pp["password"].(string))
	if err != nil{
		fmt.Printf("/account/create %s", err.Error())
		wError(w, 500, "Failed to create account: %s", err.Error())
		return
	}

	success(w)
}

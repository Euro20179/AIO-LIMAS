package api

import (
	"fmt"
	"os"

	"aiolimas/util"
	"aiolimas/accounts"
)

func CreateAccount(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	err := accounts.CreateAccount(pp["username"].(string), pp["password"].(string))
	if err != nil {
		fmt.Printf("/account/create %s", err.Error())
		util.WError(w, 500, "Failed to create account: %s", err.Error())
		return
	}

	success(w)
}

func Login(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	username := pp.Get("username", "").(string)
	password := pp.Get("password", "").(string)
	if username == "" || password == "" {
		util.WError(w, 401, "Please enter credentials")
		return
	}

	_, err := accounts.CkLogin(username, password)
	if err != nil{
		util.WError(w, 400, "Could not login: %s", err.Error())
		return
	}

	success(w)
}

func ListUsers(ctx RequestContext) {
	w := ctx.W
	aioPath := os.Getenv("AIO_DIR")
	users, err := accounts.ListUsers(aioPath)

	if err != nil{
		util.WError(w, 500, "Could not list users")
		return
	}

	w.WriteHeader(200)
	for _, user := range users{
		fmt.Fprintf(w, "%d:%s\n", user.Id, user.Username)
	}
}

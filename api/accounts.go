package api

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"aiolimas/accounts"
	"aiolimas/util"
)

func CreateAccount(ctx RequestContext) {
	data, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		util.WError(ctx.W, 500, "Could not read parameters: %s", err.Error())
		return
	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		util.WError(ctx.W, 500, "Could not read parameters: %s", err.Error())
		return
	}

	username := values.Get("username")
	password := values.Get("password")


	if username == "" || password == "" {
		util.WError(ctx.W, 400, "Username and password cannot be blank")
		return
	}

	err = accounts.CreateAccount(username, password)
	if err != nil {
		fmt.Printf("/account/create %s", err.Error())
		util.WError(ctx.W, 500, "Failed to create account: %s", err.Error())
		return
	}

	success(ctx.W)
}

func DeleteAccount(ctx RequestContext) {
	uid := ctx.Uid

	aioPath := os.Getenv("AIO_DIR")
	err := accounts.DeleteAccount(aioPath, uid)
	if err != nil {
		util.WError(ctx.W, 500, "Failed to delete account: %s", err.Error())
		return
	}
	success(ctx.W)
}

func Logout(ctx RequestContext) {
	ctx.W.Header().Add("Clear-Site-Data", "\"*\"")
	ctx.W.WriteHeader(200)
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

func Username2Id(ctx RequestContext) {
	name := ctx.PP["username"].(string)
	aioPath := os.Getenv("AIO_DIR")
	id, err := accounts.Username2Id(aioPath, name)
	if err != nil {
		util.WError(ctx.W, 500, "could not find user id: %s", err)
		return
	}
	ctx.W.WriteHeader(200)
	fmt.Fprintf(ctx.W, "%d", id)
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

//this endpoint requires the user to be logged in, therefore once we reach here
//the user IS logged in
func AuthCk(ctx RequestContext) {
	success(ctx.W)
}

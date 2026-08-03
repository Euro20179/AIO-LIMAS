package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"time"

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

func RenameAccount(ctx RequestContext) {
	aioPath := os.Getenv("AIO_DIR")
	if err := accounts.RenameUser(aioPath, ctx.Uid, ctx.PP["new-username"].(string)); err != nil {
		util.WError(ctx.W, 500, "Failed to change account name: %s\n", err)
		return
	}
	success(ctx.W)
}

//this endpoint requires the user to be logged in, therefore once we reach here
//the user IS logged in
func AuthCk(ctx RequestContext) {
	ctx.W.WriteHeader(200)
	fmt.Fprintf(ctx.W, "%d", ctx.Uid)
}

var validSyncCodes = map[string]int64{}

func GenSyncCode(ctx RequestContext) {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	code := ""
	for range 10 {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code += string(chars[idx.Int64()])
	}
	validSyncCodes[code] = ctx.Uid

	go func() {
		time.Sleep(5 * time.Minute)
		delete(validSyncCodes, code)
	}()

	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(code))
}

func VerifySyncCode(ctx RequestContext) {
	code := ctx.PP["code"].(string)

	forUid, has := validSyncCodes[code]
	if !has {
		ctx.W.WriteHeader(401)
		ctx.W.Write([]byte("Invalid sync code"))
		return
	}

	delete(validSyncCodes, code)

	if forUid != ctx.Uid {
		ctx.W.WriteHeader(401)
		ctx.W.Write([]byte("uid missmatch (requested uid is different from sync code's associated uid)"))
		return
	}

	hash, err := accounts.CreateAccessHash(forUid, ctx.PP["label"].(string))

	if err != nil {
		util.WError(ctx.W, 500, "Failed to create an access hash: %s\n", err)
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(hash))
}

func DeleteAccessCode(ctx RequestContext) {
	label := ctx.PP["label"].(string)
	if err := accounts.DeleteAccessCode(label); err != nil {
		util.WError(ctx.W, 500, "Failed to delete access code: %s\n", err)
		return
	}

	success(ctx.W)
}

func ListAccesses(ctx RequestContext) {
	list, err := accounts.ListAccessCodes(ctx.Uid)
	if err != nil {
		util.WError(ctx.W, 500, "Failed to list access codes: %s\n", err)
		return
	}

	out, err := json.Marshal(list)

	if err != nil {
		util.WError(ctx.W, 500, "Failed to marshal access codes: %s\n", err)
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write(out)
}

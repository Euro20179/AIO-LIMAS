package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"

	"aiolimas/accounts"
	"aiolimas/db"
	db_types "aiolimas/types"
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
		util.WError(ctx.W, 500, "could not find user id: %s", err.Error())
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
		util.WError(ctx.W, 500, "Failed to change account name: %s\n", err.Error())
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
	var code strings.Builder
	for range 24 {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code.WriteString(string(chars[idx.Int64()]))
	}
	validSyncCodes[code.String()] = ctx.Uid

	go func() {
		time.Sleep(1 * time.Minute)
		delete(validSyncCodes, code.String())
	}()

	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(code.String()))
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

	hash, err := accounts.CreateAccessHash(forUid, ctx.PP["label"].(string))

	if err != nil {
		util.WError(ctx.W, 500, "Failed to create an access hash: %s\n", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(hash))
}

func DeleteAccessCode(ctx RequestContext) {
	label := ctx.PP["label"].(string)
	if err := accounts.DeleteAccessCode(label); err != nil {
		util.WError(ctx.W, 500, "Failed to delete access code: %s\n", err.Error())
		return
	}

	success(ctx.W)
}

func ListAccesses(ctx RequestContext) {
	list, err := accounts.ListAccessCodes(ctx.Uid)
	if err != nil {
		util.WError(ctx.W, 500, "Failed to list access codes: %s\n", err.Error())
		return
	}

	out, err := json.Marshal(list)

	if err != nil {
		util.WError(ctx.W, 500, "Failed to marshal access codes: %s\n", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write(out)
}

func GetAccount(ctx RequestContext) {
	if ctx.Req.Method == "RSS" || (ctx.Req.Method == "GET" && strings.Contains(ctx.Req.Header.Get("Accept"), "application/rss+xml")) || ctx.PP.Has("rss") {
		dctx := actx2dctx(ctx)
		dctx.UID = ctx.PP["id"].(int64)

		eventTy := ctx.PP.Get("type", "").(string)

		events, err := db.GetEvents(dctx, -1, 20, eventTy)
		
		if err != nil {
			util.WError(ctx.W, 500, "Failed to get events: %s\n", err.Error())
			return
		}

		aioPath := os.Getenv("AIO_DIR")
		user, err := accounts.Id2Username(aioPath, ctx.PP["id"].(int64))

		if err != nil {
			util.WError(ctx.W, 500, "Failed to get username: %s\n", err.Error())
			return
		}

		ctx.W.Header().Add("Content-Type", "application/rss+xml")
		ctx.W.Write([]byte(fmt.Sprintf(`<?xml version="1.0"?>
			<rss version="2.0">
				<channel>
					<title>%s's AIO feed</title>
			`, user)))

		items := map[int64]db_types.InfoEntry{}
		userInfos := map[int64]db_types.UserViewingEntry{}

		for _, event := range events {
			var info db_types.InfoEntry
			var user db_types.UserViewingEntry

			if u, has := userInfos[event.ItemId]; has {
				user = u
			} else {
				user, err = db.GetUserViewEntryById(dctx, event.ItemId)
				if err != nil {
					continue
				}
				userInfos[event.ItemId] = user
			}

			if i, has := items[event.ItemId]; has {
				info = i
			} else {
				info, err = db.GetInfoEntryById(dctx, event.ItemId)
				if err != nil {
					continue
				}
				items[event.ItemId] = info
			}

			items[info.ItemId] = info

			if event.Timestamp == 0 {
				continue
			}

			date := time.UnixMilli(event.Timestamp)
			
			ctx.W.Write([]byte(fmt.Sprintf(
				`<item>
					<title>%s</title>
					<link>/html/by-id/%d</link>
					<pubDate>%s</pubDate>
					<description><![CDATA[%s]]></description>
				</item>`,
				event.Event + " " + info.En_Title,
				event.ItemId,
				date.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
				user.Notes,
			)))
		}
		ctx.W.Write([]byte("</channel></rss>"))
	}
}

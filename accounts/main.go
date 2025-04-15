package accounts

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"aiolimas/db"
	"aiolimas/settings"
)

type AccountInfo struct {
	Id int64
	Username string
}

func AccountsDbPath(aioPath string) string {
	return fmt.Sprintf("%s/accounts.db", aioPath)
}

func ListUsers(aioPath string) ([]AccountInfo, error) {
	dbPath := AccountsDbPath(aioPath)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	res, err := conn.Query("SELECT rowid, username FROM accounts")
	if err != nil {
		return nil, err
	}

	var out []AccountInfo
	for res.Next() {
		var acc AccountInfo
		res.Scan(&acc.Id, &acc.Username)
		out = append(out, acc)
	}

	return out, nil
}

func InitAccountsDb(aioPath string) {
	dbPath := AccountsDbPath(aioPath)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	// use implicit rowid column for primary key
	// each user will get a deticated directory for them, under $AIO_DIR/users/<rowid>
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS accounts (
					username TEXT UNIQUE,
					password TEXT
				)`)
	if err != nil {
		panic("Failed to create accounts database\n" + err.Error())
	}
}

func InitializeAccount(aioPath string, username string, hashedPassword string) error {
	accountsDbPath := AccountsDbPath(aioPath)

	conn, err := sql.Open("sqlite3", accountsDbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	res, err := conn.Exec(`INSERT INTO accounts (username, password) VALUES (?, ?)`, username, hashedPassword)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	usersDir := fmt.Sprintf("%s/users/%d", aioPath, id)

	if err := os.MkdirAll(usersDir, 0o700); err != nil {
		return err
	}

	if err := settings.InitUserSettings(id); err != nil {
		return err;
	}

	return db.InitDb(id)
}

func DeleteAccount(aioPath string, uid int64) error {
	accountsDbPath := AccountsDbPath(aioPath)

	conn, err := sql.Open("sqlite3", accountsDbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec("DELETE FROM accounts WHERE rowid = ?", uid)
	if err != nil {
		return err
	}

	usersDir := fmt.Sprintf("%s/users/%d", aioPath, uid)
	return os.RemoveAll(usersDir)
}

func CreateAccount(username string, rawPassword string) error {
	if strings.Contains(username, ":") {
		return errors.New("username may not contain ':'")
	}

	if username == "" {
		return errors.New("username cannot be blank")
	}

	h := sha256.New()
	h.Write([]byte(rawPassword))
	hash := hex.EncodeToString(h.Sum(nil))

	aioPath := os.Getenv("AIO_DIR")

	return InitializeAccount(aioPath, username, hash)
}


func CkLogin(username string, rawPassword string) (string, error) {
	h := sha256.New()
	h.Write([]byte(rawPassword))
	hash := hex.EncodeToString(h.Sum(nil))

	aioPath := os.Getenv("AIO_DIR")
	conn, err := sql.Open("sqlite3", AccountsDbPath(aioPath))
	if err != nil {
		return "", err
	}

	rows, err := conn.Query("SELECT rowid, password FROM accounts WHERE username = ?", username)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		var uid string
		var password string
		err = rows.Scan(&uid, &password)
		if err != nil {
			return "", err
		}

		if(password == hash) {
			return uid, nil
		} else {
			return "", errors.New("invalid password")
		}
	}

	// no account was found in the db
	return "", err
}

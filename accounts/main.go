package accounts

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"aiolimas/db"
	"aiolimas/settings"

	"golang.org/x/crypto/argon2"
)

type AccountInfo struct {
	Id       int64
	Username string
}

func AccountsDbPath(aioPath string) string {
	return fmt.Sprintf("%s/accounts.db", aioPath)
}

func Username2Id(aioPath string, username string) (uint64, error) {
	dbPath := AccountsDbPath(aioPath)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	res, err := conn.Query("SELECT rowid FROM accounts WHERE username = ?", username)

	if err != nil {
		return 0, err
	}

	if err = res.Err(); err != nil {
		println(err)
		return 0, err
	}

	var out uint64 = 0
	res.Next()
	err = res.Scan(&out)
	if err != nil {
		println(err.Error())
	}
	res.Close()
	return out, nil
}

func RenameUser(aioPath string, uid int64, newname string) error {
	dbPath := AccountsDbPath(aioPath)
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	_, err = conn.Exec("UPDATE accounts SET username = ? WHERE rowid = ?", newname, uid)
	if err != nil {
		return err
	}
	return nil
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
	defer res.Close()

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
		return err
	}

	return nil
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

	db.DeleteByUID(uid)

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
	defer conn.Close()

	rows, err := conn.Query("SELECT rowid, password FROM accounts WHERE username = ?", username, hash)
	if err != nil {
		return "", err
	}

	if rows.Next() {
		var uid string
		var password string
		err = rows.Scan(&uid, &password)
		if err != nil {
			rows.Close()
			return "", err
		}

		if password == hash {
			rows.Close()
			return uid, nil
		}
	}

	rows.Close()

	rows, err = conn.Query("SELECT uid, code FROM accesscodes WHERE label = ?", username)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		var uid string
		var code string
		err = rows.Scan(&uid, &code)
		if err != nil {
			return "", err
		}
		data := strings.SplitN(code, "$", 2)
		saltB64 := data[1]
		b64HSAccessHAnswer := data[0]
		salt, _ := base64.RawStdEncoding.DecodeString(saltB64)
		b64HSAccessH := base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(rawPassword), salt, 2, 32*1024, 2, 32))

		if b64HSAccessH == b64HSAccessHAnswer {
			return uid, nil
		} else {
			return "", errors.New("invalid access code")
		}
	}

	// no account was found in the db
	return "", err
}

func twoToThe(n int) *big.Int {
	two := big.NewInt(2)
	one := big.NewInt(1)
	for i := 0; i < n; i++ {
		one = one.Mul(one, two)
	}
	return one
}

func CreateAccessHash(foruser int64, label string) (string, error) {
	aioPath := os.Getenv("AIO_DIR")
	conn, err := sql.Open("sqlite3", AccountsDbPath(aioPath))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS accesscodes (
		uid INTEGER,
		label TEXT UNIQUE,
		code TEXT
	)`)

	if err != nil {
		return "", err
	}

	t1024 := twoToThe(1024)

	accessHash, _ := rand.Int(rand.Reader, twoToThe(256))
	b64AccessH := base64.RawStdEncoding.EncodeToString(accessHash.Bytes())
	salt, err := rand.Int(rand.Reader, t1024)
	if err != nil {
		return "", err
	}

	saltBytes := salt.Bytes()
	b64HSAccessH := base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(b64AccessH), saltBytes, 2, 32*1024, 2, 32))
	b64S := base64.RawStdEncoding.EncodeToString(saltBytes)
	b64HSAccessH_S := fmt.Sprintf("%s$%s", b64HSAccessH, b64S)


	_, err = conn.Exec(`INSERT INTO accesscodes (uid, label, code) VALUES (
		?, ?, ?
	)`, foruser, label, b64HSAccessH_S)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", label, b64AccessH), nil
}

func DeleteAccessCode(label string) error {
	aioPath := os.Getenv("AIO_DIR")
	conn, err := sql.Open("sqlite3", AccountsDbPath(aioPath))
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Exec(`DELETE FROM accesscodes WHERE label = ?`, label)
	if err != nil {
		return err
	}
	return nil
}

func ListAccessCodes(uid int64) ([]string, error) {
	aioPath := os.Getenv("AIO_DIR")
	conn, err := sql.Open("sqlite3", AccountsDbPath(aioPath))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query(`SELECT label from accessCodes WHERE uid = ?`, uid)
	if err != nil {
		return nil, err
	}

	out := []string{}
	for rows.Next() {
		var label string
		err = rows.Scan(&label)
		if err != nil {
			return nil, err
		}
		out = append(out, label)
	}
	return out, nil
}

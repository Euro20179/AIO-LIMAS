package accounts

import (
	"aiolimas/db"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
)

func AccountsDbPath(aioPath string) string {
	return fmt.Sprintf("%s/accounts.db", aioPath)
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
	if err != nil{
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	usersDir := fmt.Sprintf("%s/users/%d", aioPath, id)

	if err := os.MkdirAll(usersDir, 0700); err != nil {
		return err
	}

	return db.InitDb(id)
}

func CreateAccount(username string, rawPassword string) error{
	h := sha256.New()
	h.Write([]byte(rawPassword))
	hash := hex.EncodeToString(h.Sum(nil))

	aioPath := os.Getenv("AIO_DIR")

	return InitializeAccount(aioPath, username, hash)
}

func CkLogin(username string, rawPassword string) (bool, error){
	h := sha256.New()
	h.Write([]byte(rawPassword))
	hash := hex.EncodeToString(h.Sum(nil))

	aioPath := os.Getenv("AIO_DIR")
	conn, err := sql.Open("sqlite3", AccountsDbPath(aioPath))

	if err != nil {
		return false, err
	}

	rows, err := conn.Query("SELECT password FROM accounts WHERE username = ?", username)
	
	if err != nil{
		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		var password string
		err = rows.Scan(&password)

		if err != nil {
			return false, err
		}

		return password == hash, nil
	}

	//no account was found in the db
	return false, err
}

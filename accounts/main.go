package accounts

import (
	"aiolimas/db"
	"database/sql"
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
					username TEXT,
					password TEXT
				)`)

	if err != nil {
		panic("Failed to create accounts database\n" + err.Error())
	}
}

func InitializeAccount(aioPath string, username string, password string) error {
	accountsDbPath := AccountsDbPath(aioPath)

	conn, err := sql.Open("sqlite3", accountsDbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	res, err := conn.Exec(`INSERT INTO accounts (username, password) VALUES (?, ?)`)
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

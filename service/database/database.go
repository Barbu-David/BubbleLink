package database

import (
	"database/sql"
	"errors"
	"fmt"
	// "image/jpeg"
	// "time"
)

type AppDatabase interface {
	Ping() error

	CheckIfUserExists(username string) (bool, error)

	AddNewUser(username string, country string, city string, securityKey string) (int, error)

	GetUserName(userID int) (string, error)
	SetUserName(userID int, username string) error

	//	GetUserPhoto(userID int) (*gif.GIF, error)
	//	SetUserPhoto(userID int, photo *gif.GIF) error

	GetUserKey(userID int) (string, error)
	GetUserID(username string) (int, error)
	//	GetUserIDbyKey(security_key string) (int, error)

}

type appdbimpl struct {
	c *sql.DB
}

func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	usersTableStmt := `CREATE TABLE IF NOT EXISTS Users (
				id INTEGER NOT NULL PRIMARY KEY,
				username TEXT NOT NULL,
				country TEXT NOT NULL,
				city TEXT NOT NULL,
				security_key TEXT NOT NULL,
				jpeg_photo BLOB
				);`
	if _, err := db.Exec(usersTableStmt); err != nil {
		return nil, fmt.Errorf("error creating Users table: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

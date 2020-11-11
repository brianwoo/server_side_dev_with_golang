package database

import (
	"database/sql"
	"log"
	"time"

	"confusion.com/bwoo/config"
)

const dbConfigFile = "../config.json"

var DbConn *sql.DB

func SetupDatabase(config config.Config) {

	connString := config.GetConnString()

	var err error
	DbConn, err = sql.Open(config.DbDriver, connString)
	if err != nil {
		log.Fatal(err)
	}

	DbConn.SetMaxOpenConns(4)
	DbConn.SetMaxIdleConns(4)
	DbConn.SetConnMaxLifetime(60 * time.Second)
}

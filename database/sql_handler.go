package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SQLHandler ...
type SQLHandler struct {
	Db  *gorm.DB
	Err error
}

var dbConn *SQLHandler

// DBOpen は DB connectionを張る。
func DBOpen() {
	dbConn = NewSQLHandler()
}

// DBClose は DB connectionを張る。
func DBClose() {
	sqlDB, _ := dbConn.Db.DB()
	sqlDB.Close()
}

// NewSQLHandler ...
func NewSQLHandler() *SQLHandler {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")

	var db *gorm.DB
	var err error
	if os.Getenv("USE_GAE") != "1" {
		dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?parseTime=true&loc=Asia%2FTokyo"
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Println("mysql", user+":"+password+"@tcp("+host+":"+port+")/"+database+"?parseTime=true&loc=Asia%2FTokyo")
			panic(err)
		}
	} else {
		var (
			instanceConnectionName = os.Getenv("DB_CONNECTION_NAME") // e.g. 'project:region:instance'
		)

		var dbURI string
		dbURI = fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?parseTime=true", user, password, instanceConnectionName, database)

		// dbPool is the pool of database connections.
		db, err = gorm.Open(mysql.Open(dbURI), &gorm.Config{})
		if err != nil {
			panic(err)
		}
	}

	sqlDB, _ := db.DB()
	//コネクションプールの最大接続数を設定。
	sqlDB.SetMaxIdleConns(100)
	//接続の最大数を設定。 nに0以下の値を設定で、接続数は無制限。
	sqlDB.SetMaxOpenConns(100)
	//接続の再利用が可能な時間を設定。dに0以下の値を設定で、ずっと再利用可能。
	sqlDB.SetConnMaxLifetime(100 * time.Second)

	sqlHandler := new(SQLHandler)
	db.Logger.LogMode(4)
	sqlHandler.Db = db

	return sqlHandler
}

// GetDBConn ...
func GetDBConn() *SQLHandler {
	return dbConn
}

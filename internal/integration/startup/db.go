package startup

import (
	"context"
	"database/sql"
	"github.com/tsukaychan/webook/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var db *gorm.DB

func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:for.nothing@tcp(localhost:3309)/webook"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Println("waiting for connect MySQL", err)
		}
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		err = dao.InitTable(db)
		if err != nil {
			panic(err)
		}
	}
	return db
}

package startup

import (
	"context"
	"database/sql"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/lazywoo/mercury/internal/interactive/repository/dao"
)

var db *gorm.DB

func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:for.nothing@tcp(localhost:13316)/mercury"
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

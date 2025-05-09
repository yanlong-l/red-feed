package startup

import (
	"red-feed/internal/repository/dao"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitTestDB() *gorm.DB {
	type DBConfig struct {
		DSN string `yaml:"dsn"`
	}
	var dbCfg = DBConfig{
		DSN: "root:root@tcp(localhost:13316)/webook",
	}
	err := viper.UnmarshalKey("db", &dbCfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(dbCfg.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

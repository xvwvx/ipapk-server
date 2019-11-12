package models

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xvwvx/ipapk-server/conf"
)

var orm *gorm.DB

func InitDB() error {
	var err error
	orm, err = gorm.Open("sqlite3", conf.AppConfig.Database)
	if err != nil {
		return err
	}
	if gin.Mode() != "release" {
		orm.LogMode(true)
	}
	orm.AutoMigrate(&Bundle{})
	return nil
}

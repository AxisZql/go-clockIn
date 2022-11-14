package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	db *gorm.DB
)

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("./db.sqlite3"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		panic(err)
	}
}

type TUser struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"type:varchar(64)"`
	Password string `json:"password" gorm:"type:varchar(512)"`
	Email    string `json:"email" gorm:"type:varchar(32)"`
}

func GetUsers() (users []TUser, err error) {
	err = db.Find(&users).Error
	return
}

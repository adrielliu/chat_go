package dao

import (
	"chat_go/db"
	"errors"
	"time"
)

type User struct {
	Id         int `gorm:"promary_key"`
	UserName   string
	Password   string
	CreateTime time.Time
	//LastLogin time.Time
	db.DbGoChat
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) Add() (userId int, err error) {
	if u.UserName == "" || u.Password == "" {
		return 0, errors.New("user_name or password empty!")
	}
	oUser := u.CheckHaveUserName(u.UserName)
	if oUser.Id > 0 {
		return oUser.Id, nil
	}
	u.CreateTime = time.Now()
	//u.LastLogin = time.Now()
	if err = dbIns.Table(u.TableName()).Create(u).Error; err != nil {
		return 0, err
	}
	return u.Id, nil
}

func (u *User) CheckHaveUserName(userName string) (data User) {
	dbIns.Table(u.TableName()).Where("user_name=?", userName).First(&data)
	return
}

func (u *User) GetUserNameByUserId(userId int) (userName string) {
	var data User
	dbIns.Table(u.TableName()).Where("user_id=?", userId).First(&data)
	return data.UserName
}

//func (u *User) UpdateLogTime() error {
//	dbIns.Table(u.TableName()).Where("user_name=?", u.UserName).First(u)
//	if err := dbIns.Table(u.TableName()).Model(u).Update("lastlogin", time.Now()).Error; err != nil{
//		return err
//	}
//	return nil
//}

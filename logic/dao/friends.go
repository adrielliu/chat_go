package dao

import (
	"chat_go/db"
	"time"
)

type Friends struct {
	Id         int `gorm:"promary_key"`
	UId        int
	FId        int
	CreateTime   time.Time
	Enable     bool
	db.DbGoChat
}

func (f *Friends) TableName() string  {
	return "friends"
}

func (self *Friends) AddNewFriend(fid int) error {
	self.FId = fid
	self.CreateTime = time.Now()
	self.Enable = true
	if err := dbIns.Table(self.TableName()).Create(self).Error; err != nil{
		return err
	}
	return nil
}

func (self *Friends) GetAllFriendsID () ([]int, error) {
	var friends []Friends
	var ids []int
	if err := dbIns.Table(self.TableName()).Where("uid=?", self.UId).Find(&friends).Error; err != nil{
		return ids, err
	}
	for _, friend := range friends{
		ids = append(ids, friend.FId)
	}
	return ids, nil
}
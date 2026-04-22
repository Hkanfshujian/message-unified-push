package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacUserGroup struct {
	IDModel
	Code        string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Description string `json:"description" gorm:"type:varchar(255);default:''"`
	Status      int    `json:"status" gorm:"type:int;default:1"`
}

func GetUserGroupByCode(code string) (*RbacUserGroup, error) {
	var group RbacUserGroup
	if err := db.Where("code = ?", code).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func AddUserGroup(group *RbacUserGroup) error {
	return db.Create(group).Error
}

func AddUserGroupIfNotExists(group *RbacUserGroup) (*RbacUserGroup, error) {
	exist, err := GetUserGroupByCode(group.Code)
	if err == nil {
		return exist, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err = AddUserGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

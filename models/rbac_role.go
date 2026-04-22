package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacRole struct {
	IDModel
	Code        string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Description string `json:"description" gorm:"type:varchar(255);default:''"`
	Status      int    `json:"status" gorm:"type:int;default:1"`
}

func GetRoleByCode(code string) (*RbacRole, error) {
	var role RbacRole
	if err := db.Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func AddRole(role *RbacRole) error {
	return db.Create(role).Error
}

func AddRoleIfNotExists(role *RbacRole) (*RbacRole, error) {
	exist, err := GetRoleByCode(role.Code)
	if err == nil {
		return exist, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err = AddRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

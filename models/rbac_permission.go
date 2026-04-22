package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacPermission struct {
	IDModel
	Code     string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name     string `json:"name" gorm:"type:varchar(100);not null"`
	Type     string `json:"type" gorm:"type:varchar(20);default:'api'"`
	Method   string `json:"method" gorm:"type:varchar(10);default:''"`
	Path     string `json:"path" gorm:"type:varchar(255);default:''"`
	ParentID uint   `json:"parent_id" gorm:"type:integer;default:0"`
	Sort     int    `json:"sort" gorm:"type:int;default:0"`
	Status   int    `json:"status" gorm:"type:int;default:1"`
}

func GetPermissionByCode(code string) (*RbacPermission, error) {
	var permission RbacPermission
	if err := db.Where("code = ?", code).First(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

func AddPermission(permission *RbacPermission) error {
	return db.Create(permission).Error
}

func AddPermissionIfNotExists(permission *RbacPermission) (*RbacPermission, error) {
	exist, err := GetPermissionByCode(permission.Code)
	if err == nil {
		return exist, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err = AddPermission(permission); err != nil {
		return nil, err
	}
	return permission, nil
}

package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacRole struct {
	IDModel
	SoftDeleteModel
	Code        string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Description string `json:"description" gorm:"type:varchar(255);default:''"`
	Status      int    `json:"status" gorm:"type:int;default:1"`
}

func archiveRole(tx *gorm.DB, id uint) error {
	if err := archiveRolePermissions(tx, id); err != nil {
		return err
	}
	if err := archiveRoleUsers(tx, id); err != nil {
		return err
	}
	if err := archiveRoleGroups(tx, id); err != nil {
		return err
	}
	return tx.Unscoped().Delete(&RbacRole{}, id).Error
}

func GetRoleByCode(code string) (*RbacRole, error) {
	var role RbacRole
	if err := db.Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func AddRole(role *RbacRole) error {
	var archived RbacRole
	err := db.Unscoped().Where("code = ?", role.Code).First(&archived).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(role).Error
	}
	if err != nil {
		return err
	}
	if !archived.DeletedAt.Valid {
		return gorm.ErrDuplicatedKey
	}
	role.ID = archived.ID
	return db.Unscoped().Model(&archived).Updates(map[string]interface{}{
		"deleted_at": nil, "name": role.Name, "description": role.Description,
		"status": role.Status, "modified_by": role.ModifiedBy,
	}).Error
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

package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacUserGroup struct {
	IDModel
	SoftDeleteModel
	Code        string `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Description string `json:"description" gorm:"type:varchar(255);default:''"`
	Status      int    `json:"status" gorm:"type:int;default:1"`
}

func archiveUserGroup(tx *gorm.DB, id uint) error {
	if err := archiveGroupRoles(tx, id); err != nil {
		return err
	}
	if err := archiveGroupMembers(tx, id); err != nil {
		return err
	}
	return tx.Unscoped().Delete(&RbacUserGroup{}, id).Error
}

func GetUserGroupByCode(code string) (*RbacUserGroup, error) {
	var group RbacUserGroup
	if err := db.Where("code = ?", code).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func AddUserGroup(group *RbacUserGroup) error {
	var archived RbacUserGroup
	err := db.Unscoped().Where("code = ?", group.Code).First(&archived).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(group).Error
	}
	if err != nil {
		return err
	}
	if !archived.DeletedAt.Valid {
		return gorm.ErrDuplicatedKey
	}
	group.ID = archived.ID
	return db.Unscoped().Model(&archived).Updates(map[string]interface{}{
		"deleted_at": nil, "name": group.Name, "description": group.Description,
		"status": group.Status, "modified_by": group.ModifiedBy,
	}).Error
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

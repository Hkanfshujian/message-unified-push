package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacGroupRole struct {
	IDModel
	SoftDeleteModel
	GroupID uint `json:"group_id" gorm:"index:idx_group_role,unique;not null"`
	RoleID  uint `json:"role_id" gorm:"index:idx_group_role,unique;not null"`
}

func archiveGroupRoles(tx *gorm.DB, groupID uint) error {
	return tx.Unscoped().Where("group_id = ?", groupID).Delete(&RbacGroupRole{}).Error
}

func archiveRoleGroups(tx *gorm.DB, roleID uint) error {
	return tx.Unscoped().Where("role_id = ?", roleID).Delete(&RbacGroupRole{}).Error
}

func replaceGroupRoles(tx *gorm.DB, groupID uint, roleIDs []uint, operator string) error {
	if err := archiveGroupRoles(tx, groupID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		var relation RbacGroupRole
		err := tx.Unscoped().Where("group_id = ? AND role_id = ?", groupID, roleID).First(&relation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = RbacGroupRole{
				IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
				GroupID: groupID, RoleID: roleID,
			}
			if err := tx.Create(&relation).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&relation).Updates(map[string]interface{}{
			"deleted_at": nil, "modified_by": operator,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func AssignRoleToGroupIfNotExists(groupID uint, roleID uint, operator string) error {
	var relation RbacGroupRole
	err := db.Unscoped().Where("group_id = ? AND role_id = ?", groupID, roleID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&RbacGroupRole{
			IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
			GroupID: groupID, RoleID: roleID,
		}).Error
	}
	if err != nil || !relation.DeletedAt.Valid {
		return err
	}
	return db.Unscoped().Model(&relation).Updates(map[string]interface{}{
		"deleted_at": nil, "modified_by": operator,
	}).Error
}

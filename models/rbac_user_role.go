package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacUserRole struct {
	IDModel
	SoftDeleteModel
	UserID int  `json:"user_id" gorm:"index:idx_user_role,unique;not null"`
	RoleID uint `json:"role_id" gorm:"index:idx_user_role,unique;not null"`
}

func archiveUserRoles(tx *gorm.DB, userID int) error {
	return tx.Unscoped().Where("user_id = ?", userID).Delete(&RbacUserRole{}).Error
}

func archiveRoleUsers(tx *gorm.DB, roleID uint) error {
	return tx.Unscoped().Where("role_id = ?", roleID).Delete(&RbacUserRole{}).Error
}

func replaceUserRoles(tx *gorm.DB, userID int, roleIDs []uint, operator string) error {
	if err := archiveUserRoles(tx, userID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		var relation RbacUserRole
		err := tx.Unscoped().Where("user_id = ? AND role_id = ?", userID, roleID).First(&relation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			relation = RbacUserRole{
				IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
				UserID:  userID, RoleID: roleID,
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

func AssignRoleToUserIfNotExists(userID int, roleID uint, operator string) error {
	var relation RbacUserRole
	err := db.Unscoped().Where("user_id = ? AND role_id = ?", userID, roleID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&RbacUserRole{
			IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
			UserID:  userID, RoleID: roleID,
		}).Error
	}
	if err != nil || !relation.DeletedAt.Valid {
		return err
	}
	return db.Unscoped().Model(&relation).Updates(map[string]interface{}{
		"deleted_at": nil, "modified_by": operator,
	}).Error
}

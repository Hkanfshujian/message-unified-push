package models

import (
	"errors"

	"gorm.io/gorm"
)

type RbacRolePermission struct {
	IDModel
	SoftDeleteModel
	RoleID       uint `json:"role_id" gorm:"index:idx_role_permission,unique;not null"`
	PermissionID uint `json:"permission_id" gorm:"index:idx_role_permission,unique;not null"`
}

func archiveRolePermissions(tx *gorm.DB, roleID uint) error {
	return tx.Unscoped().Where("role_id = ?", roleID).Delete(&RbacRolePermission{}).Error
}

func archivePermissionRoles(tx *gorm.DB, permissionID uint) error {
	return tx.Unscoped().Where("permission_id = ?", permissionID).Delete(&RbacRolePermission{}).Error
}

func replaceRolePermissions(tx *gorm.DB, roleID uint, permissionIDs []uint, operator string) error {
	if err := archiveRolePermissions(tx, roleID); err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if err := restoreRolePermission(tx, roleID, permissionID, operator); err != nil {
			return err
		}
	}
	return nil
}

func restoreRolePermission(tx *gorm.DB, roleID uint, permissionID uint, operator string) error {
	var relation RbacRolePermission
	err := tx.Unscoped().Where("role_id = ? AND permission_id = ?", roleID, permissionID).First(&relation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		relation = RbacRolePermission{
			IDModel: IDModel{CreatedBy: operator, ModifiedBy: operator},
			RoleID:  roleID, PermissionID: permissionID,
		}
		return tx.Create(&relation).Error
	}
	if err != nil || !relation.DeletedAt.Valid {
		return err
	}
	return tx.Unscoped().Model(&relation).Updates(map[string]interface{}{
		"deleted_at": nil, "modified_by": operator,
	}).Error
}

func AssignPermissionToRoleIfNotExists(roleID uint, permissionID uint, operator string) error {
	return restoreRolePermission(db, roleID, permissionID, operator)
}

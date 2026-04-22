package models

type RbacRolePermission struct {
	IDModel
	RoleID       uint `json:"role_id" gorm:"index:idx_role_permission,unique;not null"`
	PermissionID uint `json:"permission_id" gorm:"index:idx_role_permission,unique;not null"`
}

func AssignPermissionToRoleIfNotExists(roleID uint, permissionID uint, operator string) error {
	var count int64
	if err := db.Model(&RbacRolePermission{}).Where("role_id = ? AND permission_id = ?", roleID, permissionID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&RbacRolePermission{
		IDModel: IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		RoleID:       roleID,
		PermissionID: permissionID,
	}).Error
}

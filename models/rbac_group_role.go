package models

type RbacGroupRole struct {
	IDModel
	GroupID uint `json:"group_id" gorm:"index:idx_group_role,unique;not null"`
	RoleID  uint `json:"role_id" gorm:"index:idx_group_role,unique;not null"`
}

func AssignRoleToGroupIfNotExists(groupID uint, roleID uint, operator string) error {
	var count int64
	if err := db.Model(&RbacGroupRole{}).Where("group_id = ? AND role_id = ?", groupID, roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&RbacGroupRole{
		IDModel: IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		GroupID: groupID,
		RoleID:  roleID,
	}).Error
}

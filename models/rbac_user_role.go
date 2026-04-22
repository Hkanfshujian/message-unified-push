package models

type RbacUserRole struct {
	IDModel
	UserID int  `json:"user_id" gorm:"index:idx_user_role,unique;not null"`
	RoleID uint `json:"role_id" gorm:"index:idx_user_role,unique;not null"`
}

func AssignRoleToUserIfNotExists(userID int, roleID uint, operator string) error {
	var count int64
	if err := db.Model(&RbacUserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&RbacUserRole{
		IDModel: IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		UserID: userID,
		RoleID: roleID,
	}).Error
}

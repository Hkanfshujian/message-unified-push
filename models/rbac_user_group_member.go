package models

type RbacUserGroupMember struct {
	IDModel
	UserID  int  `json:"user_id" gorm:"index:idx_user_group_member,unique;not null"`
	GroupID uint `json:"group_id" gorm:"index:idx_user_group_member,unique;not null"`
}

func AssignUserToGroupIfNotExists(userID int, groupID uint, operator string) error {
	var count int64
	if err := db.Model(&RbacUserGroupMember{}).Where("user_id = ? AND group_id = ?", userID, groupID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&RbacUserGroupMember{
		IDModel: IDModel{
			CreatedBy:  operator,
			ModifiedBy: operator,
		},
		UserID:  userID,
		GroupID: groupID,
	}).Error
}

package models

func GetPermissionCodesByUserID(userID int) ([]string, error) {
	roleTable := GetSchema(RbacRole{})
	permissionTable := GetSchema(RbacPermission{})
	userRoleTable := GetSchema(RbacUserRole{})
	rolePermissionTable := GetSchema(RbacRolePermission{})
	groupRoleTable := GetSchema(RbacGroupRole{})
	userGroupMemberTable := GetSchema(RbacUserGroupMember{})
	userGroupTable := GetSchema(RbacUserGroup{})

	var directCodes []string
	err := db.Table(permissionTable).
		Select("DISTINCT "+permissionTable+".code").
		Joins("JOIN "+rolePermissionTable+" ON "+rolePermissionTable+".permission_id = "+permissionTable+".id").
		Joins("JOIN "+roleTable+" ON "+roleTable+".id = "+rolePermissionTable+".role_id").
		Joins("JOIN "+userRoleTable+" ON "+userRoleTable+".role_id = "+roleTable+".id").
		Where(userRoleTable+".user_id = ?", userID).
		Where(permissionTable + ".status = 1").
		Where(roleTable + ".status = 1").
		Scan(&directCodes).Error
	if err != nil {
		return nil, err
	}

	var groupCodes []string
	err = db.Table(permissionTable).
		Select("DISTINCT "+permissionTable+".code").
		Joins("JOIN "+rolePermissionTable+" ON "+rolePermissionTable+".permission_id = "+permissionTable+".id").
		Joins("JOIN "+roleTable+" ON "+roleTable+".id = "+rolePermissionTable+".role_id").
		Joins("JOIN "+groupRoleTable+" ON "+groupRoleTable+".role_id = "+roleTable+".id").
		Joins("JOIN "+userGroupTable+" ON "+userGroupTable+".id = "+groupRoleTable+".group_id").
		Joins("JOIN "+userGroupMemberTable+" ON "+userGroupMemberTable+".group_id = "+userGroupTable+".id").
		Where(userGroupMemberTable+".user_id = ?", userID).
		Where(permissionTable + ".status = 1").
		Where(roleTable + ".status = 1").
		Where(userGroupTable + ".status = 1").
		Scan(&groupCodes).Error
	if err != nil {
		return nil, err
	}

	unique := make(map[string]struct{})
	for _, code := range directCodes {
		unique[code] = struct{}{}
	}
	for _, code := range groupCodes {
		unique[code] = struct{}{}
	}

	result := make([]string, 0, len(unique))
	for code := range unique {
		result = append(result, code)
	}
	return result, nil
}

func GetRoleCodesByUserID(userID int) ([]string, error) {
	roleTable := GetSchema(RbacRole{})
	userRoleTable := GetSchema(RbacUserRole{})
	groupRoleTable := GetSchema(RbacGroupRole{})
	userGroupMemberTable := GetSchema(RbacUserGroupMember{})
	userGroupTable := GetSchema(RbacUserGroup{})

	var directCodes []string
	err := db.Table(roleTable).
		Select("DISTINCT "+roleTable+".code").
		Joins("JOIN "+userRoleTable+" ON "+userRoleTable+".role_id = "+roleTable+".id").
		Where(userRoleTable+".user_id = ?", userID).
		Where(roleTable + ".status = 1").
		Scan(&directCodes).Error
	if err != nil {
		return nil, err
	}

	var groupCodes []string
	err = db.Table(roleTable).
		Select("DISTINCT "+roleTable+".code").
		Joins("JOIN "+groupRoleTable+" ON "+groupRoleTable+".role_id = "+roleTable+".id").
		Joins("JOIN "+userGroupTable+" ON "+userGroupTable+".id = "+groupRoleTable+".group_id").
		Joins("JOIN "+userGroupMemberTable+" ON "+userGroupMemberTable+".group_id = "+userGroupTable+".id").
		Where(userGroupMemberTable+".user_id = ?", userID).
		Where(roleTable + ".status = 1").
		Where(userGroupTable + ".status = 1").
		Scan(&groupCodes).Error
	if err != nil {
		return nil, err
	}

	unique := make(map[string]struct{})
	for _, code := range directCodes {
		unique[code] = struct{}{}
	}
	for _, code := range groupCodes {
		unique[code] = struct{}{}
	}

	result := make([]string, 0, len(unique))
	for code := range unique {
		result = append(result, code)
	}
	return result, nil
}

func GetGroupCodesByUserID(userID int) ([]string, error) {
	groupTable := GetSchema(RbacUserGroup{})
	memberTable := GetSchema(RbacUserGroupMember{})

	var codes []string
	err := db.Table(groupTable).
		Select("DISTINCT "+groupTable+".code").
		Joins("JOIN "+memberTable+" ON "+memberTable+".group_id = "+groupTable+".id").
		Where(memberTable+".user_id = ?", userID).
		Where(groupTable + ".status = 1").
		Scan(&codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

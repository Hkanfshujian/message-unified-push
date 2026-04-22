package models

func GetUsers(pageNum int, pageSize int, text string) ([]Auth, error) {
	var users []Auth
	query := db.Model(&Auth{}).Order("id DESC")
	if text != "" {
		like := "%" + text + "%"
		query = query.Where("username LIKE ?", like)
	}
	if pageSize > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserTotal(text string) (int64, error) {
	var total int64
	query := db.Model(&Auth{})
	if text != "" {
		like := "%" + text + "%"
		query = query.Where("username LIKE ?", like)
	}
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

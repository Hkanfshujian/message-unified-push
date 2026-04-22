package models

type LoginLog struct {
	UUIDModel

	ID       uint   `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`
	UserID   int    `json:"user_id" gorm:"type:int;index"`
	Username string `json:"username" gorm:"type:varchar(100);default:'';index"`
	IP       string `json:"ip" gorm:"type:varchar(64);default:'';"`
	UA       string `json:"ua" gorm:"type:varchar(512);default:'';"`
}

func AddLoginLog(userID int, username string, ip string, ua string) error {
	log := LoginLog{UserID: userID, Username: username, IP: ip, UA: ua}
	return db.Create(&log).Error
}

func GetRecentLoginLogs(limit int) ([]LoginLog, error) {
	if limit <= 0 {
		limit = 8
	}
	var logs []LoginLog
	err := db.Model(&LoginLog{}).Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// GetLoginLogs 获取登录日志列表（支持时间范围过滤）
func GetLoginLogs(page, pageSize int, startTime, endTime string) ([]LoginLog, int64, error) {
	var logs []LoginLog
	var total int64

	query := db.Model(&LoginLog{})

	// 时间范围过滤
	if startTime != "" {
		query = query.Where("created_on >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_on <= ?", endTime)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	err = query.Order("created_on DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteOutDateLoginLogs 按数量清理登录日志（保留最近N条）
func DeleteOutDateLoginLogs(keepNum int) (int, error) {
	if keepNum <= 0 {
		return 0, nil
	}

	var threshold LoginLog
	result := db.Model(&LoginLog{}).
		Select("id").
		Order("created_on DESC").
		Offset(keepNum - 1).
		Limit(1).
		First(&threshold)

	if result.Error != nil {
		return 0, nil
	}

	deleteResult := db.Where("id < ?", threshold.ID).Delete(&LoginLog{})
	if deleteResult.Error != nil {
		return 0, deleteResult.Error
	}

	return int(deleteResult.RowsAffected), nil
}

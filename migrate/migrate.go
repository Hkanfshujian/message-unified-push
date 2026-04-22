package migrate

import (
	"errors"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/util"
	"ops-message-unified-push/service/settings_service"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 初始化admin账户
func InitAuthTableData() {
	initSection := "init"
	initAuthKey := "account"
	initAccount := "admin"

	settingO, err := models.GetSettingByKey(initSection, initAuthKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(fmt.Sprintf("查询账号初始化失败！"))
		return
	}
	if settingO.Value == "1" {
		// 已经初始化过
		return
	}

	initAccountPasswd := util.GenerateRandomString(10)

	err = models.AddUser(initAccount, initAccountPasswd)
	if err != nil {
		logrus.Error(fmt.Sprintf("添加初始化admin账号失败！"))
		return
	} else {
		logrus.Info(fmt.Sprintf("初始化admin账号成功！您的账号：%s 密码：%s", initAccount, initAccountPasswd))
	}

	err = models.AddOneSetting(models.Settings{Section: initSection, Key: initAuthKey, Value: "1"})
	if err != nil {
		logrus.Error(fmt.Sprintf("标记admin账号初始化状态失败！err: %s", err.Error()))
		return
	}
}

func Setup() {
	db := models.Setup()
	//defer func(db *gorm.DB) {
	//	err := db.Close()
	//	if err != nil {
	//
	//	}
	//}(db)

	//if setting.AppSetting.InitData != "enable" {
	//	return
	//}

	entry := logrus.WithFields(logrus.Fields{
		"prefix": "[Init Data]",
	})

	tables := []interface{}{
		&models.Auth{},
		&models.SendWays{},
		&models.SendTasksLogs{},
		&models.SendTasksIns{},
		&models.Settings{},
		&models.CronMessages{},
		&models.LoginLog{},
		&models.Template{},
		&models.SendStats{},
		&models.RbacRole{},
		&models.RbacPermission{},
		&models.RbacUserGroup{},
		&models.RbacUserRole{},
		&models.RbacGroupRole{},
		&models.RbacUserGroupMember{},
		&models.RbacRolePermission{},
		&models.AuthIdentity{},
		&models.UserPreference{},
		// MQ 订阅相关表
		&models.MQSource{},
		&models.Subscription{},
		&models.ConsumeLog{},
	}

	for _, table := range tables {
		//tableName := db.NewScope(table).TableName()
		tableName := models.GetSchema(table)
		entry.Infof("Migrate table: %s", tableName)
		err := db.AutoMigrate(table)
		if err != nil {
			entry.Infof("Migrate table erorr: %s", err.Error())
		}
	}

	entry.Infof("Init Account data...")
	InitAuthTableData()
	entry.Infof("Init RBAC seed data...")
	InitRbacSeedData()

	entry.Infof("Init Custom Site data...")
	ss := settings_service.InitSettingService{}
	ss.InitSiteConfig()

	entry.Infof("Init Cron data...")
	ss.InitLogConfig()

	entry.Infof("Init Consume Log Clean Config data...")
	ss.InitConsumeLogConfig()

	entry.Infof("Init Login Log Clean Config data...")
	ss.InitLoginLogConfig()

	entry.Infof("Init Auth Config data...")
	ss.InitAuthConfig()

	entry.Infof("Init Storage Config data...")
	ss.InitStorageConfig()

	entry.Infof("Init MQ Status Policy data...")
	ss.InitMQStatusPolicyConfig()

	entry.Infof("All table data init done.")
}

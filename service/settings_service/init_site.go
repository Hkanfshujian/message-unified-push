package settings_service

import (
	"github.com/sirupsen/logrus"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/constant"
)

// 初始化环境的设置数据
type InitSettingService struct {
}

func (es *InitSettingService) CommonAddSetting(section string, key string, value string) {
	setting, _ := models.GetSettingByKey(section, key)
	if setting.ID <= 0 {
		err := models.AddOneSetting(models.Settings{
			Section: section,
			Key:     key,
			Value:   value,
		})
		if err != nil {
			logrus.Errorf("初始化%s:%s失败", section, key)
		} else {
			logrus.Infof("初始化%s:%s成功", section, key)
		}
	}
}

// InitSiteConfig 初始化、重置站点信息设置
func (es *InitSettingService) InitSiteConfig() {
	section := constant.SiteSettingSectionName
	for key, value := range constant.SiteSiteDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

// InitLogConfig 初始化日志清理设置
func (es *InitSettingService) InitLogConfig() {
	section := constant.LogsCleanSectionName
	for key, value := range constant.LogsCleanDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

// InitConsumeLogConfig 初始化消费日志清理设置
func (es *InitSettingService) InitConsumeLogConfig() {
	section := constant.ConsumeLogsCleanSectionName
	for key, value := range constant.ConsumeLogsCleanDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

// InitLoginLogConfig 初始化登录日志清理设置
func (es *InitSettingService) InitLoginLogConfig() {
	section := constant.LoginLogsCleanSectionName
	for key, value := range constant.LoginLogsCleanDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

func (es *InitSettingService) InitAuthConfig() {
	section := constant.AuthConfigSectionName
	for key, value := range constant.AuthConfigDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

func (es *InitSettingService) InitStorageConfig() {
	section := constant.StorageConfigSectionName
	for key, value := range constant.StorageConfigDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

func (es *InitSettingService) InitMQStatusPolicyConfig() {
	section := constant.MQStatusPolicySectionName
	for key, value := range constant.MQStatusPolicyDefaultValueMap {
		es.CommonAddSetting(section, key, value)
	}
}

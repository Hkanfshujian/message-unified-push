package send_ins_service

import (
	"encoding/json"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"strings"
)

type SendTaskInsService struct {
	ID         string
	Name       string
	CreatedBy  string
	ModifiedBy string
	CreatedOn  string

	PageNum  int
	PageSize int
	Enable   int
}

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendTaskInsService) ValidateDiffIns(ins models.SendTasksIns) (string, interface{}) {
	var empty interface{}
	if ins.WayType == constant.MessageTypeEmail {
		var emailConfig models.InsEmailConfig
		err := json.Unmarshal([]byte(ins.Config), &emailConfig)
		if err != nil {
			return "邮箱auth反序列化失败！", empty
		}

		// 检查是否为动态接收者模式
		var configMap map[string]interface{}
		json.Unmarshal([]byte(ins.Config), &configMap)
		allowMultiRecip, exists := configMap["allowMultiRecip"].(bool)

		// allowMultiRecip=true为动态模式，to_account可以为空（但如果有值也要验证格式）
		if exists && allowMultiRecip {
			// 动态模式：如果to_account为空，直接返回；如果有值，验证邮箱格式
			if emailConfig.ToAccount == "" {
				return "", emailConfig
			}
			// 有to_account值时，验证邮箱格式
			_, Msg := app.CommonPlaygroundValid(emailConfig)
			return Msg, emailConfig
		}

		// 固定模式：必须验证to_account
		_, Msg := app.CommonPlaygroundValid(emailConfig)
		return Msg, emailConfig
	}
	if ins.WayType == constant.MessageTypeDtalk {
		var Config models.InsDtalkConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeQyWeiXin {
		var Config models.InsQyWeiXinConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeFeishu {
		var Config models.InsFeishuConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeCustom {
		var Config models.InsCustomConfig
		err := json.Unmarshal([]byte(ins.Config), &Config)
		if err != nil {
			return "自定义webhook反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(Config)
		return Msg, Config
	}
	if ins.WayType == constant.MessageTypeWeChatOFAccount {
		var Config models.InsWeChatAccountConfig
		err := json.Unmarshal([]byte(ins.Config), &Config)
		if err != nil {
			return "微信公众号发送配置反序列化失败！", empty
		}

		// 检查是否为动态接收者模式
		var configMap map[string]interface{}
		json.Unmarshal([]byte(ins.Config), &configMap)
		allowMultiRecip, exists := configMap["allowMultiRecip"].(bool)

		// allowMultiRecip=true为动态模式，to_account可以为空
		// 历史数据（不存在allowMultiRecip字段）默认为固定模式
		if exists && allowMultiRecip && Config.ToAccount == "" {
			// 动态模式，不验证to_account
			return "", Config
		}

		// 固定模式或有to_account时，进行常规验证
		_, Msg := app.CommonPlaygroundValid(Config)
		return Msg, Config
	}
	if ins.WayType == constant.MessageTypeAliyunSMS {
		var Config models.InsAliyunSMSConfig
		err := json.Unmarshal([]byte(ins.Config), &Config)
		if err != nil {
			return "阿里云短信配置反序列化失败！", empty
		}

		// 检查是否为动态接收者模式
		var configMap map[string]interface{}
		json.Unmarshal([]byte(ins.Config), &configMap)
		allowMultiRecip, exists := configMap["allowMultiRecip"].(bool)

		// allowMultiRecip=true为动态模式，phone_number可以为空，但template_code仍需验证
		if exists && allowMultiRecip && Config.PhoneNumber == "" {
			// 动态模式，只验证template_code
			if Config.TemplateCode == "" {
				return "短信模板CODE不能为空", empty
			}
			return "", Config
		}

		// 固定模式或有phone_number时，进行常规验证
		_, Msg := app.CommonPlaygroundValid(Config)
		return Msg, Config
	}
	if ins.WayType == constant.MessageTypeTelegram {
		var Config models.InsTelegramConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeBark {
		var Config models.InsBarkConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypePushMe {
		var Config models.InsPushMeConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeNtfy {
		var Config models.InsNtfyConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeGotify {
		var Config models.InsGotifyConfig
		return "", Config
	}
	if ins.WayType == constant.MessageTypeQyWeiXinApp {
		// 企业微信应用安全策略：
		// 1) 固定模式（allowMultiRecip!=true）必须配置 to_user，禁止空值兜底 @all
		// 2) 动态模式（allowMultiRecip=true）允许 to_user 为空，由调用方传 recipients
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(ins.Config), &configMap); err != nil {
			return "企业微信应用实例配置反序列化失败", empty
		}
		allowMultiRecip, _ := configMap["allowMultiRecip"].(bool)
		toUser := ""
		if v, ok := configMap["to_user"]; ok {
			if s, ok := v.(string); ok {
				toUser = strings.TrimSpace(s)
			}
		}
		if !allowMultiRecip && toUser == "" {
			return "企业微信应用固定模式下 to_user 不能为空（禁止默认 @all）", empty
		}
		var Config models.InsQyWeiXinAppConfig
		return "", Config
	}
	return "未知的渠道的config校验", empty
}

func (st *SendTaskInsService) ManyAdd(taskIns []models.SendTasksIns) string {

	for _, ins := range taskIns {
		errStr, _ := st.ValidateDiffIns(ins)
		if errStr != "" {
			return errStr
		}
	}
	err := models.ManyAddTaskIns(taskIns)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return ""
}

func (st *SendTaskInsService) AddOne(ins models.SendTasksIns) string {
	errStr, _ := st.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr
	}
	err := models.AddTaskInsOne(ins)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return ""
}

func (st *SendTaskInsService) Delete() error {
	return models.DeleteMsgTaskIns(st.ID)
}

func (st *SendTaskInsService) Update(data map[string]interface{}) error {
	return models.UpdateMsgTaskIns(st.ID, data)
}

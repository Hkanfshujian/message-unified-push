package constant

import "github.com/robfig/cron/v3"

const CleanLogsTaskId = "T-IM1GBswSRY"
const CleanConsumeLogsTaskId = "T-IM1GBswSRX"
const CleanLoginLogsTaskId = "T-IM1GBswSRZ"
const SiteSettingSectionName = "site_config"
const SiteSettingChannelTestMessageKeyName = "channel_test_message"
const SiteSettingSloganInitialEnabledKeyName = "slogan_initial_enabled"

//const SiteSettingTitleKeyName = "title"
//const SiteSettingSloganKeyName = "slogan"
//const SiteSettingLogoKeyName = "logo"

// 站点信息默认值
var SiteSiteDefaultValueMap = map[string]string{
	"title":                                "消息统一推送中台",
	"pagesize":                             "8",
	"slogan":                               "Message Unified Push Platform",
	"login_title":                          "消 息 统 一 推 送 中 台",
	"logo":                                 "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 128 128' width='200' height='200'><defs><linearGradient id='g' x1='0' y1='0' x2='1' y2='1'><stop offset='0%' stop-color='#0EA5E9'/><stop offset='100%' stop-color='#2563EB'/></linearGradient></defs><rect x='8' y='8' width='112' height='112' rx='30' fill='url(#g)'/><path d='M32 94V34h16l16 24 16-24h16v60H80V60L64 84 48 60v34H32z' fill='#FFFFFF'/></svg>",
	"cookie_exp_days":                      "1",
	"theme_color":                          "blue",
	SiteSettingSloganInitialEnabledKeyName: "false",
	SiteSettingChannelTestMessageKeyName:   "This is a test message from ops-message-unified-push.",
}

// 日志清理自定义
const LogsCleanSectionName = "log_config"
const ConsumeLogsCleanSectionName = "consume_log_config"
const LoginLogsCleanSectionName = "login_log_config"
const LogsCleanCronKeyName = "cron"
const LogsCleanKeepKeyName = "keep_num"
const LogsCleanEnabledKeyName = "enabled"

// 任务日志清理默认值
var LogsCleanDefaultValueMap = map[string]string{
	"cron":     "1 0 * * *",
	"keep_num": "1000",
	"enabled":  "true",
}

// 消费日志清理默认值
var ConsumeLogsCleanDefaultValueMap = map[string]string{
	"cron":     "1 0 * * *",
	"keep_num": "1000",
	"enabled":  "false",
}

// 登录日志清理默认值
var LoginLogsCleanDefaultValueMap = map[string]string{
	"cron":     "1 0 * * *",
	"keep_num": "1000",
	"enabled":  "false",
}

const AuthConfigSectionName = "auth_config"
const StorageConfigSectionName = "storage_config"
const MQStatusPolicySectionName = "mq_status_policy"
const MQStatusPolicyEnabledKeyName = "enabled"
const MQStatusPolicyIntervalSecondsKeyName = "interval_seconds"
const MQStatusPolicyLogLevelKeyName = "log_level"

var AuthConfigDefaultValueMap = map[string]string{
	// 本地用户配置
	"register_enabled":         "false",
	"local_default_group_code": "default_group",
	// Casdoor 独立配置
	"casdoor_enabled":            "false",
	"casdoor_endpoint":           "",
	"casdoor_client_id":          "",
	"casdoor_client_secret":      "",
	"casdoor_redirect_uri":       "",
	"casdoor_auth_path":          "/login/oauth/authorize",
	"casdoor_token_path":         "/api/login/oauth/access_token",
	"casdoor_userinfo_path":      "/api/get-account",
	"casdoor_logout_path":        "/api/logout",
	"casdoor_auto_create_user":   "true",
	"casdoor_default_group_code": "default_group",
	"casdoor_button_text":        "企微登录",
	"casdoor_button_icon":        "",
}

var StorageConfigDefaultValueMap = map[string]string{
	"storage_provider":      "local",
	"s3_endpoint":           "",
	"s3_region":             "",
	"s3_bucket":             "",
	"s3_access_key":         "",
	"s3_secret_key":         "",
	"s3_use_ssl":            "true",
	"s3_public_base_url":    "",
	"s3_proxy_public_read":  "true",
	"s3_object_key_prefix":  "auth-icons",
	"local_sub_path":        "uploads",
	"default_storage_id":    "10000001",
	"storage_profiles_json": "[{\"id\":\"10000001\",\"name\":\"本地默认存储\",\"provider\":\"local\",\"enabled\":true,\"local_sub_path\":\"uploads\",\"s3_endpoint\":\"\",\"s3_region\":\"\",\"s3_bucket\":\"\",\"s3_access_key\":\"\",\"s3_secret_key\":\"\",\"s3_use_ssl\":true,\"s3_public_base_url\":\"\",\"s3_proxy_public_read\":true,\"s3_object_key_prefix\":\"auth-icons\"}]",
}

// 消息队列状态更新策略默认值
var MQStatusPolicyDefaultValueMap = map[string]string{
	MQStatusPolicyEnabledKeyName:         "false", // false=手动，true=自动
	MQStatusPolicyIntervalSecondsKeyName: "300",   // 自动模式下的探测频率（秒）
	MQStatusPolicyLogLevelKeyName:        "info",  // debug/info/warn/error
}

const AboutSectionName = "about"

// 消息格式类型常量
const (
	FormatTypeText     = "text"
	FormatTypeHTML     = "html"
	FormatTypeMarkdown = "markdown"
)

// 消息类型常量
const (
	MessageTypeEmail           = "Email"
	MessageTypeDtalk           = "Dtalk"
	MessageTypeQyWeiXin        = "QyWeiXin"
	MessageTypeFeishu          = "Feishu"
	MessageTypeCustom          = "Custom"
	MessageTypeWeChatOFAccount = "WeChatOFAccount"
	MessageTypeAliyunSMS       = "AliyunSMS"
	MessageTypeTelegram        = "Telegram"
	MessageTypeBark            = "Bark"
	MessageTypePushMe          = "PushMe"
	MessageTypeNtfy            = "Ntfy"
	MessageTypeGotify          = "Gotify"
	MessageTypeQyWeiXinApp     = "QyWeiXinApp"
)

// 限制goroutine的最大数量
var MaxSendTaskSemaphoreChan = make(chan string, 2048)

// cron消息任务内存缓存map
var CronMsgIdMapMemoryCache = make(map[string]cron.EntryID)

package send_way_service

import (
	"encoding/json"
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/app"
	"ops-message-unified-push/pkg/constant"
	"ops-message-unified-push/pkg/message"
	"strings"
)

const defaultChannelTestTitle = "Test Message"

type SendWay struct {
	ID         string
	Name       string
	Type       string
	CreatedBy  string
	ModifiedBy string
	Auth       string
	CreatedOn  string

	PageNum  int
	PageSize int
}

// WayValidator 渠道验证接口
type WayValidator interface {
	Validate(authJson string) (string, interface{})
}

// WayTester 渠道测试接口
type WayTester interface {
	Test() (string, string)
}

// 渠道注册表
var (
	validatorRegistry = map[string]func() WayValidator{
		constant.MessageTypeEmail:           func() WayValidator { return &WayDetailEmail{} },
		constant.MessageTypeDtalk:           func() WayValidator { return &WayDetailDTalk{} },
		constant.MessageTypeQyWeiXin:        func() WayValidator { return &WayDetailQyWeiXin{} },
		constant.MessageTypeFeishu:          func() WayValidator { return &WayDetailFeishu{} },
		constant.MessageTypeCustom:          func() WayValidator { return &WayDetailCustom{} },
		constant.MessageTypeWeChatOFAccount: func() WayValidator { return &WeChatOFAccount{} },
		constant.MessageTypeAliyunSMS:       func() WayValidator { return &WayDetailAliyunSMS{} },
		constant.MessageTypeTelegram:        func() WayValidator { return &WayDetailTelegram{} },
		constant.MessageTypeBark:            func() WayValidator { return &WayDetailBark{} },
		constant.MessageTypePushMe:          func() WayValidator { return &WayDetailPushMe{} },
		constant.MessageTypeNtfy:            func() WayValidator { return &WayDetailNtfy{} },
		constant.MessageTypeGotify:          func() WayValidator { return &WayDetailGotify{} },
		constant.MessageTypeQyWeiXinApp:     func() WayValidator { return &WayDetailQyWeiXinApp{} },
	}
	testerRegistry = map[string]func(interface{}) WayTester{
		constant.MessageTypeEmail:           func(m interface{}) WayTester { return m.(*WayDetailEmail) },
		constant.MessageTypeDtalk:           func(m interface{}) WayTester { return m.(*WayDetailDTalk) },
		constant.MessageTypeQyWeiXin:        func(m interface{}) WayTester { return m.(*WayDetailQyWeiXin) },
		constant.MessageTypeFeishu:          func(m interface{}) WayTester { return m.(*WayDetailFeishu) },
		constant.MessageTypeCustom:          func(m interface{}) WayTester { return m.(*WayDetailCustom) },
		constant.MessageTypeWeChatOFAccount: func(m interface{}) WayTester { return m.(*WeChatOFAccount) },
		constant.MessageTypeAliyunSMS:       func(m interface{}) WayTester { return m.(*WayDetailAliyunSMS) },
		constant.MessageTypeTelegram:        func(m interface{}) WayTester { return m.(*WayDetailTelegram) },
		constant.MessageTypeBark:            func(m interface{}) WayTester { return m.(*WayDetailBark) },
		constant.MessageTypePushMe:          func(m interface{}) WayTester { return m.(*WayDetailPushMe) },
		constant.MessageTypeNtfy:            func(m interface{}) WayTester { return m.(*WayDetailNtfy) },
		constant.MessageTypeGotify:          func(m interface{}) WayTester { return m.(*WayDetailGotify) },
		constant.MessageTypeQyWeiXinApp:     func(m interface{}) WayTester { return m.(*WayDetailQyWeiXinApp) },
	}
)

func getChannelTestMessage() string {
	fallback := constant.SiteSiteDefaultValueMap[constant.SiteSettingChannelTestMessageKeyName]
	if strings.TrimSpace(fallback) == "" {
		fallback = "This is a test message from ops-message-unified-push."
	}
	setting, _ := models.GetSettingByKey(constant.SiteSettingSectionName, constant.SiteSettingChannelTestMessageKeyName)
	if setting.ID > 0 {
		if message := strings.TrimSpace(setting.Value); message != "" {
			return message
		}
	}
	return fallback
}

// WayDetailEmail 邮箱渠道明细字段
type WayDetailEmail struct {
	Server   string `json:"server" validate:"required,max=50" label:"SMTP服务地址"`
	Port     int    `json:"port" validate:"required,max=65535" label:"SMTP服务端口"`
	Account  string `json:"account" validate:"required,email" label:"邮箱账号"`
	Passwd   string `json:"passwd" validate:"required,max=50" label:"邮箱密码"`
	FromName string `json:"from_name" validate:"max=50" label:"发信人名称"`
}

func (w *WayDetailEmail) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "邮箱auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailEmail) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var emailer message.EmailMessage
	emailer.Init(w.Server, w.Port, w.Account, w.Passwd, w.FromName)
	errMsg := emailer.SendTextMessage(w.Account, "test email", testMsg)
	return errMsg, ""
}

// WayDetailDTalk 钉钉渠道明细字段
type WayDetailDTalk struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"钉钉access_token"`
	Keys        string `json:"keys" validate:"max=200" label:"钉钉关键字"`
	Secret      string `json:"secret" validate:"max=100" label:"钉钉加签秘钥"`
}

func (w *WayDetailDTalk) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "钉钉auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailDTalk) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Dtalk{
		AccessToken: w.AccessToken,
		Secret:      w.Secret,
	}
	res, err := cli.SendMessageText(testMsg + w.Keys)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailQyWeiXin 企业微信渠道明细字段
type WayDetailQyWeiXin struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"企业微信access_token"`
}

func (w *WayDetailQyWeiXin) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "企业微信auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailQyWeiXin) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.QyWeiXin{
		AccessToken: w.AccessToken,
	}
	res, err := cli.SendMessageText(testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailFeishu 飞书渠道明细字段
type WayDetailFeishu struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"飞书access_token"`
	Keys        string `json:"keys" validate:"max=200" label:"飞书关键字"`
	Secret      string `json:"secret" validate:"max=100" label:"飞书加签秘钥"`
}

func (w *WayDetailFeishu) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "飞书auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailFeishu) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Feishu{
		AccessToken: w.AccessToken,
		Secret:      w.Secret,
	}
	res, err := cli.SendMessageText(testMsg + w.Keys)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailCustom 自定义渠道
type WayDetailCustom struct {
	Webhook string `json:"webhook" validate:"required,max=200" label:"自定义的webhook地址"`
	Body    string `json:"body" validate:"max=2000" label:"自定义的请求体"`
}

func (w *WayDetailCustom) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "自定义参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailCustom) Test() (string, string) {
	return "自定义webhook不用测试运行，请直接添加", ""
}

// WeChatOFAccount 微信公众号
type WeChatOFAccount struct {
	AppID     string `json:"appID" validate:"required,max=200" label:"微信公众号id"`
	APPSecret string `json:"appsecret" validate:"max=2000" label:"微信公众号秘钥"`
	TempID    string `json:"tempid" validate:"max=2000" label:"模板消息id"`
}

func (w *WeChatOFAccount) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "微信公众号反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WeChatOFAccount) Test() (string, string) {
	return "微信公众号模板消息不用测试运行，请直接添加", ""
}

// WayDetailAliyunSMS 阿里云短信渠道明细字段
type WayDetailAliyunSMS struct {
	AccessKeyId     string `json:"access_key_id" validate:"required,max=100" label:"AccessKeyId"`
	AccessKeySecret string `json:"access_key_secret" validate:"required,max=100" label:"AccessKeySecret"`
	SignName        string `json:"sign_name" validate:"required,max=50" label:"短信签名"`
	RegionId        string `json:"region_id" validate:"required,max=50" label:"区域ID"`
}

func (w *WayDetailAliyunSMS) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "阿里云短信auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailAliyunSMS) Test() (string, string) {
	return "阿里云短信不用测试运行，请直接添加", ""
}

// WayDetailTelegram Telegram机器人渠道明细字段
type WayDetailTelegram struct {
	BotToken string `json:"bot_token" validate:"required,max=100" label:"Telegram Bot Token"`
	ChatID   string `json:"chat_id" validate:"required,max=50" label:"Chat ID"`
	ApiHost  string `json:"api_host" validate:"max=200" label:"自定义API地址"`
	ProxyURL string `json:"proxy_url" validate:"max=200" label:"代理地址"`
}

func (w *WayDetailTelegram) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "Telegram参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailTelegram) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Telegram{
		BotToken: w.BotToken,
		ChatID:   w.ChatID,
		ApiHost:  w.ApiHost,
		ProxyURL: w.ProxyURL,
	}
	res, err := cli.SendMessageText(testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailBark Bark渠道明细字段
type WayDetailBark struct {
	PushKey string `json:"push_key" validate:"required,max=200" label:"Bark Push Key"`
	Archive string `json:"archive" validate:"max=10" label:"推送是否存档"`
	Group   string `json:"group" validate:"max=50" label:"推送分组"`
	Sound   string `json:"sound" validate:"max=50" label:"推送声音"`
	Icon    string `json:"icon" validate:"max=200" label:"推送图标"`
	Level   string `json:"level" validate:"max=20" label:"推送时效性"`
	URL     string `json:"url" validate:"max=200" label:"推送跳转URL"`
	Key     string `json:"key" validate:"max=100" label:"加密Key"`
	IV      string `json:"iv" validate:"max=100" label:"加密IV"`
}

func (w *WayDetailBark) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "Bark参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailBark) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Bark{
		PushKey: w.PushKey,
		Archive: w.Archive,
		Group:   w.Group,
		Sound:   w.Sound,
		Icon:    w.Icon,
		Level:   w.Level,
		URL:     w.URL,
		Key:     w.Key,
		IV:      w.IV,
	}
	res, err := cli.Request(defaultChannelTestTitle, testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailPushMe PushMe渠道明细字段
type WayDetailPushMe struct {
	PushKey string `json:"push_key" validate:"required,max=100" label:"PushMe Push Key"`
	URL     string `json:"url" validate:"max=200" label:"自定义API地址"`
	Date    string `json:"date" validate:"max=50" label:"日期"`
	Type    string `json:"type" validate:"max=50" label:"类型"`
}

func (w *WayDetailPushMe) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "PushMe参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailPushMe) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.PushMe{
		PushKey: w.PushKey,
		URL:     w.URL,
		Date:    w.Date,
		Type:    w.Type,
	}
	res, err := cli.Request(defaultChannelTestTitle, testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), res
	}
	return "", res
}

// WayDetailNtfy Ntfy推送渠道明细字段
type WayDetailNtfy struct {
	Url      string `json:"url" validate:"max=200" label:"自定义API地址"`
	Topic    string `json:"topic" validate:"required,max=100" label:"Topic"`
	Priority string `json:"priority" validate:"max=10" label:"优先级"`
	Icon     string `json:"icon" validate:"max=200" label:"图标URL"`
	Token    string `json:"token" validate:"max=100" label:"Token"`
	Username string `json:"username" validate:"max=100" label:"用户名"`
	Password string `json:"password" validate:"max=100" label:"密码"`
	Actions  string `json:"actions" validate:"max=500" label:"Actions"`
}

func (w *WayDetailNtfy) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "Ntfy参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailNtfy) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Ntfy{
		Url:      w.Url,
		Topic:    w.Topic,
		Priority: w.Priority,
		Icon:     w.Icon,
		Token:    w.Token,
		Username: w.Username,
		Password: w.Password,
		Actions:  w.Actions,
	}
	res, err := cli.Request(defaultChannelTestTitle, testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailGotify Gotify推送渠道明细字段
type WayDetailGotify struct {
	Url      string `json:"url" validate:"required,url" label:"Gotify服务地址"`
	Token    string `json:"token" validate:"required,max=100" label:"Token"`
	Priority int    `json:"priority" validate:"max=10" label:"优先级"`
}

func (w *WayDetailGotify) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "Gotify参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailGotify) Test() (string, string) {
	testMsg := getChannelTestMessage()
	var cli = message.Gotify{
		Url:      w.Url,
		Token:    w.Token,
		Priority: w.Priority,
	}
	res, err := cli.Request(defaultChannelTestTitle, testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailQyWeiXinApp 企业微信应用明细字段
type WayDetailQyWeiXinApp struct {
	CorpID     string `json:"corp_id" validate:"required,max=100" label:"企业ID"`
	CorpSecret string `json:"corp_secret" validate:"required,max=200" label:"应用Secret"`
	AgentID    string `json:"agent_id" validate:"required,max=100" label:"应用ID"`
	ToUser     string `json:"to_user" validate:"max=200" label:"接收者ID"`
}

func (w *WayDetailQyWeiXinApp) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "企业微信应用auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, w
}

func (w *WayDetailQyWeiXinApp) Test() (string, string) {
	testMsg := getChannelTestMessage()
	toUser := strings.TrimSpace(w.ToUser)
	if toUser == "" {
		return "测试接收者不能为空，请填写具体企微用户ID", ""
	}
	appClient := message.QyWeiXinApp{
		CorpID:     w.CorpID,
		CorpSecret: w.CorpSecret,
		AgentID:    w.AgentID,
		ToUser:     toUser,
	}
	msgId, err := appClient.SendText(testMsg)
	if err != nil {
		return err.Error(), ""
	}
	return "", msgId
}

func (sw *SendWay) GetByID() (interface{}, error) {
	return models.GetWayByID(sw.ID)
}

func (sw *SendWay) NameIsExist(method string) error {
	way, err := models.GetWayByName(sw.Name)
	if err != nil {
		return err
	}
	if method == "add" {
		if len(way.ID) > 0 {
			return fmt.Errorf("已经存在同名的渠道名：%s", way.Name)
		}
	} else {
		// 只允许当前的重名
		if len(way.ID) > 0 && way.ID != sw.ID {
			return fmt.Errorf("已经存在同名的渠道名：%s", way.Name)
		}
	}
	return nil
}

func (sw *SendWay) Add() error {
	err := sw.NameIsExist("add")
	if err != nil {
		return err
	}
	return models.AddSendWay(sw.Name, sw.Auth, sw.Type, sw.CreatedBy, sw.ModifiedBy)
}

func (sw *SendWay) Edit() error {
	err := sw.NameIsExist("edit")
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	data["modified_by"] = sw.ModifiedBy
	data["name"] = sw.Name
	data["auth"] = sw.Auth
	return models.EditSendWay(sw.ID, data)
}

func (sw *SendWay) Delete() error {
	names, err := models.FindWayUsageNames(sw.ID)
	if err != nil {
		return err
	}
	if len(names) > 0 {
		return fmt.Errorf("已经存在使用该渠道的模板，删除失败！模板名：%s", strings.Join(names, ", "))
	}
	return models.DeleteMsgWay(sw.ID)
}

func (sw *SendWay) Count() (int64, error) {
	return models.GetSendWaysTotal(sw.Name, sw.Type, sw.getMaps())
}

func (sw *SendWay) GetAll() ([]models.SendWays, error) {
	tags, err := models.GetSendWays(sw.PageNum, sw.PageSize, sw.Name, sw.Type, sw.getMaps())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (sw *SendWay) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}

// getValidator 根据渠道类型获取对应的验证器
func (sw *SendWay) getValidator() WayValidator {
	factory, exists := validatorRegistry[sw.Type]
	if !exists {
		return nil
	}
	return factory()
}

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendWay) ValidateDiffWay() (string, interface{}) {
	var empty interface{}
	validator := sw.getValidator()
	if validator == nil {
		return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), empty
	}
	return validator.Validate(sw.Auth)
}

// TestSendWay 尝试带发信测试连通性
func (sw *SendWay) TestSendWay(msgObj interface{}) (string, string) {
	factory, exists := testerRegistry[sw.Type]
	if !exists {
		return fmt.Sprintf("未知的发信渠道测试: %s", sw.Type), ""
	}
	tester := factory(msgObj)
	return tester.Test()
}

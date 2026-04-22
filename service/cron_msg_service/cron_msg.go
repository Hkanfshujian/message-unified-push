package cron_msg_service

import (
	"github.com/robfig/cron/v3"
	"ops-message-unified-push/models"
	"time"
)

type CronMsgResult struct {
	models.CronMessages

	NextTime     string   `json:"next_time"`
	TemplateName string   `json:"template_name"`
	ChannelNames []string `json:"channel_names"`
}

type CronMsgService struct {
	ID string

	Name       string
	TemplateID string
	Cron       string
	Title      string
	Url        string

	CreatedBy  string
	ModifiedBy string
	CreatedOn  string

	PageNum  int
	PageSize int
}

func (st *CronMsgService) Add() (string, error) {
	return models.AddSendCronMsg(st.Name, st.TemplateID, st.Cron, st.Name, "", st.CreatedBy)
}

func (st *CronMsgService) Edit(data map[string]interface{}) error {
	return models.EditCronMsg(st.ID, data)
}

func (st *CronMsgService) GetByID() (models.CronMessages, error) {
	return models.GetCronMsgByID(st.ID)
}

func (st *CronMsgService) Count() (int64, error) {
	return models.GetCronMessagesTotal(st.Name, st.getMaps())
}

func (st *CronMsgService) GetAll() ([]CronMsgResult, error) {
	msgs, err := models.GetCronMessages(st.PageNum, st.PageSize, st.Name, st.getMaps())
	if err != nil {
		return nil, err
	}
	templateIDs := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if msg.TemplateID != "" {
			templateIDs = append(templateIDs, msg.TemplateID)
		}
	}

	templateNameMap := make(map[string]string)
	if len(templateIDs) > 0 {
		templates, _ := models.GetTemplatesByIDs(templateIDs)
		for _, t := range templates {
			templateNameMap[t.ID] = t.Name
		}
	}

	insNameMap := make(map[string][]string)
	if len(templateIDs) > 0 {
		insList, _ := models.GetTemplateInsByTemplateIDs(templateIDs)
		for _, ins := range insList {
			insNameMap[ins.TemplateID] = append(insNameMap[ins.TemplateID], ins.WayName)
		}
	}

	return st.FillNextExecTime(msgs, templateNameMap, insNameMap), nil
}

func (st *CronMsgService) FillNextExecTime(msgs []models.CronMessages, templateNameMap map[string]string, insNameMap map[string][]string) []CronMsgResult {
	var result []CronMsgResult
	for _, msg := range msgs {
		channelNames := insNameMap[msg.TemplateID]
		r := CronMsgResult{
			CronMessages: msg,
			NextTime:     GetCronNextTime(msg.Cron),
			TemplateName: templateNameMap[msg.TemplateID],
			ChannelNames: channelNames,
		}
		result = append(result, r)
	}
	return result
}

func (st *CronMsgService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}

func (st *CronMsgService) Delete() error {
	return models.DeleteCronMsg(st.ID)
}

// GetCronNextTime 获取下次的执行时间
func GetCronNextTime(cronExpr string) string {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	schedule, err := specParser.Parse(cronExpr)
	if err != nil {
		return ""
	}
	nextTime := schedule.Next(time.Now()).Format("2006-01-02 15:04:05")
	return nextTime
}

// SendNow 立即发送定时消息（根据定时消息ID）
func (st *CronMsgService) SendNow(callerIP string) error {
	// 获取定时消息详情
	msg, err := st.GetByID()
	if err != nil {
		return err
	}

	// 调用发送服务
	return SendCronMessage(msg, callerIP)
}

// SendNowByParams 立即发送定时消息（根据传入的参数）
func (st *CronMsgService) SendNowByParams(callerIP string) error {
	if st.ID != "" {
		return st.SendNow(callerIP)
	}
	// 直接使用传入的参数构造消息对象
	msg := models.CronMessages{
		TemplateID: st.TemplateID,
		Name:       st.Name,
		Title:      st.Name,
	}

	// 调用发送服务
	return SendCronMessage(msg, callerIP)
}

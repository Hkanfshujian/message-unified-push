package channels

import (
	"fmt"
	"ops-message-unified-push/models"
	"ops-message-unified-push/pkg/message"
	"ops-message-unified-push/service/send_ins_service"
	"ops-message-unified-push/service/send_way_service"
)

type PushMeChannel struct{ *BaseChannel }

func NewPushMeChannel() *PushMeChannel {
	return &PushMeChannel{BaseChannel: NewBaseChannel(MessageTypePushMe, []string{FormatTypeText})}
}

func (c *PushMeChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailPushMe)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsPushMeConfig)
	if !ok {
		return "PushMe config校验失败", ""
	}

	cli := message.PushMe{
		PushKey: auth.PushKey,
		URL:     auth.URL,
		Date:    auth.Date,
		Type:    auth.Type,
	}

	res, err := cli.Request(content.Title, content.Text)
	if err != nil {
		return string(res), fmt.Sprintf("发送失败：%s", err.Error())
	}
	return string(res), ""
}

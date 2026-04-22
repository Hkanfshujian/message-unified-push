package message

import (
	"fmt"
	"sync"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
	"github.com/sirupsen/logrus"
)

type QyWeiXinApp struct {
	CorpID     string
	CorpSecret string
	AgentID    string
	ToUser     string // 默认发送给谁，多个接收者用‘|’分隔
}

var qywxAppCaches sync.Map

func getQywxAppCache(corpID, corpSecret string) cache.Cache {
	// 注意：wechat SDK 的 token cache key 默认只按 corpID 区分。
	// 同一企业下多个应用若使用不同 corpSecret，会发生 token 串用。
	// 这里按 corpID+corpSecret 隔离 cache，避免 301002（token 与 agent 不匹配）。
	cacheKey := corpID + "|" + corpSecret
	if v, ok := qywxAppCaches.Load(cacheKey); ok {
		if c, ok := v.(cache.Cache); ok {
			return c
		}
	}
	c := cache.NewMemory()
	qywxAppCaches.Store(cacheKey, c)
	return c
}

// SendText 发送文本消息
func (cw *QyWeiXinApp) SendText(content string, at ...string) (string, error) {
	client := cw.getClient()

	toUser := cw.ToUser
	if toUser == "" {
		toUser = "@all"
	}

	req := message.SendTextRequest{
		SendRequestCommon: &message.SendRequestCommon{
			ToUser:  toUser,
			MsgType: "text",
			AgentID: cw.AgentID,
		},
		Text: message.TextField{
			Content: content,
		},
	}

	res, err := client.SendText(req)
	if err != nil {
		logrus.Errorf("企业微信应用发送文本消息失败:%s", err)
		return "", err
	}
	if res.ErrCode != 0 {
		errMsg := fmt.Sprintf("企业微信应用发送文本消息失败, errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
		logrus.Error(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	return res.MsgID, nil
}

// SendMarkdown 发送Markdown消息
func (cw *QyWeiXinApp) SendMarkdown(content string, at ...string) (string, error) {
	client := cw.getClient()

	toUser := cw.ToUser
	if toUser == "" {
		toUser = "@all"
	}

	req := map[string]interface{}{
		"touser":  toUser,
		"msgtype": "markdown",
		"agentid": cw.AgentID,
		"markdown": map[string]interface{}{
			"content": content,
		},
	}

	res, err := client.Send("MessageSendMarkdown", req)
	if err != nil {
		logrus.Errorf("企业微信应用发送Markdown消息失败:%s", err)
		return "", err
	}
	if res.ErrCode != 0 {
		errMsg := fmt.Sprintf("企业微信应用发送Markdown消息失败, errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
		logrus.Error(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	return res.MsgID, nil
}

func (cw *QyWeiXinApp) getClient() *message.Client {
	wc := wechat.NewWechat()
	cfg := &config.Config{
		CorpID:     cw.CorpID,
		CorpSecret: cw.CorpSecret,
		Cache:      getQywxAppCache(cw.CorpID, cw.CorpSecret),
	}
	work := wc.GetWork(cfg)
	return work.GetMessage()
}

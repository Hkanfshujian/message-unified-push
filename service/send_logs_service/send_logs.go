package send_logs_service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"ops-message-unified-push/models"
	"net/url"
)

type SendTaskLogsService struct {
	ID        int
	TaskId    string
	Name      string
	Query     string
	StartTime string
	EndTime   string

	PageNum  int
	PageSize int
}

func (st *SendTaskLogsService) Count() (int64, error) {
	maps := st.getMaps()
	st.addTimeRangeToMaps(maps)
	return models.GetSendLogsTotal(st.Name, st.TaskId, maps)
}

func (st *SendTaskLogsService) GetAll() ([]models.LogsResult, error) {
	maps := st.getMaps()
	st.addTimeRangeToMaps(maps)
	tasks, err := models.GetSendLogs(st.PageNum, st.PageSize, st.Name, st.TaskId, maps)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *SendTaskLogsService) addTimeRangeToMaps(maps map[string]interface{}) {
	if st.StartTime != "" {
		maps["start_time"] = st.StartTime
	}
	if st.EndTime != "" {
		maps["end_time"] = st.EndTime
	}
}

func (st *SendTaskLogsService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	if len(st.Query) > 0 {
		decodedString, err := url.QueryUnescape(st.Query)
		if err != nil {
			logrus.Errorf("queryUrl编码解码失败: %s", err)
			return maps
		}
		err = json.Unmarshal([]byte(decodedString), &maps)
		if err != nil {
			logrus.Errorf("queryJson反序列化失败: %s", err)
			return maps
		}
	}
	return maps
}

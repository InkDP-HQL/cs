package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

type Log struct {
	ID        string `json:"id"`
	Timestamp string `json:"@timestamp"`
	Container struct {
		ID string `json:"id"`
	} `json:"container"`
	Log struct {
		Offset int64 `json:"offset"`
		File   struct {
			Path string `json:"path"`
		} `json:"file"`
	} `json:"log"`
	Message string `json:"message"`
	Input   struct {
		Type string `json"type"`
	} `json:"input"`
	Agent struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Version  string `json:"version"`
		Hostname string `json:"hostname"`
	} `json:"agent"`
}

func (*Log) Index() string {
	return "filebeat-7.11.1-2021.03.08-000001"
}

// LogList 日志列表
func LogList(param map[string]interface{}) (list []Log, err error) {
	log.Info(fmt.Sprintf("列表参数: %+v", param))
	var highlight *elastic.Highlight
	query := elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "objectId":
			query.Filter(elastic.NewTermQuery("objectId", val))
		case "userId":
			query.Filter(elastic.NewTermQuery("userId", val))
		case "startTime":
			query.Filter(elastic.NewRangeQuery("@timestamp").Gte(val).Lte(param["endTime"]))
		case "highLinght":
			highlight = elastic.NewHighlight().Field("message")
		case "message":
		}
	}

	sour, _ := query.Source()
	log.Info(fmt.Sprintf("ES查询body: %+v", sour))

	search := client.Search(new(Log).Index()).
		Query(query).
		From(0).Size(20).
		Pretty(true)
	if highlight != nil {
		search.Highlight(highlight)
	}
	searchResult, err := search.Do(context.Background())
	if err != nil {
		return
	}

	list = make([]Log, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		// log.Info(fmt.Sprintf("对象信息: %+v", string(val.Source)))
		_tag := Log{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return
}

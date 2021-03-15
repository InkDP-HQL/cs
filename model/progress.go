package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

// 分发状态
const (
	DistributeStatusStatrt     = "start"
	DistributeStatusFailed     = "failed"
	DistributeStatusSuccessful = "successful"
)

// DistributeProgress 分发进度
type DistributeProgress struct {
	ID         string `json:"id,omitempty"`
	OrderID    string `json:"orderId,omitempty"`
	UserID     string `json:"userId,omitempty"`   // 用户id
	ObjectID   string `json:"objectId,omitempty"` // 对象id
	IP         string `json:"ip,omitempty"`       //
	Type       string `json:"type,omitempty"`     // controller | collector
	Md5        string `json:"md5,omitempty"`
	Verison    string `json:"version,omitempty"`
	Name       string `json:"name,omitempty"`
	All        int    `json:"all,omitempty"`
	Progress   int    `json:"progress,omitempty"`
	Status     string `json:"status,omitempty"`     // success | failed  | distributing
	CreateTime string `json:"createTime,omitempty"` // 创建时间
	Message    string `json:"message"`              // 状态信息
}

// Index 索引
func (dp *DistributeProgress) Index() string {
	return "distribute_progress"
}

// Update 修改
func (dp *DistributeProgress) Update() error {
	id := dp.ID
	dp.ID = ""
	res, err := client.Update().
		Index(dp.Index()).
		Id(id).
		Doc(dp).
		Do(context.Background())
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("%+v", res))
	return nil
}

// DistributeProgressList 实例列表
func DistributeProgressList(param map[string]interface{}) (list []DistributeProgress, err error) {
	log.Info(fmt.Sprintf("列表参数: %+v", param))
	query := elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "id":
			getResult, err := client.Get().
				Index(new(DistributeProgress).Index()).
				Id(val.(string)).
				Pretty(true).
				Do(context.Background())
			if err != nil {
				return list, err
			}
			list = make([]DistributeProgress, 0, 1)
			log.Info(fmt.Sprintf("对象信息: %+v", getResult.Source))
			_tag := DistributeProgress{}
			json.Unmarshal([]byte(getResult.Source), &_tag)
			_tag.ID = getResult.Id
			list = append(list, _tag)
			return list, err
		case "orderId":
			query.Filter(elastic.NewTermQuery("orderId", val))
		case "objectId":
			query.Filter(elastic.NewTermQuery("objectId", val))
		case "userId":
			query.Filter(elastic.NewTermQuery("userId", val))
		case "type":
			query.Filter(elastic.NewTermQuery("type", val))
		}
	}

	searchResult, err := client.Search(new(DistributeProgress).Index()).
		Query(query).
		From(0).Size(100).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	list = make([]DistributeProgress, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		// log.Info(fmt.Sprintf("对象信息: %+v", string(val.Source)))
		_tag := DistributeProgress{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return
}

// AddDistributeProgress 添加分发进度
func AddDistributeProgress(mod *DistributeProgress) (id string, err error) {
	log.Info(fmt.Sprintf("对象信息: %+v", mod))
	res, err := client.Index().
		Index(mod.Index()).
		BodyJson(mod).
		Refresh("wait_for").
		Do(context.Background())

	return res.Id, err
}

// OneDisributeProgress 获取一条分发记录
func OneDisributeProgress(id string) (mod *DistributeProgress, err error) {
	getResult, err := client.Get().
		Index(new(DistributeProgress).Index()).
		Id(id).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	mod = new(DistributeProgress)
	log.Info(fmt.Sprintf("对象信息: %+v", getResult.Source))
	json.Unmarshal(getResult.Source, mod)
	mod.ID = getResult.Id
	return
}

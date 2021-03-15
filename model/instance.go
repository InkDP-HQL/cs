package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

// 实例状态
const (
	InstanceStatusUnregister = "unregister"
	InstanceStatusRegister   = "register"
	InstanceStatusRuning     = "runing"
	InstanceStatusStoped     = "stoped"
	InstanceStatusError      = "error"
)

// 实例类型
const (
	InstanceTypeAgent = "agent"
	InstanceTypeBeat  = "beat"
)

// Instance 实例
type Instance struct {
	ID       string `json:"id,omitempty"`
	UserID   string `json:"userId,omitempty"`   // 用户id
	ObjectID string `json:"objectId,omitempty"` // 对象id
	IP       string `json:"ip,omitempty"`       //
	Type     string `json:"type,omitempty"`     // controller | collector
	Verison  string `json:"version,omitempty"`
	MD5      string `json:"md5,omitempty"`
	Name     string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"` // register | unregister  | runing | error
	Message  string `json:"message"`          // 状态信息
}

// Index 索引
func (*Instance) Index() string {
	return "object_instance"
}

// Update 修改
func (ins *Instance) Update() error {
	id := ins.ID
	ins.ID = ""
	res, err := client.Update().
		Index(ins.Index()).
		Id(id).
		Doc(ins).
		Do(context.Background())
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("%+v", res))
	return nil
}

// InstanceList 实例列表
func InstanceList(param map[string]interface{}) (list []Instance, err error) {
	log.Info(fmt.Sprintf("列表参数: %+v", param))
	query := elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "name":
			query.Filter(elastic.NewPrefixQuery("name", val.(string)))
		case "objectId":
			query.Filter(elastic.NewTermQuery("objectId", val))
		case "userId":
			query.Filter(elastic.NewTermQuery("userId", val))
		case "type":
			query.Filter(elastic.NewTermQuery("type", val))
		}
	}

	searchResult, err := client.Search(new(Instance).Index()).
		Query(query).
		From(0).Size(100).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	list = make([]Instance, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		// log.Info(fmt.Sprintf("对象信息: %+v", string(val.Source)))
		_tag := Instance{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return
}

// InstanceAdd 添加实例
func InstanceAdd(mod *Instance) (err error) {
	log.Info(fmt.Sprintf("对象信息: %+v", mod))
	_, err = client.Index().
		Index(mod.Index()).
		BodyJson(mod).
		Refresh("wait_for").
		Do(context.Background())
	return err
}

// OneInstance 获取实例信息
func OneInstance(id string) (ins *Instance, err error) {
	getResult, err := client.Get().
		Index(new(Instance).Index()).Id(id).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	ins = new(Instance)
	ins.ID = getResult.Id
	err = json.Unmarshal([]byte(getResult.Source), ins)
	if err != nil {
		return
	}

	return
}

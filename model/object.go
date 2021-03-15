package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

// Object 监控对象
type Object struct {
	ID     string `json:"id,omitempty"`
	UserID string `json:"userId,omitempty"`
	IP     string `json:"ip,omitempty"`
	Name   string `json:"name,omitempty"`

	SSHAccount  string `json:"sshAccount,omitempty"`
	SSHPassword string `json:"sshPassword,omitempty"`
	SSHPort     int    `json:"sshPort,omitempty"`

	Category string `json:"category,omitempty"`
	Level    string `json:"level,omitempty"`
	Group    string `json:"group,omitempty"`
	Register bool   `json:"register"`
	Status   string `json:"status,omitempty"`
	Message  string `json:"message,omitempty"`
}

// Index 索引
func (*Object) Index() string {
	return "object"
}

// Update 修改
func (o *Object) Update() error {
	id := o.ID
	o.ID = ""
	res, err := client.Update().
		Index(o.Index()).
		Id(id).
		Doc(o).
		Do(context.Background())
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("%+v", res))
	return nil
}

// ObjectList 对象列表
func ObjectList(param map[string]interface{}) (list []Object, err error) {
	log.Info(fmt.Sprintf("列表参数: %+v", param))
	query := elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "tag":
			_val := val.([]string)
			tags := []interface{}{}
			for _, val := range _val {
				tags = append(tags, val)
			}
			query.Should(elastic.NewTermsQuery("category", tags...))
			query.Should(elastic.NewTermsQuery("level", tags...))
			query.Should(elastic.NewTermsQuery("group", tags...))
			query.MinimumNumberShouldMatch(1)

		case "userId":
			query.Filter(elastic.NewTermQuery("userId", val))
		case "register":
			query.Filter(elastic.NewTermQuery("register", val))
		}
	}

	sour, _ := query.Source()
	esQuey, _ := json.Marshal(sour)
	log.Info(fmt.Sprintf("ES请求参数: %s", string(esQuey)))
	searchResult, err := client.Search(new(Object).Index()).
		Query(query).
		From(0).Size(100).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	list = make([]Object, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		// log.Info(fmt.Sprintf("对象信息: %+v", string(val.Source)))
		_tag := Object{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return

}

// ObjectAdd 添加对象
func ObjectAdd(mod *Object) (err error) {
	log.Info(fmt.Sprintf("对象信息: %+v", mod))
	_, err = client.Index().
		Index(mod.Index()).
		BodyJson(mod).
		Refresh("wait_for").
		Do(context.Background())
	return err
}

// OneObject 获取对象信息
func OneObject(id string) (ins *Object, err error) {
	getResult, err := client.Get().Index(new(Object).Index()).Id(id).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	ins = new(Object)
	err = json.Unmarshal([]byte(getResult.Source), ins)
	if err != nil {
		return
	}
	return
}

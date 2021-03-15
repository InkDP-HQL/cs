package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

// Conf 配置信息
type Conf struct {
	ID       string `json:"id,omitempty"`
	ObjectID string `json:"objectId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	Pacel    struct {
		URL      string `json:"url,omitempty"`
		Account  string `json:"account,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"parcel,omitempty"`
	SSH struct {
		Port     int    `json:"port,omitempty"`
		Account  string `json:"account,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"ssh,omitempty"`
	Filebeat struct {
		Inputs []string `json:"inputs,omitempty"`
		Output string   `json:"output,omitempty"`
	} `json:"filebeat,omitempty"`

	Agent struct {
		Port int `json:"port,omitempty"`
	} `json:"agent,omitempty"`
}

// Index 索引
func (*Conf) Index() string {
	return "conf"
}

// IsExsit 配置是否存在
func (conf *Conf) IsExsit() (b bool, err error) {
	log.Info(fmt.Sprintf("标签信息: %+v", conf))
	query := elastic.NewBoolQuery()
	query.Filter(elastic.NewTermQuery("userId", conf.UserID))

	searchResult, err := client.Search(conf.Index()).
		Query(query).            // specify the query
		From(0).Size(1).         // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("标签信息: %+v", searchResult.Hits.Hits))

	return len(searchResult.Hits.Hits) != 0, nil
}

// OneConf 获取配置信息
func OneConf(userid string) (conf *Conf, err error) {
	query := elastic.NewBoolQuery().Filter(elastic.NewTermQuery("userId", userid))
	searchResult, err := client.Search(new(Conf).Index()).
		Query(query).
		From(0).Size(1).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	if len(searchResult.Hits.Hits) == 0 {
		return
	}
	val := searchResult.Hits.Hits[0]
	conf = new(Conf)
	json.Unmarshal([]byte(val.Source), conf)
	conf.ID = val.Id
	return
}

// UpdateConf 更新配置
func UpdateConf(mod *Conf) (err error) {
	log.Info(fmt.Sprintf("对象信息: %+v", mod))
	exsit, err := mod.IsExsit()
	if err != nil {
		return
	}
	if exsit {
		res, err := client.Update().
			Index(mod.Index()).
			Id(mod.ID).
			Doc(mod).
			Do(context.Background())
		if err != nil {
			log.Error(err.Error())
			return err
		}

		log.Info(fmt.Sprintf("修改结果: %+v", res))

		return err
	} else {
		_, err = client.Index().
			Index(mod.Index()).
			BodyJson(mod).
			Refresh("wait_for").
			Do(context.Background())
		return
	}
}

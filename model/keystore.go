package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
)

// Keystore 用户名密码键值对
type Keystore struct {
	ID       string `json:"id,omitempty"`
	UserID   string `json:"userId,omitempty"`
	Name     string `json:"name,omitempty"`
	Account  string `json:"account,omitempty"`
	Password string `json:"password,omitempty"`
}

// Index 索引
func (*Keystore) Index() string {
	return "keystore"
}

// KeystoreList 键值对列表
func KeystoreList(param map[string]interface{}) (list []Keystore, err error) {
	fmt.Printf("%+v\n", param)
	var query = elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "userId":
			query.Filter(elastic.NewTermQuery("userId", val))
		}
	}

	querySource := elastic.NewFetchSourceContext(true)
	querySource.Include("name", "userId", "account", "password")
	searchResult, err := client.Search(new(Keystore).Index()).
		FetchSourceContext(querySource).Type("_doc").
		Query(query).
		From(0).Size(100).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return
	}

	list = make([]Keystore, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		_tag := Keystore{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return

}

// KeystoreAdd 添加键值对
func KeystoreAdd(mod *Keystore) (err error) {
	_, err = client.Index().
		Index(mod.Index()).
		BodyJson(mod).
		Refresh("wait_for").
		Do(context.Background())
	return err
}

// DeleteKeystore 删除键值对
func DeleteKeystore(id string) (err error) {
	_, err = client.Delete().
		Index(new(Keystore).Index()).
		Id(id).
		Do(context.Background())
	return err
}

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"ylsz/hit/log"

	"github.com/olivere/elastic/v7"
)

// 标签类型
const (
	TagTypeCategory = "category"
	TagTypeLevel    = "level"
	TagTypeGroup    = "group"
)

// Tag 标签
type Tag struct {
	ID      string `json:"id,omitempty"`
	UserID  string `json:"userId"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Popular bool   `json:"popular"`
}

// Index 索引
func (*Tag) Index() string {
	return "tag"
}

// IsExsit 标签是否存在
func (t *Tag) IsExsit() (b bool, err error) {
	log.Info(fmt.Sprintf("标签信息: %+v", t))
	query := elastic.NewBoolQuery()
	query.Filter(elastic.NewTermQuery("name", t.Name)).
		Filter(elastic.NewTermQuery("userId", t.UserID)).
		Filter(elastic.NewTermQuery("type", t.Type))

	searchResult, err := client.Search(t.Index()).
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

// TagList 标签列表
func TagList(param map[string]interface{}) (list []Tag, err error) {
	fmt.Printf("%+v\n", param)
	var query = elastic.NewBoolQuery()
	for key, val := range param {
		switch key {
		case "name":
			query.Filter(elastic.NewPrefixQuery("name", val.(string)))
		case "type":
			query.Filter(elastic.NewTermQuery("type", val))
		case "popular":
			query.Filter(elastic.NewTermQuery("popular", true))
		case "userId":
			que := elastic.NewBoolQuery()
			que.Should(elastic.NewTermsQuery("userId", val))
			que.Should(elastic.NewTermsQuery("userId", ""))
			que.MinimumNumberShouldMatch(1)
			query.Filter(que)
		}
	}

	querySource := elastic.NewFetchSourceContext(true)
	querySource.Include("name", "userId", "popular", "type")
	searchResult, err := client.Search(new(Tag).Index()).
		FetchSourceContext(querySource).Type("_doc").
		Query(query).            // specify the query
		From(0).Size(100).       // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		return
	}

	list = make([]Tag, 0, len(searchResult.Hits.Hits))
	for _, val := range searchResult.Hits.Hits {
		_tag := Tag{}
		json.Unmarshal([]byte(val.Source), &_tag)
		_tag.ID = val.Id
		list = append(list, _tag)
	}
	return

}

// TagAdd 添加标签
func TagAdd(mod *Tag) (err error) {
	_, err = client.Index().
		Index(mod.Index()).
		BodyJson(mod).
		Refresh("wait_for").
		Do(context.Background())
	return err
}

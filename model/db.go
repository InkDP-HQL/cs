package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"ylsz/hit/log"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/sony/sonyflake"
)

type Ormer interface {
	Type() string
	Index() string
	ID() string
	SetID() string
}

// ElasticAddr es服务地址
var ElasticAddr string

var flake = sonyflake.NewSonyflake(sonyflake.Settings{})

func uuid() string {
	var id uint64
	for {
		_id, err := flake.NextID()
		if err == nil {
			id = _id
			break
		}
	}

	return strconv.FormatUint(id, 32)
}

func Uuid() string {
	return uuid()
}
func newESConfig() elastic.Config {
	conf := elastic.Config{}
	conf.Addresses = []string{ElasticAddr}
	return conf
}

func newESClient() *elastic.Client {
	conf := elastic.Config{}
	conf.EnableDebugLogger = true
	conf.Addresses = []string{ElasticAddr}
	cli, _ := elastic.NewClient(conf)
	return cli
}

// DB 数据层接口，使用 ElasticSearch 做数据持久化
// 1.输出 ElasticSearch 日志
// 2.不支持并发
// 3.基于 ElasticSearch7
type DB interface {
	Create(Modeler) (string, error)
	Update(Modeler) error
	// 设置 index
	SetIndex(string) DB
	// 设置检索条件, interface{} 最多为一级指针对象
	Set(string, interface{}) DB

	// 分页
	From(offset int) DB
	Size(offset int) DB

	// 执行 Match 搜索, 并将结果映射到 interface{}
	// 1.未找到匹配值 interface{} 为零值
	// 2. interface{} 是一个指针对象, 值类型接受 array、slice、struct
	Match(interface{}) error
	// 类似 Match, 不过执行match_phrase_prefix 搜索
	MatchPhrasePrefix(interface{}) error
	MatchPhrase(interface{}) error
}

type Modeler interface {
	index() string
	id() string
}

func NewDB() DB {
	db := new(esdb)
	return db
}

type esdb struct {
	from   int
	size   int
	cond   []*jsonvalue.V
	term   []*jsonvalue.V
	terms  []*jsonvalue.V
	filter []*jsonvalue.V
	index  string
}

func (db *esdb) Create(v Modeler) (id string, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}

	var buff bytes.Buffer
	buff.Write(b)
	res, err := newESClient().Create(db.index, v.id(), &buff)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	jsonV, err := jsonvalue.Unmarshal(b)
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("向 ES 请求结果: %s", jsonV.MustMarshal()))

	if code, _ := jsonV.GetInt("status"); code != 200 && code != 0 {
		str, _ := jsonV.GetString("error", "reason")
		return "", errors.New(str)
	}

	id, _ = jsonV.GetString("_id")

	return
}

// TODO: 设置数组得不到想要的结果待修复
func (db *esdb) Set(key string, val interface{}) DB {
	if len(db.cond) == 0 {
		db.cond = make([]*jsonvalue.V, 0)
	}
	typ := reflect.TypeOf(val)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	cond := jsonvalue.NewObject()
	switch typ.Kind() {
	case reflect.String:
		cond.SetString(reflect.ValueOf(val).String()).At(key)
	case reflect.Bool:
		cond.SetBool(reflect.ValueOf(val).Bool()).At(key)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cond.SetInt64(reflect.ValueOf(val).Int()).At(key)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cond.SetUint64(reflect.ValueOf(val).Uint()).At(key)
	case reflect.Float32, reflect.Float64:
		cond.SetFloat64(reflect.ValueOf(val).Float(), -1).At(key)
	case reflect.Array, reflect.Slice:
		_val := reflect.ValueOf(val)
		for i := 0; i < _val.Len(); i++ {
			cond := jsonvalue.NewObject()
			switch typ.Elem().Kind() {
			case reflect.String:
				cond.SetString(_val.Index(i).String()).At(key)
			case reflect.Bool:
				cond.SetBool(_val.Index(i).Bool()).At(key)
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				cond.SetInt64(_val.Index(i).Int()).At(key)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				cond.SetUint64(_val.Index(i).Uint()).At(key)
			case reflect.Float32, reflect.Float64:
				cond.SetFloat64(_val.Index(i).Float(), -1).At(key)
			}
			db.filter = append(db.filter, cond)
		}

		return db
	case reflect.Map:
	case reflect.Ptr:
	case reflect.Struct:
		b, _ := json.Marshal(val)
		v, _ := jsonvalue.Unmarshal(b)
		cond.Set(v).At(key)
	}
	db.cond = append(db.cond, cond)
	return db
}

func (db *esdb) Match(v interface{}) (err error) {
	return db.match("match", v)
}

func (db *esdb) MatchPhrasePrefix(v interface{}) (err error) {
	return db.match("match_phrase_prefix", v)
}
func (db *esdb) MatchPhrase(v interface{}) (err error) {
	return db.match("match_phrase", v)
}
func (db *esdb) SetIndex(index string) DB {
	db.index = index
	return db
}

func (db *esdb) Update(v Modeler) (err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("json字符串: %s", b))
	var buff bytes.Buffer
	buff.WriteString(`{"doc":`)
	buff.Write(b)
	buff.WriteString(`}`)
	res, err := newESClient().Update(db.index, v.id(), &buff)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	jsonV, err := jsonvalue.Unmarshal(b)
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("向 ES 请求结果: %s", res.String()))

	if code, _ := jsonV.GetInt("status"); code != 200 && code != 0 {
		str, _ := jsonV.GetString("error", "reason")
		return errors.New(str)
	}
	return
}

// From 包含 offset 那条记录
func (db *esdb) From(offset int) DB {
	db.from = offset
	return db
}

func (db *esdb) Size(size int) DB {
	db.size = size
	return db
}

func (db *esdb) match(query string, v interface{}) (err error) {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return ErrNeedsPtr
	}
	typ = typ.Elem()

	jsonV := jsonvalue.NewObject()
	for i, cond := range db.cond {
		jsonV.Set(cond).At("query", "bool", "must", i, query)
	}

	for i, cond := range db.filter {
		jsonV.Set(cond).At("query", "bool", "filter", i, "term")
	}

	jsonV.SetInt(db.from).At("from")
	if db.size > 0 {
		jsonV.SetInt(db.size).At("size")
	}

	log.Info(fmt.Sprintf("向 ES 请求body参数: %s", jsonV.MustMarshal()))

	var buff bytes.Buffer
	buff.Write(jsonV.MustMarshal())
	// 向ES请求数据
	es := newESClient()
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(db.index),
		es.Search.WithBody(&buff),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	jsonRes, err := jsonvalue.Unmarshal(b)
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("向 ES 请求结果: %s", jsonRes.MustMarshal()))

	// 检验请求结果
	if code, _ := jsonRes.GetInt("status"); code != 200 && code != 0 {
		if errType, _ := jsonRes.GetString("error", "root_cause", 0, "type"); errType == "index_not_found_exception" {
			return ErrIndexNotFoundException
		}
		str, _ := jsonRes.GetString("error", "reason")
		return errors.New(str)
	}

	// 映射结果到到 interface{}

	total, err := jsonRes.GetInt("hits", "total", "value")
	if err != nil {
		log.Error(err.Error())
	}
	if total == 0 {
		return
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Slice:
		sli := reflect.MakeSlice(typ, 0, total)
		for i := 0; i < total; i++ {
			_val := reflect.New(typ.Elem())
			jsonVal, _ := jsonRes.Get("hits", "hits", i, "_source")
			if jsonVal == nil {
				continue
			}
			err = json.Unmarshal(jsonVal.MustMarshal(), _val.Interface())
			if err != nil {
				return
			}
			sli = reflect.Append(sli, _val.Elem())
		}
		reflect.ValueOf(v).Elem().Set(sli)
	case reflect.Struct:
		jsonVal, _ := jsonRes.Get("hits", "hits", 0, "_source")
		err = json.Unmarshal(jsonVal.MustMarshal(), v)
		if err != nil {
			return
		}
	}

	return
}

func (db *esdb) search(body []byte, v interface{}) (err error) {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return ErrNeedsPtr
	}
	typ = typ.Elem()

	var buff bytes.Buffer
	buff.Write(body)
	// 向ES请求数据
	es := newESClient()
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(db.index),
		es.Search.WithBody(&buff),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	jsonRes, err := jsonvalue.Unmarshal(b)
	if err != nil {
		return
	}

	log.Info(fmt.Sprintf("向 ES 请求结果: %s", jsonRes.MustMarshal()))

	// 检验请求结果
	if code, _ := jsonRes.GetInt("status"); code != 200 && code != 0 {
		if errType, _ := jsonRes.GetString("error", "root_cause", 0, "type"); errType == "index_not_found_exception" {
			return ErrIndexNotFoundException
		}
		str, _ := jsonRes.GetString("error", "reason")
		return errors.New(str)
	}

	// 映射结果到到 interface{}

	total, err := jsonRes.GetInt("hits", "total", "value")
	if err != nil {
		log.Error(err.Error())
	}
	if total == 0 {
		return
	}

	switch typ.Kind() {
	case reflect.Array, reflect.Slice:
		sli := reflect.MakeSlice(typ, 0, total)
		for i := 0; i < total; i++ {
			_val := reflect.New(typ.Elem())
			jsonVal, _ := jsonRes.Get("hits", "hits", i, "_source")
			if jsonVal == nil {
				continue
			}
			err = json.Unmarshal(jsonVal.MustMarshal(), _val.Interface())
			if err != nil {
				return
			}
			sli = reflect.Append(sli, _val.Elem())
		}
		reflect.ValueOf(v).Elem().Set(sli)
	case reflect.Struct:
		jsonVal, _ := jsonRes.Get("hits", "hits", 0, "_source")
		err = json.Unmarshal(jsonVal.MustMarshal(), v)
		if err != nil {
			return
		}
	}

	return
}

type ESSearch struct {
}

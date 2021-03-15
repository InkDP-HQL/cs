package model

import (
	"encoding/json"
	"fmt"
	"ylsz/hit/log"
)

// Parcel 存储库文件详情
type Parcel struct {
	ID         string `json:"id,omitempty"`
	Addr       string `json:"addr,omitempty"` //  存储库地址
	File       string `json:"file,omitempty"` // 分发文件名,全称
	Num        int    `json:"num,omitempty"`  // 分发次数
	Status     string `json:"status,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

type filter struct {
	Terms interface{} `json:"terms,omitempty"`
	Term  interface{} `json:"term,omitempty"`
}

type Search struct {
	Source []string `json:"_source,omitempty"`
	Query  struct {
		Bool struct {
			Filter []filter `json:"filter,omitempty"`
		} `json:"bool,omitempty"`
	} `json:"query,omitempty"`
	From int `json:"from,omitempty"`
	Size int `json:"size,omitempty"`
}

func NewParcel() *Parcel {
	return &Parcel{}
}

func (p *Parcel) index() string {
	return "member"
}

func (p *Parcel) id() string {
	return p.ID
}

func (p *Parcel) SetID(id string) {
	p.ID = id
}

func (p *Parcel) GetID() (id string) {
	return p.ID
}

func (p *Parcel) Increase(n int) error {
	_p := new(Parcel)
	err := NewDB().SetIndex(p.index()).Set("file", p.File).Match(_p)
	if err != nil && err != ErrIndexNotFoundException {
		return err
	}
	if p.id() != "" {
		p.Num = _p.Num + n
		err = NewDB().SetIndex(p.index()).Update(p)
		return err
	}

	p.ID = uuid()
	_, err = NewDB().SetIndex(p.index()).Create(p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Parcel) List(n int) (list []Parcel, err error) {
	list = make([]Parcel, 0)
	search := new(Search)
	search.Source = []string{
		"id",
		"file",
		"num",
		"status",
	}

	from := 0
	size := 100
	search.Size = size
	var li []Parcel
	for {
		search.From = from
		li = make([]Parcel, 0, size)
		b, err := json.Marshal(search)
		if err != nil {
			return nil, err
		}

		log.Info(fmt.Sprintf("请求数据: %s", b))

		db := new(esdb)
		db.SetIndex(p.index())
		err = db.search(b, &li)
		if err != nil {
			return nil, err
		}

		list = append(list, li...)
		if len(li) < size {
			break
		}
		from += size
	}

	return
}

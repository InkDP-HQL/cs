package model

import (
	"github.com/olivere/elastic/v7"
)

func ESCLient() (cli *elastic.Client, err error) {
	// Create a client
	cli, err = elastic.NewClient(elastic.SetURL(ESURL))
	if err != nil {
		// Handle error
		return
	}
	return
}

var client *elastic.Client
var ESURL string = "http://192.168.10.15:9200"

func init() {
	cli, err := elastic.NewClient(elastic.SetURL(ESURL))
	if err != nil {
		panic(err)
	}
	client = cli
}

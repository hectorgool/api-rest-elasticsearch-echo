package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hectorgool/api-rest-elasticsearch-gin/common"
	"github.com/satori/go.uuid"
	elastic "gopkg.in/olivere/elastic.v5"
	"log"
	"os"
)

type (
	Document struct {
		Id         string `json:"id"`
		Ciudad     string `json:"ciudad"`
		Colonia    string `json:"colonia"`
		Cp         string `json:"cp"`
		Delegacion string `json:"delegacion"`
		Location   `json:"location"`
	}

	Location struct {
		Lat float32 `json:"lat"`
		Lon float32 `json:"lon"`
	}
)

var (
	client *elastic.Client
	ctx    = context.Background()
)

const mapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                    "autocomplete": {
                        "tokenizer": "whitespace",
                        "filter": [
                            "lowercase",
                            "engram"
                        ]
                    }
                },
                "filter": {
                    "engram": {
                        "type": "edgeNGram",
                        "min_gram": 1,
                        "max_gram": 10
                    }
                }
            }
        }
    },
    "mappings": {
        "postal_code": {
            "properties": {
            	"id": {
                    "type": "text",
                    "store" : "yes"
                },
                "cp": {
                    "type": "text",
                    "store" : "yes"
                },
                "colonia": {
                    "type": "text",
                    "store": "yes",
                    "fielddata": true
                },                 
                "ciudad": {
                    "type": "text",
                    "store": "yes"
                },                                
                "delegacion": {
                    "type": "text",
                    "store": "yes"
                },    
                "location": {
                    "type": "geo_point"
                }
            }
        }
    }
}
`

func init() {

	var err error

	client, err = elastic.NewClient(
		elastic.SetURL(os.Getenv("ELASTICSEARCH_ENTRYPOINT")),
		elastic.SetBasicAuth(os.Getenv("ELASTICSEARCH_USERNAME"), os.Getenv("ELASTICSEARCH_PASSWORD")),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)),
	)
	common.CheckError(err)

}

func CreateIndex() interface{} {

	exists, err := client.IndexExists(os.Getenv("ELASTICSEARCH_INDEX")).Do(ctx)
	common.CheckError(err)

	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex(os.Getenv("ELASTICSEARCH_INDEX")).BodyString(mapping).Do(ctx)
		common.CheckError(err)
		if !createIndex.Acknowledged {
			// Not acknowledged
			log.Fatalln("Not CreateIndex.")
		}
	}

	return "Index created"

}

func Ping() (string, error) {

	info, code, err := client.Ping(os.Getenv("ELASTICSEARCH_ENTRYPOINT")).Do(ctx)
	common.CheckError(err)

	msg := fmt.Sprintf("Elasticsearch returned with code %d and version %s", code, info.Version.Number)
	return msg, nil

}

func TermToJson(term string) (string, error) {

	if len(term) == 0 {
		return "", errors.New("No string supplied\n")
	}
	searchJson := fmt.Sprintf(
		`{
	   "query": {
	     	"match": {
	        	"_all": {
	            	"operator": "and",
	            	"query": "%v"
	         	}
	      	}
	   },
	   "size": 10,
	   "sort": [
	      	{
	        	"colonia": {
	            	"order": "asc"
	         	}
	      	}
	   ]
	}`, term)

	return searchJson, nil

}

func SearchTerm(term string) (*elastic.SearchResult, error) {

	if len(term) == 0 {
		return nil, errors.New("No string supplied!!!\n")
	}

	//Convert string to json query for elasticsearch
	searchJson, err := TermToJson(term)
	common.CheckError(err)

	// Search with a term source
	searchResult, err := client.Search().
		Index(os.Getenv("ELASTICSEARCH_INDEX")).
		Type(os.Getenv("ELASTICSEARCH_TYPE")).
		Source(searchJson).
		Do(ctx)
	common.CheckError(err)

	return searchResult, nil

}

func DisplayResults(searchResult *elastic.SearchResult) ([]*Document, error) {

	var Documents []*Document

	for _, hit := range searchResult.Hits.Hits {
		d := &Document{}
		//parses *hit.Source into the instance of the Document struct
		err := json.Unmarshal(*hit.Source, &d)
		common.CheckError(err)
		//Puts d into a map for later access
		Documents = append(Documents, d)
	}
	return Documents, nil

}

func Search(term string) ([]*Document, error) {

	searchResult, err := SearchTerm(term)
	common.CheckError(err)

	result, err := DisplayResults(searchResult)
	common.CheckError(err)

	return result, nil

}

func DeleteDocument(id string) interface{} {

	res, err := client.Delete().
		Index(os.Getenv("ELASTICSEARCH_INDEX")).
		Type(os.Getenv("ELASTICSEARCH_TYPE")).
		Id(id).
		Do(ctx)
	common.CheckError(err)

	return res.Found

}

func ReadDocument(id string) interface{} {

	res, err := client.Get().
		Index(os.Getenv("ELASTICSEARCH_INDEX")).
		Type(os.Getenv("ELASTICSEARCH_TYPE")).
		Id(id).
		Do(ctx)
	common.CheckError(err)

	return res.Source
}

func CreateDocument(d Document) interface{} {

	id := uuid.NewV4().String()

	doc := Document{
		Id:         id,
		Ciudad:     d.Ciudad,
		Colonia:    d.Colonia,
		Cp:         d.Cp,
		Delegacion: d.Delegacion,
		Location: Location{
			Lat: d.Lat,
			Lon: d.Lon,
		},
	}

	res, err := client.Index().
		Index(os.Getenv("ELASTICSEARCH_INDEX")).
		Type(os.Getenv("ELASTICSEARCH_TYPE")).
		Id(id).
		BodyJson(doc).
		Do(ctx)
	common.CheckError(err)

	return res.Id

}

func UpdateDocument(d Document) interface{} {

	update, err := client.Update().Index(os.Getenv("ELASTICSEARCH_INDEX")).Type(os.Getenv("ELASTICSEARCH_TYPE")).Id(d.Id).
		Script(elastic.NewScriptInline("ctx._source.ciudad = d.Ciudad")).
		Upsert(map[string]interface{}{"ciudad": ""}).
		Do(ctx)
	common.CheckError(err)

	return update.Id

}

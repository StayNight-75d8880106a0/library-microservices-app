package elasticsearch

import (
	"book-management-services/internal/helper"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

var ElasticseearchClient *elasticsearch.Client

func InitElasticsearch(url string) error {

	config := elasticsearch.Config{
		Addresses: []string{
			url,
		},
	}

	esClient, errClient := elasticsearch.NewClient(config)

	if errClient != nil {
		log.Printf("Error creating Elasticsearch client: %s", errClient)
		return helper.NewInternalServerError("An Error During Connect To Elasticsearch!", helper.ErrorDetail{Detail: errClient.Error()})
	}

	response, errInfo := esClient.Info()

	if errInfo != nil {
		log.Printf("Error getting Elasticsearch info: %s", errInfo)
		return helper.NewInternalServerError("An Error During Connect To Elasticsearch!", helper.ErrorDetail{Detail: errInfo.Error()})
	}

	defer response.Body.Close()

	if response.IsError() {
		log.Printf("Elasticsearch returned an error response: %s", response.String())
		return helper.NewInternalServerError("An Error During Connect To Elasticsearch!", helper.ErrorDetail{Detail: response.String()})
	}

	ElasticseearchClient = esClient

	log.Println("Success Connect To Elasticsearch ✅")

	return nil

}

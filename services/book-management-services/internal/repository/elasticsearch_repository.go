package repository

import (
	"book-management-services/internal/dto"
	"book-management-services/internal/models"
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ElasticsearchRepositoryInterface interface {
	Search(ctx context.Context, param dto.GetBooksQuery) ([]models.Books, int64, error)
	IndexToElasticsearch(ctx context.Context, book *models.Books) error
	DeleteFromElasticsearch(ctx context.Context, ID string) error
	UpdateInElasticsearch(ctx context.Context, book *models.Books) error
}

type ElasticSearchRepository struct {
	elasticsearch *elasticsearch.Client
}

func NewElasticSearchRepository(es *elasticsearch.Client) *ElasticSearchRepository {
	return &ElasticSearchRepository{
		elasticsearch: es,
	}
}

func (repo *ElasticSearchRepository) Search(ctx context.Context, param dto.GetBooksQuery) ([]models.Books, int64, error) {

	from := (param.Page - 1) * param.Limit

	var mustQueries []map[string]interface{}

	mustQueries = append(mustQueries, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  param.Keywords,
			"fields": []string{"title^3", "authors^2", "publisher", "description", "isbn", "isbn.keyword"},
			"type":   "best_fields",
		},
	})

	if param.Category != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"category.keyword": param.Category,
			},
		})
	}

	esQuery := map[string]interface{}{
		"from":  from,
		"size":  param.Limit,
		"query": map[string]interface{}{"bool": map[string]interface{}{"must": mustQueries}},
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, 0, err
	}

	response, errResponse := repo.elasticsearch.Search(
		repo.elasticsearch.Search.WithContext(ctx),
		repo.elasticsearch.Search.WithIndex("books_index"),
		repo.elasticsearch.Search.WithBody(&buf),
	)

	if errResponse != nil {
		return nil, 0, errResponse
	}

	defer response.Body.Close()

	var result map[string]interface{}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	hitsMap, _ := result["hits"].(map[string]interface{})
	totalMap, _ := hitsMap["total"].(map[string]interface{})
	total := int64(totalMap["value"].(float64))

	hits := hitsMap["hits"].([]interface{})
	var books []models.Books

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		data, _ := json.Marshal(source)
		var book models.Books
		_ = json.Unmarshal(data, &book)
		books = append(books, book)
	}

	return books, total, nil

}

func (repo *ElasticSearchRepository) IndexToElasticsearch(ctx context.Context, book *models.Books) error {

	data, errData := json.Marshal(book)

	if errData != nil {
		return errData
	}

	request := esapi.IndexRequest{
		Index:      "books_index",
		DocumentID: book.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	response, errResponse := request.Do(ctx, repo.elasticsearch)

	if errResponse != nil {
		return errResponse
	}

	defer response.Body.Close()

	return nil

}

func (repo *ElasticSearchRepository) DeleteFromElasticsearch(ctx context.Context, ID string) error {

	request := esapi.DeleteRequest{
		Index:      "books_index",
		DocumentID: ID,
		Refresh:    "true",
	}

	response, errResponse := request.Do(ctx, repo.elasticsearch)

	if errResponse != nil {
		return errResponse
	}

	defer response.Body.Close()

	return nil

}

func (repo *ElasticSearchRepository) UpdateInElasticsearch(ctx context.Context, book *models.Books) error {

	data, errData := json.Marshal(book)
	if errData != nil {
		return errData
	}

	request := esapi.IndexRequest{
		Index:      "books_index",
		DocumentID: book.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	response, errResponse := request.Do(ctx, repo.elasticsearch)
	if errResponse != nil {
		return errResponse
	}
	defer response.Body.Close()

	return nil

}

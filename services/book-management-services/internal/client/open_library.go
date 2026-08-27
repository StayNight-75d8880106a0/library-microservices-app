package client

import (
	"book-management-services/internal/config"
	"book-management-services/internal/dto"
	"book-management-services/internal/helper"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OpenLibraryClientInterface interface {
	FetchByISBN(ctx context.Context, isbn string) (*dto.OpenLibraryBook, error)
}

type OpenLibraryClient struct {
	httpClient *http.Client
	cfg        config.OpenLibraryAPIConfig
}

func NewOpenLibraryClient(cfg config.OpenLibraryAPIConfig) *OpenLibraryClient {
	return &OpenLibraryClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		cfg:        cfg,
	}
}

func (c *OpenLibraryClient) FetchByISBN(ctx context.Context, isbn string) (*dto.OpenLibraryBook, error) {

	url := fmt.Sprintf("%s?bibkeys=ISBN:%s&jscmd=data&format=json", c.cfg.OpenLibraryAPIURL, isbn)

	request, errRequest := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if errRequest != nil {
		return nil, helper.NewInternalServerError("An Error During Request To Open Library!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := c.httpClient.Do(request)

	if errResponse != nil || response.StatusCode != http.StatusOK {
		return nil, nil
	}

	defer response.Body.Close()

	var rawResponse map[string]interface{}

	errDecode := json.NewDecoder(response.Body).Decode(&rawResponse)

	if errDecode != nil {
		return nil, helper.NewInternalServerError("An Error During Decoding Response From Open Library!", helper.ErrorDetail{Detail: errDecode.Error()})
	}

	key := fmt.Sprintf("ISBN:%s", isbn)

	data, exists := rawResponse[key].(map[string]interface{})

	if !exists {
		return nil, nil
	}

	result := &dto.OpenLibraryBook{}

	if title, ok := data["title"].(string); ok {
		result.Title = title
	}

	if publishers, ok := data["publishers"].([]interface{}); ok && len(publishers) > 0 {
		if p, ok := publishers[0].(map[string]interface{}); ok {
			result.Publisher, _ = p["name"].(string)
		}
	}

	if pages, ok := data["number_of_pages"].(float64); ok {
		result.Page = int(pages)
	}

	if pubDate, ok := data["publish_date"].(string); ok {
		result.PublishedDate = pubDate
	}

	if cover, ok := data["cover"].(map[string]interface{}); ok {
		result.CoverURL, _ = cover["medium"].(string)
	}

	if authors, ok := data["authors"].([]interface{}); ok && len(authors) > 0 {
		for _, author := range authors {
			a, ok := author.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := a["name"].(string); ok && name != "" {
				result.Authors = append(result.Authors, name)
			}
		}
	}

	if subjects, ok := data["subjects"].([]interface{}); ok {
		for _, subject := range subjects {
			s, ok := subject.(map[string]interface{})
			if !ok {
				continue
			}
			if name, ok := s["name"].(string); ok && name != "" {
				result.Subjects = append(result.Subjects, name)
			}
		}
	}

	return result, nil

}

package dto

import "time"

type OpenLibraryBook struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publish_date"`
	Page          int      `json:"number_of_pages"`
	CoverURL      string   `json:"cover_url"`
	Subjects      []string `json:"subjects"`
}

type BookResponse struct {
	ID             *string `json:"id"`
	ISBN           *string `json:"isbn"`
	Title          *string `json:"title"`
	Authors        *string `json:"authors"`
	Publisher      *string `json:"publisher"`
	PublishedDate  *string `json:"publishedDate"`
	Page           *int    `json:"page"`
	Description    *string `json:"description"`
	CoverURL       *string `json:"coverURL"`
	Category       *string `json:"category"`
	TotalStock     *int    `json:"totalStock"`
	AvailableStock *int    `json:"availableStock"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type CreateBookRequest struct {
	ISBN           *string `json:"isbn" binding:"required"`
	Title          string  `json:"title"`
	Authors        string  `json:"authors"`
	Publisher      string  `json:"publisher"`
	PublishedDate  string  `json:"publishedDate"`
	Page           int     `json:"page"`
	Description    string  `json:"description"`
	CoverURL       string  `json:"coverURL"`
	Category       string  `json:"category"`
	TotalStock     *int    `json:"totalStock" binding:"required"`
	AvailableStock *int    `json:"availableStock" binding:"required"`
}

type UpdateBookRequest struct {
	ISBN           *string `json:"isbn" binding:"required"`
	Title          *string `json:"title" binding:"required"`
	Authors        *string `json:"authors" binding:"required"`
	Publisher      *string `json:"publisher" binding:"required"`
	PublishedDate  *string `json:"publishedDate" binding:"required"`
	Page           *int    `json:"page" binding:"required"`
	Description    *string `json:"description" binding:"required"`
	CoverURL       *string `json:"coverURL" binding:"required"`
	Category       *string `json:"category" binding:"required"`
	TotalStock     *int    `json:"totalStock" binding:"required"`
	AvailableStock *int    `json:"availableStock" binding:"required"`
}

type GetBooksQuery struct {
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
	Keywords string `form:"keywords"`
	Category string `form:"category"`
}

type BorrowingCreatedEvent struct {
	BookID    string    `json:"bookID"`
	Quantity  int       `json:"quantity"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

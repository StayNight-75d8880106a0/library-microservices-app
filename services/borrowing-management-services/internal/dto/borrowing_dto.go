package dto

import "time"

type CreateBorrowingRequest struct {
	BookID  *string    `json:"bookID" binding:"required"`
	Duedate *time.Time `json:"dueDate"`
}

type BorrowingResponse struct {
	ID         string `json:"ID"`
	UserID     string `json:"userID"`
	BookID     string `json:"bookID"`
	BorrowCode string `json:"borrowCode"`
	BorrowedAt string `json:"borrowedAt"`
	DueDate    string `json:"dueDate"`
	ReturnedAt string `json:"returnedAt"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type BorrowingUpdateRequest struct {
	Status *string `json:"status" binding:"required"`
}

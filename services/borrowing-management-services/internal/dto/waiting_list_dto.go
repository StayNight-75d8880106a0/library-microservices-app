package dto

type CreateWaitingListRequest struct {
	BookID *string `json:"bookID" binding:"required"`
}

type WaitingListResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userID"`
	BookID         string  `json:"bookID"`
	ArrivalRate    float64 `json:"arrivalRate"`
	ServiceRate    float64 `json:"serviceRate"`
	NumServers     int     `json:"numServers"`
	Utilization    float64 `json:"utilization"`
	ProbWait       float64 `json:"probWait"`
	AvgQueueLen    float64 `json:"avgQueueLen"`
	AvgWaitMin     float64 `json:"avgWaitMin"`
	OptimalServers *int    `json:"optimalServers"`
	QueueNumber    int     `json:"queueNumber"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

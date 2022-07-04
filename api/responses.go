package api

type UserStatsResponse struct {
	UserName           string  `json:"userName"`
	SuccessfulRequests int     `json:"successfulRequests"`
	FailedRequests     int     `json:"failedRequests"`
	TotalRequests      int     `json:"totalRequests"`
	TotalTimeElapsed   float64 `json:"totalTimeElapsed"`
	AverageRequestTime float64 `json:"averageRequestTime"`
	ThrottledRequests  int     `json:"throttledRequests"`
}

type ElapsedTimeResponse struct {
	ElapsedTime float64 `json:"elapsedTime"`
}

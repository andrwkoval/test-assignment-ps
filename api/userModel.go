package api

type UserModel struct {
	SuccessfulRequests    int     `json:"successfulRequests"`
	FailedRequests        int     `json:"failedRequests"`
	SuccessfulTimeElapsed float64 `json:"successfulTimeElapsed"`
	FailedTimeElapsed     float64 `json:"failedTimeElapsed"`
	ThrottledRequests     int     `json:"throttledRequests"`
}

func (user UserModel) getTotalRequests() int {
	return user.SuccessfulRequests + user.FailedRequests
}

func (user UserModel) getTotalTime() float64 {
	return user.SuccessfulTimeElapsed + user.FailedTimeElapsed
}

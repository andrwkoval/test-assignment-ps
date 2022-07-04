package api

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultLimit = 10

	// RequestStatusFailed statuses for requests to update user stats
	RequestStatusFailed     = 0
	RequestStatusSuccessful = 1
	RequestStatusThrottled  = 2
)

var SuccessCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "success_requests_count",
		Help: "Number of succeeded requests per endpoint",
	},
	[]string{"endpoint"},
)

var FailedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "failed_requests_count",
		Help: "Number of failed requests per endpoint",
	},
	[]string{"endpoint"},
)

type UpdatingCounter struct {
	Counter      int
	NeedToUpdate bool
}

type API struct {
	UserStats    map[string]*UserModel
	UserCounters map[string]*UpdatingCounter
	mtx          sync.Mutex
}

func (api *API) GetAddress(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	validUrl, err := url.ParseRequestURI(params.ByName("address")[1:])
	name := params.ByName("name")

	if err != nil {
		api.errorHandler(w, r, http.StatusNotFound)
		api.updateStats(RequestStatusFailed, 0, name, "/user/:name/url/*address")
		return
	}

	limitExceeded := api.userLimitExceeded(name)
	if limitExceeded {
		api.errorHandler(w, r, http.StatusForbidden)
		return
	}

	start := time.Now()

	resp, err := http.Get(validUrl.String())

	if err != nil {
		api.errorHandler(w, r, http.StatusNotFound)
		elapsed := time.Since(start).Seconds()
		api.updateStats(RequestStatusFailed, elapsed, name, "/user/:name/url/*address")
		return
	}

	_, err = io.ReadAll(resp.Body)

	if err != nil {
		api.errorHandler(w, r, http.StatusInternalServerError)
		elapsed := time.Since(start).Seconds()
		api.updateStats(RequestStatusFailed, elapsed, name, "/user/:name/url/*address")
		return
	}

	elapsed := time.Since(start).Seconds()

	defer resp.Body.Close()

	elapsedTime := ElapsedTimeResponse{elapsed}
	elapsedJson, err := json.Marshal(elapsedTime)

	if err != nil {
		api.errorHandler(w, r, http.StatusInternalServerError)
		return
	}

	api.updateStats(RequestStatusSuccessful, elapsed, name, "/user/:name/url/*address")

	if counter, ok := api.UserCounters[name]; ok {
		counter.Counter += 1
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(elapsedJson)
}

func (api *API) ShowUserStats(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	if entry, ok := api.UserStats[params.ByName("name")]; ok {
		response := UserStatsResponse{
			UserName:           params.ByName("name"),
			SuccessfulRequests: entry.SuccessfulRequests,
			FailedRequests:     entry.FailedRequests,
			TotalRequests:      entry.getTotalRequests(),
			TotalTimeElapsed:   entry.getTotalTime(),
			AverageRequestTime: entry.getTotalTime() / float64(entry.getTotalRequests()),
			ThrottledRequests:  entry.ThrottledRequests,
		}

		statsJson, err := json.Marshal(response)

		if err != nil {
			api.errorHandler(w, r, http.StatusNotFound)
			FailedCounter.WithLabelValues("/user/stats").Inc()
			return
		}

		SuccessCounter.WithLabelValues("/user/stats").Inc()

		w.Header().Set("Content-Type", "application/json")
		w.Write(statsJson)
	} else {
		api.errorHandler(w, r, http.StatusNotFound)
		FailedCounter.WithLabelValues("/user/stats").Inc()
		return
	}
}

func (api *API) ShowAllStats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	response := UserStatsResponse{}

	for _, element := range api.UserStats {
		response.SuccessfulRequests += element.SuccessfulRequests
		response.FailedRequests += element.FailedRequests
		response.TotalRequests += element.getTotalRequests()
		response.TotalTimeElapsed += element.getTotalTime()
		response.ThrottledRequests += element.ThrottledRequests
	}

	response.AverageRequestTime = response.TotalTimeElapsed / float64(response.TotalRequests)

	statsJson, err := json.Marshal(response)

	if err != nil {
		api.errorHandler(w, r, http.StatusNotFound)
		FailedCounter.WithLabelValues("/stats").Inc()
		return
	}

	SuccessCounter.WithLabelValues("/stats").Inc()

	w.Header().Set("Content-Type", "application/json")
	w.Write(statsJson)
}

func (api *API) errorHandler(w http.ResponseWriter, _ *http.Request, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusForbidden:
		fmt.Fprintf(w, "Forbidden 403.\nDescription: Exceeded request limit per minute.")
	case http.StatusNotFound:
		fmt.Fprintf(w, "Not found 404")
	default:
		fmt.Fprintf(w, "Server error 500")
	}
}

func (api *API) userLimitExceeded(name string) bool {

	var err error = nil
	limit := os.Getenv("REQUESTS_PER_MINUTE_LIMIT")

	var intLimit int

	if limit == "" {
		intLimit = DefaultLimit
	} else {
		intLimit, err = strconv.Atoi(limit)

		if err != nil {
			intLimit = DefaultLimit
		}
	}

	if counter, ok := api.UserCounters[name]; ok {
		if counter.Counter >= intLimit {
			api.updateStats(RequestStatusThrottled, 0, name, "")
			return true
		}
	} else {
		api.UserCounters[name] = &UpdatingCounter{
			Counter:      0,
			NeedToUpdate: true,
		}
	}

	if counter, ok := api.UserCounters[name]; ok && counter.NeedToUpdate {
		counter.NeedToUpdate = false

		go func() {
			<-time.After(1 * time.Minute)
			api.mtx.Lock()
			defer api.mtx.Unlock()

			counter.Counter = 0
			counter.NeedToUpdate = true
		}()
	}

	return false
}

func (api *API) getOrCreate(name string) *UserModel {
	if entry, ok := api.UserStats[name]; ok {
		return entry
	} else {
		userModel := &UserModel{}
		api.UserStats[name] = userModel
		return userModel
	}
}

func (api *API) updateStats(status int, elapsed float64, name string, endpoint string) {
	userModel := api.getOrCreate(name)

	switch status {
	case RequestStatusFailed:
		userModel.FailedRequests += 1
		userModel.FailedTimeElapsed += elapsed
		FailedCounter.WithLabelValues(endpoint).Inc()
	case RequestStatusSuccessful:
		userModel.SuccessfulRequests += 1
		userModel.SuccessfulTimeElapsed += elapsed
		SuccessCounter.WithLabelValues(endpoint).Inc()
	case RequestStatusThrottled:
		userModel.ThrottledRequests += 1
	default:
		return
	}
}

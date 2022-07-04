package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"ps-assignment/api"
)

func main() {

	prometheus.MustRegister(api.SuccessCounter)
	prometheus.MustRegister(api.FailedCounter)

	apiServer := api.API{}
	apiServer.UserStats = make(map[string]*api.UserModel)
	apiServer.UserCounters = make(map[string]*api.UpdatingCounter)

	router := httprouter.New()
	router.GET("/user/:name/url/*address", apiServer.GetAddress)
	router.GET("/user/:name/stats", apiServer.ShowUserStats)
	router.GET("/stats", apiServer.ShowAllStats)
	router.Handler(http.MethodGet, "/metrics", promhttp.Handler())

	fmt.Println("Server started at: localhost:10000. Try it!")
	log.Fatal(http.ListenAndServe(":10000", router))
}

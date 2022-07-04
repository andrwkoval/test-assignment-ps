# PerfectScale test assignment
## Assignment

1. Learn go

    Write a webserver in Go:

    a. Implement the following Apis:
    * /USER/:USERNAME/URL/*ADDRESS
        * your webserver will download the WEB_ADDRESS http response and respond with number of milliseconds the request took.
        * you should store per use statistics:
            * how many requests for this user? 
            * how many requests succeeded/failed?
            * total requests time (per success/fail)
    * /USER/:USERNAME/STATS
        * return statistics - number of successful and failed requests, avg req time. 
    * /STATS
        * return statistics like above, ignore USER_NAME dimension (total stats of all users)			

    b. add a configurable rate limiter - X requests per minute is allowed
    * please add to user stats also number of throttled (=not allowed) requests

2. Learn k8s
    * k8s tutorial
    * Deploy any k8s, you can use minikube on your laptop (or k3s, or kind..)
    * helm tutorial
    * create helm for your web server
    * deploy your helm chart on minikube. Note that for high availability, we want at least two running instances of the web server (rate limiter is distinct for every instance)
    * port-forward to your service running on k8s and make sure you can access service APIs (should you port-forward to pod? service?)

3. Learn Prometheus & Grafana ( we can help a bit here). 

    a. deploy prometheus and grafana to your cluster ( https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

    b. Create a grafana dashboard
    * Show how much memory / cpu does your service use. 

    c. expose number of success / failed requests per endpoint
    * for this you should instrument your webserver with prometheus metrics
    * add to your helm chart a servicemonitor resource (so prometheus can scrape your custom metrics)
    * add graphs to the grafana dashboard above

___
## Throttler solution X limits per minute

`map[userID]counter`

1. if below limit - fire request, otherwise skip and add to throttled
```
if counter[userID] >= limit { 
	throttled += 1
} else { 
	fire request
	counter[userID] += 1
}
```

2. counter = 0 after 1 minute

```
go func() {
	<-timer.After(1*time.Minute)
	mutex.lock
	defer { mutex.unlock }
	counter[userID] = 0
}()

```

___


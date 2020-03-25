package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hitsPerCountry = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "getmyip_hits",
	Help: "Number of hits in the geolocation service",
},
	[]string{"country"},
)

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/oschwald/geoip2-golang"
)

func floatToString(num float64) string {
	return strconv.FormatFloat(num, 'f', 6, 64)
}

func getRemoteIP(req *http.Request) string {
	forward := req.Header.Get("X-Forwarded-For")
	ipstr := ""
	if forward != "" {
		ipstr = forward
	} else {
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Fatal(err)
		}
		netIP := net.ParseIP(ip)
		ipstr = netIP.String()
	}
	return ipstr
}

type geodb struct {
	db *geoip2.Reader
}

func (g *geodb) getRecordForIP(ipstr string) *geoip2.City {
	ip := net.ParseIP(ipstr)
	record, err := g.db.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	return record
}

type jsonHandler struct {
	geoipdb *geodb
}

func (jh *jsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ipstr := getRemoteIP(req)
	record := jh.geoipdb.getRecordForIP(ipstr)
	data := map[string]string{
		"ip":   ipstr,
		"cc":   record.Country.IsoCode,
		"city": record.City.Names["en"],
		"lat":  floatToString(record.Location.Latitude),
		"lon":  floatToString(record.Location.Longitude),
	}
	dataJSON, _ := json.Marshal(data)
	fmt.Fprintf(w, string(dataJSON))
}

type txtHandler struct {
	geoipdb *geodb
}

func (th *txtHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ipstr := getRemoteIP(req)
	record := th.geoipdb.getRecordForIP(ipstr)

	fmt.Fprintf(w, "Your IP: %s\n", ipstr)
	fmt.Fprintf(w, "Your Country: %s\n", record.Country.IsoCode)
	fmt.Fprintf(w, "Your City: %s\n", record.City.Names["en"])
	fmt.Fprintf(w, "Your Coordinates: %s, %s\n",
		floatToString(record.Location.Latitude),
		floatToString(record.Location.Longitude))
}

func main() {
	var port = flag.Int("port", 9001, "port where the service listens on")
	var dbpath = flag.String("geodb", "/var/lib/GeoIP/GeoLite2-City.mmdb", "path to the GeoLite2-City database")
	flag.Parse()

	db, err := geoip2.Open(*dbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	geoipdb := geodb{db}

	mux := http.NewServeMux()
	jh := &jsonHandler{&geoipdb}
	mux.Handle("/json", jh)

	th := &txtHandler{&geoipdb}
	mux.Handle("/", th)

	log.Println("Started Geolocation Service")
	log.Printf("Listening on port %v...\n", *port)
	http.ListenAndServe(":"+strconv.Itoa(*port), mux)
}

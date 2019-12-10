// Copyright (c) 2018 LEAP Encryption Access Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/hongshibao/go-kdtree"
	"github.com/oschwald/geoip2-golang"
	"github.com/tidwall/cities"
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
	db          *geoip2.Reader
	Forbidden   []string
	Gateways    []gateway
	GatewayTree *kdtree.KDTree
	GatewayMap  map[[3]float64][]gateway
	earth       *ellipsoid.Ellipsoid
}

func (g *geodb) getPointForLocation(lat float64, lon float64) *EuclideanPoint {
	x, y, z := g.earth.ToECEF(lat, lon, 0)
	p := NewEuclideanPoint(x, y, z)
	return p
}

func randomizeGateways(gws []gateway) []gateway {
	dest := make([]gateway, len(gws))
	perm := rand.Perm(len(gws))
	for i, v := range perm {
		dest[v] = gws[i]
	}
	return dest
}

func (g *geodb) sortGateways(lat float64, lon float64) []string {
	ret := make([]string, 0)
	t := g.getPointForLocation(lat, lon)
	nn := g.GatewayTree.KNN(t, len(g.Gateways))
	for i := 0; i < len(nn); i++ {
		p := [3]float64{nn[i].GetValue(0), nn[i].GetValue(1), nn[i].GetValue(2)}
		cityGateways := g.GatewayMap[p]
		if len(cityGateways) > 1 {
			cityGateways = randomizeGateways(cityGateways)
		}
		for _, gw := range cityGateways {
			if !stringInSlice(gw.Host, g.Forbidden) {
				if !stringInSlice(gw.Host, ret) {
					ret = append(ret, gw.Host)
				}
			}
		}
	}
	return ret
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (g *geodb) geolocateGateways(b *bonafide) {
	g.GatewayMap = make(map[[3]float64][]gateway)
	gatewayPoints := make([]kdtree.Point, 0)

	for i := 0; i < len(b.eip.Gateways); i++ {
		gw := b.eip.Gateways[i]
		coord := geolocateCity(gw.Location)
		gw.Coordinates = coord
		b.eip.Gateways[i] = gw

		p := g.getPointForLocation(coord.Latitude, coord.Longitude)

		gatewayPoints = append(gatewayPoints, *p)
		var i [3]float64
		copy(i[:], p.Vec)
		g.GatewayMap[i] = append(g.GatewayMap[i], gw)
	}
	g.Gateways = b.eip.Gateways
	g.GatewayTree = kdtree.NewKDTree(gatewayPoints)
}

func (g *geodb) getRecordForIP(ipstr string) *geoip2.City {
	ip := net.ParseIP(ipstr)
	record, err := g.db.City(ip)
	if err != nil {
		log.Fatal(err)
	}
	return record
}

func geolocateCity(city string) coordinates {
	// because some cities apparently are not good enough for the top 10k
	missingCities := make(map[string]coordinates)
	missingCities["hongkong"] = coordinates{22.319201099, 114.1696121}

	re := regexp.MustCompile("-| ")
	for _, c := range cities.Cities {
		canonical := strings.ToLower(city)
		canonical = re.ReplaceAllString(canonical, "")
		if strings.ToLower(c.City) == canonical {
			return coordinates{c.Latitude, c.Longitude}
		}
		v, ok := missingCities[canonical]
		if ok == true {
			return v
		}

	}
	return coordinates{0, 0}
}

type jsonHandler struct {
	geoipdb *geodb
}

type GeolocationJSON struct {
	Ip        string   `json:"ip"`
	Cc        string   `json:"cc"`
	City      string   `json:"city"`
	Latitude  float64  `json:"lat"`
	Longitude float64  `json:"lon"`
	Gateways  []string `json:"gateways"`
}

func (jh *jsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ipstr := getRemoteIP(req)
	record := jh.geoipdb.getRecordForIP(ipstr)
	sortedGateways := jh.geoipdb.sortGateways(record.Location.Latitude, record.Location.Longitude)

	data := &GeolocationJSON{
		ipstr,
		record.Country.IsoCode,
		record.City.Names["en"],
		record.Location.Latitude,
		record.Location.Longitude,
		sortedGateways,
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
	rand.Seed(time.Now().UnixNano())
	var port = flag.Int("port", 9001, "port where the service listens on")
	var dbpath = flag.String("geodb", "/var/lib/GeoIP/GeoLite2-City.mmdb", "path to the GeoLite2-City database")
	var notls = flag.Bool("notls", false, "disable TLS on the service")
	var key = flag.String("server_key", "", "path to the key file for TLS")
	var crt = flag.String("server_crt", "", "path to the cert file for TLS")
	var forbidstr = flag.String("forbid", "", "comma-separated list of forbidden gateways")
	flag.Parse()

	forbidden := strings.Split(*forbidstr, ",")
	fmt.Println("Forbidden gateways:", forbidden)

	if *notls == false {
		if *key == "" || *crt == "" {
			log.Fatal("you must provide -server_key and -server_crt parameters")
		}
		if _, err := os.Stat(*crt); os.IsNotExist(err) {
			log.Fatal("path for crt file does not exist!")
		}
		if _, err := os.Stat(*key); os.IsNotExist(err) {
			log.Fatal("path for key file does not exist!")
		}
	}

	db, err := geoip2.Open(*dbpath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	earth := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
	geoipdb := geodb{db, forbidden, nil, nil, nil, &earth}

	log.Println("Seeding gateway list...")
	bonafide := newBonafide()
	bonafide.getGateways()

	geoipdb.geolocateGateways(bonafide)
	bonafide.listGateways()

	mux := http.NewServeMux()
	jh := &jsonHandler{&geoipdb}
	mux.Handle("/json", jh)

	th := &txtHandler{&geoipdb}
	mux.Handle("/", th)

	log.Println("Started Geolocation Service")
	log.Printf("Listening on port %v...\n", *port)

	pstr := ":" + strconv.Itoa(*port)
	if *notls == true {
		err = http.ListenAndServe(pstr, mux)
	} else {
		err = http.ListenAndServeTLS(pstr, *crt, *key, mux)
	}

	if err != nil {
		log.Fatal("error in listenAndServe[TLS]: ", err)
	}
}

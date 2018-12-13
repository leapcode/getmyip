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
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
	Gateways    []gateway
	GatewayTree *kdtree.KDTree
	GatewayMap  map[[3]float64]gateway
	earth       *ellipsoid.Ellipsoid
}

func (g *geodb) getPointForLocation(lat float64, lon float64) *EuclideanPoint {
	x, y, z := g.earth.ToECEF(lat, lon, 0)
	p := NewEuclideanPoint(x, y, z)
	return p
}

func (g *geodb) getClosestGateway(lat float64, lon float64) gateway {
	t := g.getPointForLocation(lat, lon)
	nn := g.GatewayTree.KNN(t, 1)[0]
	p := [3]float64{nn.GetValue(0), nn.GetValue(1), nn.GetValue(2)}
	closestGateway := g.GatewayMap[p]
	return closestGateway
}

func (g *geodb) geolocateGateways(b *bonafide) {
	g.GatewayMap = make(map[[3]float64]gateway)
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
		g.GatewayMap[i] = gw
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
	for i := 0; i < len(cities.Cities); i++ {
		c := cities.Cities[i]
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

func (jh *jsonHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ipstr := getRemoteIP(req)
	record := jh.geoipdb.getRecordForIP(ipstr)
	closestGateway := jh.geoipdb.getClosestGateway(record.Location.Latitude, record.Location.Longitude)

	data := map[string]string{
		"ip":   ipstr,
		"cc":   record.Country.IsoCode,
		"city": record.City.Names["en"],
		"lat":  floatToString(record.Location.Latitude),
		"lon":  floatToString(record.Location.Longitude),
		"gw":   closestGateway.Location,
		"gwip": closestGateway.IPAddress,
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

	earth := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
	geoipdb := geodb{db, nil, nil, nil, &earth}

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
	http.ListenAndServe(":"+strconv.Itoa(*port), mux)
}

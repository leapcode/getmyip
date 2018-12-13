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
	"fmt"
	"net/http"
)

const (
	// yes, I am cheating. The config file is also exposed on the top-level
	// domain, which is served behind a letsencrypt certificate. this saves passing
	// the certificate for the ca etc.
	eipAPI = "https://black.riseup.net/1/config/eip-service.json"
)

type bonafide struct {
	client *http.Client
	eip    *eipService
}

type eipService struct {
	Gateways  []gateway
	Locations map[string]struct {
		CountryCode string
		Hemisphere  string
		Name        string
		Timezone    string
	}
}

type gateway struct {
	Host        string
	Location    string
	IPAddress   string `json:"ip_address"`
	Coordinates coordinates
}

type coordinates struct {
	Latitude  float64
	Longitude float64
}

func newBonafide() *bonafide {
	client := &http.Client{}
	return &bonafide{client, nil}
}

func (b *bonafide) getGateways() ([]gateway, error) {
	if b.eip == nil {
		err := b.fetchEipJSON()
		if err != nil {
			return nil, err
		}
	}
	return b.eip.Gateways, nil
}

func (b *bonafide) fetchEipJSON() error {
	resp, err := b.client.Post(eipAPI, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("get eip json has failed with status: %s", resp.Status)
	}
	var eip eipService
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&eip)
	if err != nil {
		return err
	}
	b.eip = &eip
	return nil
}

func (b *bonafide) listGateways() error {
	if b.eip == nil {
		return fmt.Errorf("cannot list gateways, it is empty")
	}

	for i := 0; i < len(b.eip.Gateways); i++ {
		fmt.Printf("\t%v\n", b.eip.Gateways[i])
	}
	return nil

}

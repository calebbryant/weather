package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var zipMap map[zipCode]latLong

type Properties struct {
	StationIdentifier string `json:"stationIdentifier"`
	Name              string `json:"name"`
}

type Feature struct {
	Id         string     `json:"id"`
	Properties Properties `json:"properties"`
}

type Station struct {
	Features []Feature `json:"features"`
}

func (s *Station) String() string {
	return s.Features[0].Properties.StationIdentifier
}

type zipCode string
type latLong [2]string

func zipToLatLong(z zipCode) (latLong, error) {
	if l, ok := zipMap[z]; ok {
		return l, nil
	}
	return latLong{}, fmt.Errorf("zip code not found")
}

func readZips() map[zipCode]latLong {
	var zipMap = make(map[zipCode]latLong)
	ziptext, err := ioutil.ReadFile("zip-data.csv")
	if err != nil {
		fmt.Println("no file")
	}
	r := csv.NewReader(strings.NewReader(string(ziptext)))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("%s", err)
			return nil
		}
		if record[0] == "ZIP" {
			continue
		}
		trimRecord := func(record string) string {
			record = strings.Trim(record, " ")
			s := strings.SplitN(record, ".", 2)
			for len(s[1]) > 2 {
				s[1] = s[1][0 : len(s[1])-1]
			}
			return fmt.Sprintf("%s.%s", s[0], s[1])
		}
		zipMap[zipCode(record[0])] = latLong([2]string{
			trimRecord(record[1]),
			trimRecord(record[2])})
	}
	return zipMap
}

func stationFromZip(z zipCode) (output *Station, err error) {
	l, err := zipToLatLong(z)
	if err != nil {
		return
	}
	var resp *http.Response

	client := &http.Client{}
	req, _ := http.NewRequest("GET",
		fmt.Sprintf(
			"https://api.weather.gov/points/%s,%s/stations",
			l[0], l[1]),
		nil)
	req.Proto = "HTTP/1.1"
	req.Header.Set("Accept", "*/*")
	if resp, err = client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	output = new(Station)
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(output); err != nil {
		return
	}
	return
}

func main() {
	zipMap = readZips()
	w, err := stationFromZip(zipCode("78704"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", w.String())
}

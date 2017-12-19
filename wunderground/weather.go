package wunderground

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var validQuery = regexp.MustCompile("^[A-Z a-z]+$")
var APIKey = os.Getenv("WUNDERGROUND_API_KEY")

type IntOrNANString struct {
	value string
}

func (i *IntOrNANString) String() string {
	return i.value
}

func (i *IntOrNANString) UnmarshalJSON(data []byte) (err error) {
	var d interface{}

	if err = json.Unmarshal(data, &d); err != nil {
		return err
	}
	i.value = fmt.Sprintf("%s", d)
	return nil
}

type ResponseFeatures struct {
	Conditions int `json:"conditions"`
}

type ResponseMetadata struct {
	Version          string            `json:"version"`
	TermsofService   string            `json:"termsofService"`
	ResponseFeatures *ResponseFeatures `json:"features"`
}

type CurrentObservation struct {
	Image                 *Image               `json:"image"`
	DisplayLocation       *DisplayLocation     `json:"display_location"`
	ObservationLocation   *ObservationLocation `json:"observation_location"`
	StationId             string               `json:"station_id"`
	ObservationTime       string               `json:"observation_time"`
	ObservationTimeRfc822 string               `json:"observation_time_rfc822"`
	ObservationEpoch      string               `json:"observation_epoch"`
	LocalTimeRfc822       string               `json:"local_time_rfc822"`
	LocalEpoch            string               `json:"local_epoch"`
	LocalTzShort          string               `json:"local_tz_short"`
	LocalTzLong           string               `json:"local_tz_long"`
	LocalTzOffset         string               `json:"local_tz_offset"`
	Weather               string               `json:"weather"`
	TemperatureString     string               `json:"temperature_string"`
	TempF                 float32              `json:"temp_f"`
	TempC                 float32              `json:"temp_c"`
	RelativeHumidity      string               `json:"relative_humidity"`
	WindString            string               `json:"wind_string"`
	WindDir               string               `json:"wind_dir"`
	WindDegrees           float32              `json:"wind_degrees"`
	WindMph               IntOrNANString       `json:"wind_mph"`
	WindGustMph           IntOrNANString       `json:"wind_gust_mph"`
	WindKph               IntOrNANString       `json:"wind_kph"`
	WindGustKph           IntOrNANString       `json:"wind_gust_kph"`
	PressureMb            string               `json:"pressure_mb"`
	PressureIn            string               `json:"pressure_in"`
	PressureTrend         string               `json:"pressure_trend"`
	DewpointString        string               `json:"dewpoint_string"`
	DewpointF             IntOrNANString       `json:"dewpoint_f"`
	DewpointC             IntOrNANString       `json:"dewpoint_c"`
	HeatIndexString       IntOrNANString       `json:"heat_index_string"`
	HeatIndexF            IntOrNANString       `json:"heat_index_f"`
	HeatIndexC            IntOrNANString       `json:"heat_index_c"`
	WindchillString       string               `json:"windchill_string"`
	WindchillF            IntOrNANString       `json:"windchill_f"`
	WindchillC            IntOrNANString       `json:"windchill_c"`
	FeelslikeString       string               `json:"feelslike_string"`
	FeelslikeF            IntOrNANString       `json:"feelslike_f"`
	FeelslikeC            IntOrNANString       `json:"feelslike_c"`
	VisibilityMi          string               `json:"visibility_mi"`
	VisibilityKm          string               `json:"visibility_km"`
	Solarradiation        string               `json:"solarradiation"`
	UV                    string               `json:"UV"`
	Precip1hrIn           string               `json:"precip_1hr_in"`
	Precip1hrMetric       string               `json:"precip_1hr_metric"`
	Precip1hrString       string               `json:"precip_1hr_string"`
	PrecipTodayString     string               `json:"precip_today_string"`
	PrecipTodayIn         string               `json:"precip_today_in"`
	PrecipTodayMetric     string               `json:"precip_today_metric"`
	Icon                  string               `json:"icon"`
	IconUrl               string               `json:"icon_url"`
	ForecastUrl           string               `json:"forecast_url"`
	HistoryUrl            string               `json:"history_url"`
	ObUrl                 string               `json:"ob_url"`
	Nowcast               string               `json:"nowcast"`
}

type Image struct {
	Url   string `json:"url"`
	Title string `json:"title"`
	Link  string `json:"link"`
}

type DisplayLocation struct {
	Full           string `json:"full"`
	City           string `json:"city"`
	State          string `json:"state"`
	StateName      string `json:"state_name"`
	Country        string `json:"country"`
	CountryIso3166 string `json:"country_iso3166"`
	Zip            string `json:"zip"`
	Magic          string `json:"magic"`
	Wmo            string `json:"wmo"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Elevation      string `json:"elevation"`
}

type ObservationLocation struct {
	Full           string `json:"full"`
	City           string `json:"city"`
	State          string `json:"state"`
	Country        string `json:"country"`
	CountryIso3166 string `json:"country_iso3166"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Elevation      string `json:"elevation"`
}

func (w *Weather) String() string {
	t := template.New("weather")
	t, _ = t.Parse(`Current Weather For {{.DisplayLocation.Full}}
Observatory: {{.ObservationLocation.Full}}
{{.CurrentObservation.ObservationTime}}
Conditions: {{.Weather}}
Temperature: {{.TemperatureString}}
Relative humidity: {{.RelativeHumidity}}
Heat index: {{.HeatIndexString}}
Wind speed: {{.WindString}}
Wind chill: {{.WindchillString}}
Precipitation in the last hour: {{.Precip1hrIn}} m
Dewpoint: {{.DewpointString}}
`)

	buf := new(bytes.Buffer)
	t.Execute(buf, w)

	return buf.String()
}

type Weather struct {
	ResponseMetadata   `json:"response"`
	CurrentObservation `json:"current_observation"`
}

func GetWeather(query string) (result *Weather, err error) {
	location := strings.SplitN(query, ",", 2)
	if len(location) == 2 {
		location[0] = strings.TrimSpace(location[0])
		location[1] = strings.TrimSpace(location[1])
	}

	if len(location) != 2 ||
		len(location[1]) != 2 ||
		len(location[0]) > 32 || !validQuery.MatchString(location[0]) || !validQuery.MatchString(location[1]) {
		return nil, fmt.Errorf("Query \"%s\" should be in form City, State e.g. Austin, TX", query)
	}

	var resp *http.Response

	if resp, err = http.Get(
		fmt.Sprintf(
			"https://api.wunderground.com/api/%s/conditions/q/%s/%s.json",
			APIKey,
			location[1],
			location[0])); err != nil {
		return nil, err
	}

	output := new(Weather)

	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(output); err != nil {
		return nil, err
	}

	return output, nil
}

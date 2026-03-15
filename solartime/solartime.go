package solartime

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pachecot/angle"
	"github.com/pachecot/julian"
	"github.com/pachecot/solar"
)

// SolarTime holds the computed solar event times and sun position data
// for a given location and date.
type SolarTime struct {

	// the reference day in UTC
	Day time.Time

	// the reference date/time in UTC
	DateTime time.Time

	// sunrise as a UTC time value
	SunriseTime time.Time

	// sunset as a UTC time value
	SunsetTime time.Time

	// UTC timezone offset string
	Offset string

	// date formatted as YYYY-MM-DD
	Date string

	// time of solar noon in kitchen format
	SolarNoon string

	// time of sunrise in kitchen format
	Sunrise string

	// time of sunset in kitchen format
	Sunset string

	// total duration of sunlight
	SunlightDuration string

	// corrected sun elevation angle in degrees
	SolarElevation float64

	// sun azimuth angle in degrees
	SolarAzimuth float64
}

type Request struct {
	Latitude  float64
	Longitude float64
	Date      *time.Time
	Offset    *int
}

var missingQuery = errors.New("missing query parameter")
var badQuery = errors.New("bad query parameter")

func parseQuery(r *http.Request) (Request, error) {
	var request Request
	var err error

	q := r.URL.Query()

	if _, ok := q["lat"]; !ok {
		return request, missingQuery
	}
	if _, ok := q["long"]; !ok {
		return request, missingQuery
	}

	sLat := q.Get("lat")
	sLng := q.Get("long")
	sDate := q.Get("date")
	sOffset := q.Get("offset")

	request.Latitude, err = strconv.ParseFloat(sLat, 64)
	if err != nil {
		return request, badQuery
	}
	request.Longitude, err = strconv.ParseFloat(sLng, 64)
	if err != nil {
		return request, badQuery
	}
	if sDate != "" {
		// request.Date, err = time.Parse("2006-01-02T15:04:05Z", sDate)
		*request.Date, err = time.Parse("2006-01-02", sDate)
		if err != nil {
			fmt.Println("error - ", err)
			*request.Date = time.Time{}
		}
	}
	if sOffset != "" {
		offset, err := strconv.ParseInt(sOffset, 0, 64)
		if err != nil {
			return request, badQuery
		}
		*request.Offset = int(offset)
	}

	return request, nil
}

func kitchenTime(day time.Time, h time.Duration) string {
	return day.Add(h).Format(time.Kitchen)
}

func calcSolarTime(request Request) SolarTime {

	day, offset := parseDay(request)
	date := time.Now()
	if request.Date != nil {
		date = *request.Date
	}

	solLoc := solar.Location{
		Longitude:      angle.Degrees(request.Longitude),
		Latitude:       angle.Degrees(request.Latitude),
		TimeZoneOffset: offset,
	}
	JD := julian.Time(day)
	JT := julian.Time(date)
	sunrise, sunset := solLoc.SunTimes(JD)
	elevation := solLoc.SolarElevationCorrected(JT)
	azimuth := solLoc.SolarAzimuth(JT)

	TransitTime := sunset - sunrise
	SolarNoon := sunrise + TransitTime/2
	return SolarTime{
		Day:              day.UTC(),
		DateTime:         date.UTC(),
		Offset:           offset.String(),
		Date:             day.Format("2006-01-02"),
		SolarNoon:        kitchenTime(day, SolarNoon),
		Sunrise:          kitchenTime(day, sunrise),
		Sunset:           kitchenTime(day, sunset),
		SunriseTime:      day.Add(sunrise).UTC(),
		SunsetTime:       day.Add(sunset).UTC(),
		SunlightDuration: TransitTime.String(),
		SolarElevation:   elevation.Degrees(),
		SolarAzimuth:     azimuth.Degrees(),
	}
}

func parseDay(request Request) (time.Time, time.Duration) {
	when := time.Now()

	day := truncateDay(when)
	offset := tzOffset(when)

	if request.Date != nil {
		day = *request.Date
	}

	if request.Offset != nil {
		offset = time.Duration(*request.Offset) * time.Hour
	}
	return day, offset
}

// tzOffset returns the time zone offset at the specified time.
func tzOffset(when time.Time) time.Duration {
	_, offset_sec := when.Zone()
	return time.Duration(offset_sec) * time.Second
}

// truncateDay returns the time truncated to midnight of the same day.
func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func getSolar(r *http.Request) (SolarTime, error) {
	request, err := parseQuery(r)
	if err != nil {
		return SolarTime{}, err
	}
	return calcSolarTime(request), nil
}

func XmlHandler(w http.ResponseWriter, r *http.Request) {
	st, err := getSolar(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	x, err := xml.MarshalIndent(st, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func JsonHandler(w http.ResponseWriter, r *http.Request) {
	st, err := getSolar(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	js, err := json.MarshalIndent(st, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

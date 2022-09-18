package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mmadfox/go-geojson2h3"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
	"github.com/uber/h3-go/v3"
)

func getPoint(latlon string) (*geojson.Point, error) {
	s := strings.Split(latlon, ",")
	if len(s) != 2 {
		return nil, fmt.Errorf("points has to be represented by lat,lon coordinate without spaces")
	}
	lat, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse lat to float64")
	}
	lon, err := strconv.ParseFloat(s[1], 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse lon to float64")
	}
	return geojson.NewPoint(geometry.Point{lat, lon}), nil
}

func getIndexes(start, stop *geojson.Point) ([]h3.H3Index, error) {
	points, err := getRoute(start, stop)
	if err != nil {
		return nil, fmt.Errorf("cannot get points from Graphhopper API: %w", err)
	}

	obj, err := geojson.Parse(string(points), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot parse GeoJSON: %w", err)
	}
	indx, err := geojson2h3.ToH3(11, obj)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to H3: %w", err)
	}
	return indx, nil

}

func getIndexHashes(indx []h3.H3Index) []string {
	res := []string{}
	for _, i := range indx {
		res = append(res, h3.ToString(i))
	}
	return res
}

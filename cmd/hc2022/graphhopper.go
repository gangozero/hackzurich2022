package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/tidwall/geojson"
)

func getRoute(start, stop *geojson.Point) (json.RawMessage, error) {
	apiKey := os.Getenv("GRAPHHOPPER_APIKEY")
	url := fmt.Sprintf("https://graphhopper.com/api/1/route?point=%f,%f&point=%f,%f&profile=truck&points_encoded=false&key=%s", start.Base().X, start.Base().Y, stop.Base().X, stop.Base().Y, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot get response: %w", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %s\n", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting data, http code %d, message: %s\n", resp.StatusCode, string(respBody))
	}

	var payload ResponseGH
	err = json.Unmarshal(respBody, &payload)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal response: %w", err)
	}
	return payload.Paths[0].Points, nil
}

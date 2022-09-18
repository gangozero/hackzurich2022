package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const dbfile string = "server.db"

func newServer() (*server, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, fmt.Errorf("cannot open SQLite file: %w", err)
	}
	return &server{db: db}, nil
}

func (srv *server) suggestDriver(hashes []string, limit int) ([]ResponseDriver, error) {
	query := fmt.Sprintf(`
	SELECT
		DRIVER,
		SUM(STRESS) SUM_STRESS
	FROM
		STRESS
	WHERE H3 IN ('%s')
	GROUP BY
		DRIVER  
	ORDER BY SUM_STRESS ASC
	LIMIT ?;	
`, strings.Join(hashes, "','"))

	rows, err := srv.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting data from DB: %w", err)
	}

	defer rows.Close()
	var drivers []ResponseDriver

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var drv ResponseDriver
		if err := rows.Scan(&drv.User, &drv.Score); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		drivers = append(drivers, drv)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting rows: %w", err)
	}
	return drivers, nil
}

func (srv *server) getDriverFeedback(drivers []string, startHash, stopHash string) ([]driverReview, error) {
	query := fmt.Sprintf(`
	SELECT
		DRIVER,
		REVIEW
	FROM
		FEEDBACK
	WHERE H3_START = ? AND H3_STOP = ? AND
			DRIVER IN ('%s');	
`, strings.Join(drivers, "','"))

	rows, err := srv.db.Query(query, startHash, stopHash)
	if err != nil {
		return nil, fmt.Errorf("error getting data from DB: %w", err)
	}

	defer rows.Close()
	var reviews []driverReview

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var drv driverReview
		if err := rows.Scan(&drv.user, &drv.review); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		reviews = append(reviews, drv)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting rows: %w", err)
	}
	return reviews, nil
}

func errorResponse(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err = fmt.Fprintf(w, "%v", err)
	if err != nil {
		log.Println("Error writing response: ", err)
	}
}

func okResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Error writing response: ", err)
	}
}

func getDriverList(l []ResponseDriver) []string {
	result := []string{}
	for _, r := range l {
		result = append(result, r.User)
	}

	return result
}

func sortResult(inp map[string]float64) []ResponseDriver {
	keys := make([]string, 0, len(inp))

	for key := range inp {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return inp[keys[i]] < inp[keys[j]]
	})

	result := []ResponseDriver{}
	for _, key := range keys {
		result = append(result, ResponseDriver{User: key, Score: inp[key]})
	}

	return result
}

func applyReviews(drivers []ResponseDriver, reviews []driverReview) []ResponseDriver {
	driverMap := map[string]float64{}
	reviewMap := map[string]float64{}
	resultMap := map[string]float64{}

	for _, d := range drivers {
		driverMap[d.User] = d.Score
	}

	for _, r := range reviews {
		reviewMap[r.user] = r.review
	}

	for k, v := range driverMap {
		coef := 0.5
		if r, ok := reviewMap[k]; ok {
			coef = r
		}
		resultMap[k] = v * (1.2 - 0.4*coef)
	}

	return sortResult(resultMap)
}

func (srv *server) getDriverhandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		startPoint, ok := query["start"]
		if !ok || len(startPoint) != 1 {
			errorResponse(w, fmt.Errorf("Start coordinate not found\n"))
			return
		}
		start, err := getPoint(startPoint[0])
		if err != nil {
			errorResponse(w, fmt.Errorf("Error parsing start coordinate: %w", err))
			return
		}

		stopPoint, ok := query["stop"]
		if !ok || len(stopPoint) != 1 {
			errorResponse(w, fmt.Errorf("Stop coordinate not found\n"))
			return
		}
		stop, err := getPoint(stopPoint[0])
		if err != nil {
			errorResponse(w, fmt.Errorf("Error parsing start coordinate: %w", err))
			return
		}

		limit := 1
		_, feedbackFlag := query["feedback"]
		if feedbackFlag {
			limit = 10
		}

		indx, err := getIndexes(start, stop)
		if err != nil {
			errorResponse(w, fmt.Errorf("Error getting indexes: %w", err))
			return
		}
		res := getIndexHashes(indx)

		drivers, err := srv.suggestDriver(res, limit)
		if err != nil {
			errorResponse(w, fmt.Errorf("Error getting suggestion from DB: %w", err))
			return
		}

		if feedbackFlag {
			reviews, err := srv.getDriverFeedback(getDriverList(drivers), res[0], res[len(res)-1])
			if err != nil {
				errorResponse(w, fmt.Errorf("Error getting reviews from DB: %w", err))
				return
			}
			drivers = applyReviews(drivers, reviews)
		}

		okResponse(w, drivers[0])
	}
	return http.HandlerFunc(fn)
}

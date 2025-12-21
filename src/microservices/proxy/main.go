package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

var (
	moviesMigrationPercent                               int = 100
	monolithMigrationPercent, monolithCount, moviesCount int
	monolithURL, moviesURL                               string
	moviesRate, monolithRate                             float32
)

func main() {
	port := getEnv("PORT", "8000")
	monolithURL = getEnv("MONOLITH_URL", "http://monolith:8080")
	moviesURL = getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081")

	gradualMigration := getEnv("GRADUAL_MIGRATION", "false")
	if gradualMigration == "true" {
		var err error

		moviesMigrationPercent, err = strconv.Atoi(getEnv("MOVIES_MIGRATION_PERCENT", "100"))
		if err != nil {
			moviesMigrationPercent = 100
		}

		switch {
		case moviesMigrationPercent < 0:
			moviesMigrationPercent = 0
		case moviesMigrationPercent > 100:
			moviesMigrationPercent = 100
		}
	}

	monolithMigrationPercent = 100 - moviesMigrationPercent
	moviesRate = float32(moviesMigrationPercent) / float32(monolithMigrationPercent)

	if moviesRate < 1 {
		monolithRate = 1 / moviesRate
	}

	proxy, err := NewProxy()
	if err != nil {
		panic(err)
	}

	http.Handle("/", proxy)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func NewProxy() (*httputil.ReverseProxy, error) {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(defineURL())
		},
	}

	return proxy, nil
}

func defineURL() *url.URL {
	var rawURL string

	switch {
	case moviesMigrationPercent == 100:
		rawURL = moviesURL
	case monolithMigrationPercent == 100:
		rawURL = monolithURL
	default:
		if monolithCount+moviesCount >= 100 {
			monolithCount = 0
			moviesCount = 0
		}

		if moviesRate >= 1 {
			if float32(moviesCount/(monolithCount+1)) < moviesRate {
				moviesCount++
				rawURL = moviesURL
			} else {
				monolithCount++
				rawURL = monolithURL
			}
		} else {
			if float32(monolithCount/(moviesCount+1)) < monolithRate {
				monolithCount++
				rawURL = monolithURL
			} else {
				moviesCount++
				rawURL = moviesURL
			}
		}
	}

	url, _ := url.Parse(rawURL)

	log.Printf("INFO: movies: %v, monolith: %v, moviesRate: %v, monolithRate: %v", moviesCount, monolithCount, moviesRate, monolithRate)

	return url
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-httpproxy/httpproxy"
)

var (
	moviesMigrationPercent                               int = 100
	monolithMigrationPercent, monolithCount, moviesCount int
	monolithURL, moviesUrl                               string
	moviesRate, monolithRate                             float32
)

func main() {
	port := getEnv("PORT", "8000")
	monolithURL = getEnv("MONOLITH_URL", "http://monolith:8080")
	moviesUrl = getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081")

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
	moviesRate = float32(moviesMigrationPercent / monolithMigrationPercent)

	if moviesRate < 1 {
		monolithRate = 1 / moviesRate
	}

	// Create a new proxy with default certificate pair.
	prx, _ := httpproxy.NewProxy()

	// Set handlers.
	prx.OnError = OnError
	prx.OnAccept = OnAccept
	prx.OnAuth = OnAuth
	prx.OnConnect = OnConnect
	prx.OnRequest = OnRequest
	prx.OnResponse = OnResponse

	// Listen...
	http.ListenAndServe(":"+port, prx)
}

func OnError(ctx *httpproxy.Context, where string,
	err *httpproxy.Error, opErr error) {
	// Log errors.
	log.Printf("ERR: %s: %s [%s]", where, err, opErr)
}

func OnAccept(ctx *httpproxy.Context, w http.ResponseWriter,
	r *http.Request) bool {
	// Handle local request has path "/info"
	if r.Method == "GET" && !r.URL.IsAbs() && r.URL.Path == "/info" {
		w.Write([]byte("This is go-httpproxy."))
		return true
	}
	return false
}

func OnAuth(ctx *httpproxy.Context, authType string, user string, pass string) bool {
	// Auth test user.
	if user == "test" && pass == "1234" {
		return true
	}
	return false
}

func OnConnect(ctx *httpproxy.Context, host string) (
	ConnectAction httpproxy.ConnectAction, newHost string) {

	switch {
	case moviesMigrationPercent == 100:
		host = moviesUrl
	case monolithMigrationPercent == 100:
		host = monolithURL
	default:
		if monolithCount+moviesCount >= 100 {
			monolithCount = 0
			moviesCount = 0
		}

		if moviesRate >= 1 {
			if float32(moviesCount/(monolithCount+1)) < moviesRate {
				moviesCount++
				host = moviesUrl
			} else {
				monolithCount++
				host = monolithURL
			}
		} else {
			if float32(monolithCount/(moviesCount+1)) < monolithRate {
				monolithCount++
				host = monolithURL
			} else {
				moviesCount++
				host = moviesUrl
			}
		}
	}

	log.Printf("INFO: movies: %v, monolith: %v, moviesRate: %v, monolithRate: %v", moviesCount, monolithCount, moviesRate, monolithRate)
	return httpproxy.ConnectProxy, host
}

func OnRequest(ctx *httpproxy.Context, req *http.Request) (
	resp *http.Response) {
	// Log proxying requests.
	log.Printf("INFO: Proxy: %s %s", req.Method, req.URL.String())
	return
}

func OnResponse(ctx *httpproxy.Context, req *http.Request,
	resp *http.Response) {
	// Add header "Via: go-httpproxy".
	resp.Header.Add("Via", "go-httpproxy")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

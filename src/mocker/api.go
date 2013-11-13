package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MockApi struct {
	EndpointsMutex sync.Mutex
	Endpoints      map[string]*Endpoint
}

func NewMockApi() *MockApi {
	api := new(MockApi)
	api.Endpoints = make(map[string]*Endpoint)
	return api
}

func (api *MockApi) Register(router *mux.Router) {
	router.HandleFunc("/response/{status:[2345][0-9][0-9]}", api.ResponseHandler).Methods("GET")
	router.HandleFunc("/response/{status:[2345][0-9][0-9]}/{path:[- \\w\\/]+}", api.ResponseHandler).Methods("GET")

	router.HandleFunc("/timeout/{timeout:[0-9]+}", api.TimeoutHandler).Methods("GET")
	router.HandleFunc("/timeout/{timeout:[0-9]+}/{path:[- \\w\\/]+}", api.TimeoutHandler).Methods("GET")
	router.HandleFunc("/timeout/{min:[0-9]+},{median:[0-9]+},{max:[0-9]+}", api.TimeoutHandler).Methods("GET")
	router.HandleFunc("/timeout/{min:[0-9]+},{median:[0-9]+},{max:[0-9]+}", api.TimeoutHandler).Methods("GET")

	router.HandleFunc("/mock", api.MockHandler).Methods("POST")
	router.HandleFunc("/mock/{endpoint}", api.MockEndpointHandler).Methods("POST")
	router.HandleFunc("/mock/{endpoint}/{path:[- \\w\\/]+}", api.MockEndpointHandler).Methods("POST")

	router.HandleFunc("/endpoint/{endpoint}", api.EndpointHandler).Methods("GET")
	router.HandleFunc("/endpoint/{endpoint}/{path:[- \\w\\/]+}", api.EndpointHandler).Methods("GET")

	router.HandleFunc("/settings/{endpoint}", api.SettingsHandler).Methods("GET", "POST")
}

// /response/{status:[2345][0-9][0-9]}
// Returns the specified status code
func (api *MockApi) ResponseHandler(rw http.ResponseWriter, req *http.Request) {
	status, _ := strconv.Atoi(mux.Vars(req)["status"])
	rw.WriteHeader(status)
}

// /mock
// Creates a uuid endpoint which will always return 200 and the specified payload
func (api *MockApi) MockHandler(rw http.ResponseWriter, req *http.Request) {
	api.EndpointsMutex.Lock()
	defer api.EndpointsMutex.Unlock()
	endpointName := uuid.NewUUID().String()
	endpoint := NewEndpoint()
	endpoint.AddResponse(req, "")
	api.Endpoints[endpointName] = endpoint
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(endpointName))
}

// /mock/{endpoint}
// Adds to the endpoint for the given query parameters, when requested with those parameters will return the specified payload
func (api *MockApi) MockEndpointHandler(rw http.ResponseWriter, req *http.Request) {
	api.EndpointsMutex.Lock()
	defer api.EndpointsMutex.Unlock()
	vars := mux.Vars(req)
	endpointName := vars["endpoint"]
	path := vars["path"]
	endpoint, endpointOk := api.Endpoints[endpointName]
	if !endpointOk {
		endpoint = NewEndpoint()
		api.Endpoints[endpointName] = endpoint
	}
	endpoint.AddResponse(req, path)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(endpointName))
}

func (api *MockApi) serveEndpoint(endpoint *Endpoint, rw http.ResponseWriter, req *http.Request) {
	path := mux.Vars(req)["path"]
	response, responseOk := endpoint.Lookup(req, path)
	if !responseOk {
		http.Error(rw, "No response for given parameters", http.StatusBadRequest)
		return
	}
	endpoint.LatencyInjector()
	rw.Header().Set("Content-Type", response.ContentType)
	rw.WriteHeader(http.StatusOK)
	rw.Write(response.Content)
}

// /endpoint/{endpoint}
// Mocked request, matches against the endpoint and option query parameters to find and return the payload, 400 if no parameter, 404 is no endpoint
func (api *MockApi) EndpointHandler(rw http.ResponseWriter, req *http.Request) {
	endpointName := mux.Vars(req)["endpoint"]
	endpoint, endpointOk := api.Endpoints[endpointName]
	if !endpointOk {
		http.Error(rw, "Endpoint does not exist", http.StatusNotFound)
		return
	}
	api.serveEndpoint(endpoint, rw, req)
}

// /settings/{endpoint}
// Sets a value for the given endpoint
func (api *MockApi) SettingsHandler(rw http.ResponseWriter, req *http.Request) {
	api.EndpointsMutex.Lock()
	defer api.EndpointsMutex.Unlock()
	endpointName := mux.Vars(req)["endpoint"]
	endpoint, endpointOk := api.Endpoints[endpointName]
	if !endpointOk {
		http.Error(rw, "Endpoint does not exist", http.StatusNotFound)
		return
	}
	latency := req.FormValue("latency")
	if latency != "" {
		if latency == "static" {
			latencyMsStr := req.FormValue("latency_ms")
			latencyMs, latencyMsErr := strconv.ParseInt(latencyMsStr, 10, 64)
			if latencyMsErr != nil {
				fmt.Fprintf(rw, "Invalid static latency valid %q\n", latencyMsErr.Error())
			}
			endpoint.SetLatencyInjector(NewStaticLatencyInjector(latencyMs))
		} else if latency == "normal" {
			latencyMinMsStr := req.FormValue("latency_min_ms")
			latencyMedianMsStr := req.FormValue("latency_median_ms")
			latencyMaxMsStr := req.FormValue("latency_max_ms")
			latencyMinMs, latencyMinMsErr := strconv.ParseInt(latencyMinMsStr, 10, 64)
			latencyMedianMs, latencyMedianMsErr := strconv.ParseInt(latencyMedianMsStr, 10, 64)
			latencyMaxMs, latencyMaxMsErr := strconv.ParseInt(latencyMaxMsStr, 10, 64)
			hasLatencyErr := (latencyMinMsErr != nil || latencyMedianMsErr != nil || latencyMaxMsErr != nil)
			if hasLatencyErr {
				rw.WriteHeader(http.StatusBadRequest)
			}
			if latencyMinMsErr != nil {
				fmt.Fprintf(rw, "Invalid minimum latency valid %q\n", latencyMinMsErr.Error())
			}
			if latencyMedianMsErr != nil {
				fmt.Fprintf(rw, "Invalid median latency valid %q\n", latencyMedianMsErr.Error())
			}
			if latencyMaxMsErr != nil {
				fmt.Fprintf(rw, "Invalid maximum latency valid %q\n", latencyMaxMsErr.Error())
			}
			if hasLatencyErr {
				return
			}
			endpoint.SetLatencyInjector(NewNormalLatencyInjector(time.Now().UTC().UnixNano(), float64(latencyMinMs), float64(latencyMedianMs), float64(latencyMaxMs)))
		} else if latency == "none" {
			endpoint.SetLatencyInjector(NewNoLatencyInjector())
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "Invalid latency %q\n", latency)
			return
		}
	}
	rw.WriteHeader(http.StatusOK)
}

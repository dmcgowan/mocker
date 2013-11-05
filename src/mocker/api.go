package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"sync"
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
	router.HandleFunc("/mock", api.MockHandler).Methods("POST")
	router.HandleFunc("/mock/{endpoint}", api.MockEndpointHandler).Methods("POST")
	router.HandleFunc("/endpoint/{endpoint}", api.EndpointHandler).Methods("GET")
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
	api.Endpoints[endpointName] = NewEndpoint(req)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(endpointName))
}

// /mock/{endpoint}
// Adds to the endpoint for the given query parameters, when requested with those parameters will return the specified payload
func (api *MockApi) MockEndpointHandler(rw http.ResponseWriter, req *http.Request) {
	api.EndpointsMutex.Lock()
	defer api.EndpointsMutex.Unlock()
	endpointName := mux.Vars(req)["endpoint"]
	endpoint, endpointOk := api.Endpoints[endpointName]
	if !endpointOk {
		endpoint := NewEndpoint(req)
		api.Endpoints[endpointName] = endpoint
	} else {
		endpoint.AddResponse(req)
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(endpointName))
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
	response, responseOk := endpoint.Lookup(req)
	if !responseOk {
		http.Error(rw, "No response for given parameters", http.StatusBadRequest)
		return
	}
	rw.Header().Set("Content-Type", response.ContentType)
	rw.WriteHeader(http.StatusOK)
	rw.Write(response.Content)
}

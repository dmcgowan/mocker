package main

import (
	"crypto/sha1"
	"hash"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

type Response struct {
	ContentType string
	Content     []byte
}

type Endpoint struct {
	Endpoints map[string]*Response
}

func addSortedKeys(values []string, h hash.Hash) {
	if len(values) > 0 {
		if len(values) == 1 {
			h.Write([]byte(values[0]))
			h.Write([]byte{0x00})
		} else {
			sort.Strings(values)
			for _, value := range values {
				h.Write([]byte(value))
				h.Write([]byte{0x00})
			}
		}
	}
}

type UrlValue struct {
	Key    string
	Values []string
}

type UrlValueSlice []*UrlValue

func (slice UrlValueSlice) Len() int {
	return len(slice)
}

func (slice UrlValueSlice) Less(i, j int) bool {
	return slice[i].Key < slice[j].Key
}

func (slice UrlValueSlice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func calculateHash(values url.Values, path string) string {
	hash := sha1.New()
	hash.Write([]byte(path))
	hash.Write([]byte{0x00})
	if len(values) > 0 {
		if len(values) == 1 {
			for key, value := range values {
				hash.Write([]byte(key))
				hash.Write([]byte{0x00})
				addSortedKeys(value, hash)
			}
		} else {
			urlValues := make(UrlValueSlice, 0, len(values))
			for key, value := range values {
				urlValue := new(UrlValue)
				urlValue.Key = key
				urlValue.Values = value
				urlValues = append(urlValues, urlValue)
			}
			sort.Sort(urlValues)
			for _, sortedUrlValue := range urlValues {
				hash.Write([]byte(sortedUrlValue.Key))
				hash.Write([]byte{0x00})
				addSortedKeys(sortedUrlValue.Values, hash)
			}
		}
	}
	return string(hash.Sum([]byte{0x00}))
}

func NewEndpoint() *Endpoint {
	endpoint := new(Endpoint)
	endpoint.Endpoints = make(map[string]*Response)
	return endpoint
}

func (endpoint *Endpoint) AddResponse(req *http.Request, path string) {
	response := new(Response)
	response.ContentType = req.Header.Get("Content-Type")
	response.Content, _ = ioutil.ReadAll(req.Body)
	endpoint.Endpoints[calculateHash(req.URL.Query(), path)] = response
}

func (endpoint *Endpoint) Lookup(req *http.Request, path string) (*Response, bool) {
	r, rOk := endpoint.Endpoints[calculateHash(req.URL.Query(), path)]
	return r, rOk
}

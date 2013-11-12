package main

import (
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type LatencyInjector func()

func NewNoLatencyInjector() LatencyInjector {
	return func() {}
}

func NewStaticLatencyInjector(waitforms int64) LatencyInjector {
	return func() {
		time.Sleep(time.Millisecond * time.Duration(waitforms))
	}
}

func NewNormalLatencyInjector(seed int64, minms, medianms, maxms float64) LatencyInjector {
	r := rand.New(rand.NewSource(seed))
	leftdev := (medianms - minms) / 2.328
	rightdev := (maxms - medianms) / 2.328
	return func() {
		norm := r.NormFloat64()
		var sleepms float64
		if norm < 0 {
			randval := norm*leftdev + medianms
			if randval < minms {
				sleepms = minms
			} else {
				sleepms = randval
			}
		} else {
			randval := norm*rightdev + medianms
			if randval > maxms {
				sleepms = maxms
			} else {
				sleepms = randval
			}
		}
		time.Sleep(time.Duration(float64(time.Millisecond) * sleepms))
	}
}

// /timeout/{timeout:[0-9]+}
// /timeout/{min:[0-9]+},{median:[0-9]+},{max:[0-9]+}
// An endpoint that will wait the specified amount of time before returning 200
// Can take a single value or a 3 values to form a normal distribution
func (api *MockApi) TimeoutHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	timeoutStr, timeoutOk := vars["timeout"]
	if !timeoutOk {
		minStr, minOk := vars["min"]
		medianStr, medianOk := vars["median"]
		maxStr, maxOk := vars["max"]
		if !minOk || !medianOk || !maxOk {
			rw.WriteHeader(http.StatusBadRequest)
		} else {
			min, _ := strconv.ParseFloat(minStr, 64)
			median, _ := strconv.ParseFloat(medianStr, 64)
			max, _ := strconv.ParseFloat(maxStr, 64)
			NewNormalLatencyInjector(time.Now().UTC().UnixNano(), min, median, max)()
		}
	} else {
		timeout, _ := strconv.ParseInt(timeoutStr, 10, 64)
		NewStaticLatencyInjector(timeout)()
		rw.WriteHeader(http.StatusOK)
	}
}

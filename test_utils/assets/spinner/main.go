package main

import (
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var spinFlag int32
var niceMillis int64 = 1

func main() {
	http.HandleFunc("/spin", func(w http.ResponseWriter, r *http.Request) {

		go func() {
			markSpinning()
			for {
				if !isSpinning() {
					return
				}
				// be nice!
				time.Sleep(time.Millisecond)
				// spin!
			}
		}()

		if spinTime, ok := getSpinTime(r); ok {
			time.AfterFunc(time.Duration(spinTime)*time.Millisecond, func() {
				atomic.StoreInt32(&spinFlag, 0)
			})
		}
	})

	http.HandleFunc("/unspin", func(w http.ResponseWriter, r *http.Request) {
		markNotSpinning()
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}

func getSpinTime(r *http.Request) (int, bool) {
	millis, ok := r.URL.Query()["spinTime"]
	if !ok || len(millis) == 0 {
		return 0, false
	}
	ms, err := strconv.Atoi(millis[0])
	if err != nil {
		return 0, false
	}

	return ms, true
}

func markSpinning() {
	atomic.StoreInt32(&spinFlag, 1)
}

func markNotSpinning() {
	atomic.StoreInt32(&spinFlag, 0)
}

func isSpinning() bool {
	return atomic.LoadInt32(&spinFlag) == 1
}

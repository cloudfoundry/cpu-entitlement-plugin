package main

import (
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var spinning int32
var niceNS int64

func main() {
	http.HandleFunc("/spin", func(w http.ResponseWriter, r *http.Request) {
		myVal := atomic.AddInt32(&spinning, 1)
		go func() {
			for {
				isSpinning := atomic.LoadInt32(&spinning)
				if isSpinning < myVal {
					break
				}
				nanos := atomic.LoadInt64(&niceNS)
				time.Sleep(time.Duration(nanos) * time.Nanosecond)
				// spin!
			}
		}()
	})

	http.HandleFunc("/unspin", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&spinning, -1)
	})

	http.HandleFunc("/nice", func(w http.ResponseWriter, r *http.Request) {
		nanoVals, ok := r.URL.Query()["nanos"]
		if !ok || len(nanoVals) < 1 {
			return
		}
		val, err := strconv.Atoi(nanoVals[0])
		if err != nil {
			return
		}
		atomic.StoreInt64(&niceNS, int64(val))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}

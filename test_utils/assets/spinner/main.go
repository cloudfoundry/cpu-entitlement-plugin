package main

import (
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var spinning int32
var niceMicros int64

func main() {
	niceMicros = 1000
	http.HandleFunc("/spin", func(w http.ResponseWriter, r *http.Request) {
		myVal := atomic.AddInt32(&spinning, 1)
		go func() {
			for {
				isSpinning := atomic.LoadInt32(&spinning)
				if isSpinning < myVal {
					break
				}
				micros := atomic.LoadInt64(&niceMicros)
				time.Sleep(time.Duration(micros) * time.Microsecond)
				// spin!
			}
		}()
	})

	http.HandleFunc("/unspin", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&spinning, -1)
	})

	http.HandleFunc("/nice", func(w http.ResponseWriter, r *http.Request) {
		microVals, ok := r.URL.Query()["micros"]
		if !ok || len(microVals) < 1 {
			return
		}
		val, err := strconv.Atoi(microVals[0])
		if err != nil {
			return
		}
		atomic.StoreInt64(&niceMicros, int64(val))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}

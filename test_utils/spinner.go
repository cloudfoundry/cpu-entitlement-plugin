package test_utils

import (
	"crypto/tls"
	"net/http"
	"time"

	. "github.com/onsi/gomega"
)

func Spin(appURL string) {
	httpGet(appURL + "/spin")
}

func Unspin(appURL string) {
	httpGet(appURL + "/unspin")
}

func SpinFor(appURL string, duration time.Duration) {
	Spin(appURL)
	defer Unspin(appURL)
	time.Sleep(duration)
}

func httpGet(url string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(url)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
}

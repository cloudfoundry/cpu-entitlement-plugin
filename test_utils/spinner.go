package test_utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func Spin(appURL string, spinSeconds int) {
	url := appURL + "/spin"
	if spinSeconds > 0 {
		url += fmt.Sprintf("?spinTime=%d", spinSeconds*1000)
	}
	httpGet(url)
}

func Unspin(appURL string) {
	httpGet(appURL + "/unspin")
}

func SpinFor(appURL string, duration time.Duration) {
	Spin(appURL, 0)
	defer Unspin(appURL)
	time.Sleep(duration)
}

func PushSpinner(appName string, instances int) {
	ExpectWithOffset(1, Cmd("cf", "push", appName, "-i", strconv.Itoa(instances)).WithDir("../test_utils/assets/spinner").WithTimeout("3m").Run()).To(gexec.Exit(0))
}

func httpGet(url string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(url)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, resp.StatusCode).To(Equal(http.StatusOK))
}

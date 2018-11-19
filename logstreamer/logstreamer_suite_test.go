package logstreamer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLogstreamer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logstreamer Suite")
}

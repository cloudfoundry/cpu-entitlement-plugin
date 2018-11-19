package usagemetric_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUsagemetric(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Usagemetric Suite")
}

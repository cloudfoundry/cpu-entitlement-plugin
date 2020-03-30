package output_test

import (
	"testing"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOutput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Output Suite")
}

var (
	logger lager.Logger
)

var _ = BeforeSuite(func() {
	logger = lagertest.NewTestLogger("output-test")
})

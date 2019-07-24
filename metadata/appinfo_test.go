package metadata_test

import (
	"errors"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "code.cloudfoundry.org/cpu-entitlement-plugin/metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CFAppInfo", func() {

	var (
		infoGetter InfoGetter
		cli        *pluginfakes.FakeCliConnection
		info       CFAppInfo
		err        error
	)

	JustBeforeEach(func() {
		info, err = infoGetter.GetCFAppInfo("myapp")
	})

	BeforeEach(func() {
		cli = new(pluginfakes.FakeCliConnection)

		cli.GetAppReturns(plugin_models.GetAppModel{Guid: "qwerty"}, nil)
		cli.UsernameReturns("infouser", nil)
		cli.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "currentorg"}}, nil)
		cli.GetCurrentSpaceReturns(plugin_models.Space{SpaceFields: plugin_models.SpaceFields{Name: "currentspace"}}, nil)

		infoGetter = NewInfoGetter(cli)
	})

	It("gets the application info", func() {
		Expect(err).NotTo(HaveOccurred())

		Expect(cli.GetAppCallCount()).To(Equal(1))
		actualAppName := cli.GetAppArgsForCall(0)
		Expect(actualAppName).To(Equal("myapp"))

		Expect(info.App.Guid).To(Equal("qwerty"))
		Expect(info.Username).To(Equal("infouser"))
		Expect(info.Org).To(Equal("currentorg"))
		Expect(info.Space).To(Equal("currentspace"))
	})

	Context("get app errors", func() {
		BeforeEach(func() {
			cli.GetAppReturns(plugin_models.GetAppModel{}, errors.New("app error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("app error"))
		})
	})

	Context("get username errors", func() {
		BeforeEach(func() {
			cli.UsernameReturns("", errors.New("username error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("username error"))
		})
	})

	Context("get org errors", func() {
		BeforeEach(func() {
			cli.GetCurrentOrgReturns(plugin_models.Organization{}, errors.New("org error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("org error"))
		})
	})

	Context("get space errors", func() {
		BeforeEach(func() {
			cli.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("space error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("space error"))
		})
	})
})

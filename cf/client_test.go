package cf_test

import (
	"errors"
	"fmt"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-plugin/cf"
	"code.cloudfoundry.org/cpu-entitlement-plugin/cf/cffakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		fakeCli  *cffakes.FakeCli
		cfClient cf.Client
		err      error
		spaces   []cf.Space
	)

	BeforeEach(func() {
		fakeCli = new(cffakes.FakeCli)
		cfClient = cf.NewClient(fakeCli)

		fakeCli.GetSpacesReturns([]plugin_models.GetSpaces_Model{
			{Guid: "space1-guid", Name: "space-1"},
			{Guid: "space2-guid", Name: "space-2"},
		}, nil)

		fakeCli.GetSpaceStub = func(spaceName string) (plugin_models.GetSpace_Model, error) {
			switch spaceName {
			case "space-1":
				return plugin_models.GetSpace_Model{
					Applications: []plugin_models.GetSpace_Apps{
						{Name: "app-1", Guid: "space-1-app-1-guid"},
						{Name: "app-2", Guid: "space-1-app-2-guid"},
					},
				}, nil
			case "space-2":
				return plugin_models.GetSpace_Model{
					Applications: []plugin_models.GetSpace_Apps{
						{Name: "app-1", Guid: "space-2-app-1-guid"},
					},
				}, nil
			}
			return plugin_models.GetSpace_Model{}, fmt.Errorf("Space '%s' not found", spaceName)
		}
	})

	JustBeforeEach(func() {
		spaces, err = cfClient.GetSpaces()
	})

	It("fetches all spaces", func() {
		Expect(err).NotTo(HaveOccurred())
		Expect(spaces).To(Equal([]cf.Space{
			{
				Name: "space-1",
				Applications: []cf.Application{
					{Name: "app-1", Guid: "space-1-app-1-guid"},
					{Name: "app-2", Guid: "space-1-app-2-guid"},
				},
			},
			{
				Name: "space-2",
				Applications: []cf.Application{
					{Name: "app-1", Guid: "space-2-app-1-guid"},
				},
			},
		}))
	})

	When("fetching the list of spaces fails", func() {
		BeforeEach(func() {
			fakeCli.GetSpacesReturns(nil, errors.New("get-spaces-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("get-spaces-error"))
		})
	})

	When("fetching details about a space fails", func() {
		BeforeEach(func() {
			fakeCli.GetSpaceReturns(plugin_models.GetSpace_Model{}, errors.New("get-space-error"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("get-space-error"))
		})
	})
})

package cf_test

import (
	"errors"
	"fmt"
	"time"

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
	)

	BeforeEach(func() {
		fakeCli = new(cffakes.FakeCli)
		cfClient = cf.NewClient(fakeCli)

	})

	Describe("Spaces", func() {
		var spaces []cf.Space

		BeforeEach(func() {
			fakeCli.GetCurrentSpaceReturns(plugin_models.Space{SpaceFields: plugin_models.SpaceFields{Name: "the-space"}}, nil)
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
						{Name: "app-1", Guid: "space-1-app-1-guid", Space: "space-1"},
						{Name: "app-2", Guid: "space-1-app-2-guid", Space: "space-1"},
					},
				},
				{
					Name: "space-2",
					Applications: []cf.Application{
						{Name: "app-1", Guid: "space-2-app-1-guid", Space: "space-2"},
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

	Describe("Application", func() {
		var application cf.Application

		BeforeEach(func() {
			fakeCli.GetCurrentSpaceReturns(plugin_models.Space{SpaceFields: plugin_models.SpaceFields{Name: "the-space"}}, nil)
			fakeCli.UsernameReturns("the-user", nil)
			fakeCli.GetAppReturns(plugin_models.GetAppModel{
				Guid: "qwerty",
				Name: "YTREWQ",
				Instances: []plugin_models.GetApp_AppInstanceFields{
					plugin_models.GetApp_AppInstanceFields{Since: time.Unix(123, 456)},
					plugin_models.GetApp_AppInstanceFields{Since: time.Unix(789, 0)},
				},
			}, nil)
		})

		JustBeforeEach(func() {
			application, err = cfClient.GetApplication("myapp")
		})

		It("gets the application info", func() {
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCli.GetAppCallCount()).To(Equal(1))
			actualAppName := fakeCli.GetAppArgsForCall(0)
			Expect(actualAppName).To(Equal("myapp"))

			Expect(application.Guid).To(Equal("qwerty"))
			Expect(application.Name).To(Equal("YTREWQ"))
			Expect(application.Space).To(Equal("the-space"))
		})

		It("gets the application instances", func() {
			Expect(len(application.Instances)).To(Equal(2))
			Expect(application.Instances[0].InstanceID).To(Equal(0))
			Expect(application.Instances[0].Since).To(Equal(time.Unix(123, 456)))
			Expect(application.Instances[1].InstanceID).To(Equal(1))
			Expect(application.Instances[1].Since).To(Equal(time.Unix(789, 0)))
		})

		When("get app errors", func() {
			BeforeEach(func() {
				fakeCli.GetAppReturns(plugin_models.GetAppModel{}, errors.New("app error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("app error"))
			})
		})

		When("get space errors", func() {
			BeforeEach(func() {
				fakeCli.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("space error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("space error"))
			})
		})
	})

	Describe("CurrentOrg", func() {
		var (
			org string
			err error
		)

		BeforeEach(func() {
			fakeCli.GetCurrentOrgReturns(plugin_models.Organization{OrganizationFields: plugin_models.OrganizationFields{Name: "the-org"}}, nil)
		})

		JustBeforeEach(func() {
			org, err = cfClient.GetCurrentOrg()
		})

		It("returns the org", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(org).To(Equal("the-org"))
		})

		When("get current org errors", func() {
			BeforeEach(func() {
				fakeCli.GetCurrentOrgReturns(plugin_models.Organization{}, errors.New("org error"))
			})
			It("returns the error", func() {
				Expect(err).To(MatchError("org error"))
			})
		})
	})

	Describe("CurrentSpace", func() {
		var (
			space string
			err   error
		)

		BeforeEach(func() {
			fakeCli.GetCurrentSpaceReturns(plugin_models.Space{SpaceFields: plugin_models.SpaceFields{Name: "the-space"}}, nil)
		})

		JustBeforeEach(func() {
			space, err = cfClient.GetCurrentSpace()
		})

		It("returns the space", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(space).To(Equal("the-space"))
		})

		When("get current space errors", func() {
			BeforeEach(func() {
				fakeCli.GetCurrentSpaceReturns(plugin_models.Space{}, errors.New("space error"))
			})
			It("returns the error", func() {
				Expect(err).To(MatchError("space error"))
			})
		})
	})

	Describe("Username", func() {
		var (
			user string
			err  error
		)

		BeforeEach(func() {
			fakeCli.UsernameReturns("the-user", nil)
		})

		JustBeforeEach(func() {
			user, err = cfClient.Username()
		})

		It("returns the user", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(user).To(Equal("the-user"))
		})

		When("get username errors", func() {
			BeforeEach(func() {
				fakeCli.UsernameReturns("", errors.New("username error"))
			})
			It("returns the error", func() {
				Expect(err).To(MatchError("username error"))
			})
		})
	})
})

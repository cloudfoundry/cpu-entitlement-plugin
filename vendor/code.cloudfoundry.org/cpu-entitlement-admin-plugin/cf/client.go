package cf

import (
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"
)

//go:generate counterfeiter . Cli

type Cli interface {
	GetSpaces() ([]plugin_models.GetSpaces_Model, error)
	GetSpace(spaceName string) (plugin_models.GetSpace_Model, error)
}

type Client struct {
	cli Cli
}

func NewClient(cli Cli) Client {
	return Client{cli: cli}
}

func (c Client) GetSpaces() ([]reporter.Space, error) {
	var spaces []reporter.Space

	cfSpaces, err := c.cli.GetSpaces()
	if err != nil {
		return nil, err
	}

	for _, cfSpace := range cfSpaces {
		cfSpaceDetails, err := c.cli.GetSpace(cfSpace.Name)
		if err != nil {
			return nil, err
		}

		var applications []reporter.Application
		for _, cfApp := range cfSpaceDetails.Applications {
			applications = append(applications, reporter.Application{Guid: cfApp.Guid, Name: cfApp.Name})
		}

		spaces = append(spaces, reporter.Space{Name: cfSpace.Name, Applications: applications})
	}

	return spaces, nil
}

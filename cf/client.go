package cf

import (
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
)

//go:generate counterfeiter . Cli

type Cli interface {
	GetSpaces() ([]plugin_models.GetSpaces_Model, error)
	GetSpace(spaceName string) (plugin_models.GetSpace_Model, error)
}

type Space struct {
	Name         string
	Applications []Application
}

type Application struct {
	Name string
	Guid string
}

type Client struct {
	cli Cli
}

func NewClient(cli Cli) Client {
	return Client{cli: cli}
}

func (c Client) GetSpaces() ([]Space, error) {
	var spaces []Space

	cfSpaces, err := c.cli.GetSpaces()
	if err != nil {
		return nil, err
	}

	for _, cfSpace := range cfSpaces {
		cfSpaceDetails, err := c.cli.GetSpace(cfSpace.Name)
		if err != nil {
			return nil, err
		}

		var applications []Application
		for _, cfApp := range cfSpaceDetails.Applications {
			applications = append(applications, Application{Guid: cfApp.Guid, Name: cfApp.Name})
		}

		spaces = append(spaces, Space{Name: cfSpace.Name, Applications: applications})
	}

	return spaces, nil
}

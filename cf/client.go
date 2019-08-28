package cf

import (
	"time"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
)

//go:generate counterfeiter . Cli

type Cli interface {
	GetApp(string) (plugin_models.GetAppModel, error)
	GetCurrentOrg() (plugin_models.Organization, error)
	GetCurrentSpace() (plugin_models.Space, error)
	GetSpace(spaceName string) (plugin_models.GetSpace_Model, error)
	GetSpaces() ([]plugin_models.GetSpaces_Model, error)
	Username() (string, error)
}

type Space struct {
	Name         string
	Applications []Application
}

type Application struct {
	Name      string
	Guid      string
	Username  string
	Org       string
	Space     string
	Instances map[int]Instance
}

type Instance struct {
	InstanceID int
	Since      time.Time
}

type Client struct {
	cli Cli
}

func NewClient(cli Cli) Client {
	return Client{cli: cli}
}

func (c Client) GetSpaces() ([]Space, error) {
	org, err := c.cli.GetCurrentOrg()
	if err != nil {
		return nil, err
	}

	user, err := c.cli.Username()
	if err != nil {
		return nil, err
	}

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
			applications = append(applications, Application{Guid: cfApp.Guid, Name: cfApp.Name, Username: user, Org: org.Name, Space: cfSpace.Name})
		}

		spaces = append(spaces, Space{Name: cfSpace.Name, Applications: applications})
	}

	return spaces, nil
}

func (c Client) GetApplication(appName string) (Application, error) {
	app, err := c.cli.GetApp(appName)
	if err != nil {
		return Application{}, err
	}

	user, err := c.cli.Username()
	if err != nil {
		return Application{}, err
	}

	org, err := c.cli.GetCurrentOrg()
	if err != nil {
		return Application{}, err
	}

	space, err := c.cli.GetCurrentSpace()
	if err != nil {
		return Application{}, err
	}

	instances := map[int]Instance{}
	for id, instance := range app.Instances {
		instances[id] = Instance{InstanceID: id, Since: instance.Since}
	}

	return Application{Name: app.Name, Guid: app.Guid, Username: user, Org: org.Name, Space: space.Name, Instances: instances}, nil
}

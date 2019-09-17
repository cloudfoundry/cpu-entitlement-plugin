package cf

import (
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

//go:generate counterfeiter . ProcessInstanceIDFetcher
type ProcessInstanceIDFetcher interface {
	Fetch(appGUID string) (map[int]string, error)
}

type Space struct {
	Name         string
	Applications []Application
}

type Application struct {
	Name      string
	Guid      string
	Space     string
	Instances map[int]Instance
}

type Instance struct {
	InstanceID        int
	ProcessInstanceID string
}

type Client struct {
	cli                      Cli
	processInstanceIDFetcher ProcessInstanceIDFetcher
}

func NewClient(cli Cli, processInstanceIDFetcher ProcessInstanceIDFetcher) Client {
	return Client{cli: cli, processInstanceIDFetcher: processInstanceIDFetcher}
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
			processInstanceIDs, err := c.processInstanceIDFetcher.Fetch(cfApp.Guid)
			if err != nil {
				return []Space{}, err
			}

			instances := map[int]Instance{}
			for instanceID, processInstanceID := range processInstanceIDs {
				instances[instanceID] = Instance{InstanceID: instanceID, ProcessInstanceID: processInstanceID}
			}
			applications = append(applications, Application{Guid: cfApp.Guid, Name: cfApp.Name, Space: cfSpace.Name, Instances: instances})
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

	space, err := c.cli.GetCurrentSpace()
	if err != nil {
		return Application{}, err
	}

	processInstanceIDs, err := c.processInstanceIDFetcher.Fetch(app.Guid)
	if err != nil {
		return Application{}, err
	}

	instances := map[int]Instance{}
	for id, _ := range app.Instances {
		processInstanceID, hasProcessInstanceID := processInstanceIDs[id]
		if !hasProcessInstanceID {
			continue
		}
		instances[id] = Instance{InstanceID: id, ProcessInstanceID: processInstanceID}
	}

	return Application{Name: app.Name, Guid: app.Guid, Space: space.Name, Instances: instances}, nil
}

func (c Client) GetCurrentOrg() (string, error) {
	org, err := c.cli.GetCurrentOrg()
	if err != nil {
		return "", err
	}
	return org.Name, nil
}

func (c Client) GetCurrentSpace() (string, error) {
	space, err := c.cli.GetCurrentSpace()
	if err != nil {
		return "", err
	}
	return space.Name, nil
}

func (c Client) Username() (string, error) {
	return c.cli.Username()
}

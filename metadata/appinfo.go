package metadata

import (
	"time"

	"code.cloudfoundry.org/cli/plugin"
)

type CFAppInfo struct {
	Name      string
	Guid      string
	Username  string
	Org       string
	Space     string
	Instances map[int]CFAppInstance
}

type CFAppInstance struct {
	InstanceID int
	Since      time.Time
}

type InfoGetter struct {
	cli plugin.CliConnection
}

func NewInfoGetter(cli plugin.CliConnection) InfoGetter {
	return InfoGetter{cli: cli}
}

func (g InfoGetter) GetCFAppInfo(appName string) (CFAppInfo, error) {
	app, err := g.cli.GetApp(appName)
	if err != nil {
		return CFAppInfo{}, err
	}

	user, err := g.cli.Username()
	if err != nil {
		return CFAppInfo{}, err
	}

	org, err := g.cli.GetCurrentOrg()
	if err != nil {
		return CFAppInfo{}, err
	}

	space, err := g.cli.GetCurrentSpace()
	if err != nil {
		return CFAppInfo{}, err
	}

	instances := map[int]CFAppInstance{}
	for id, instance := range app.Instances {
		instances[id] = CFAppInstance{InstanceID: id, Since: instance.Since}
	}

	return CFAppInfo{
		Name:      app.Name,
		Guid:      app.Guid,
		Username:  user,
		Org:       org.Name,
		Space:     space.Name,
		Instances: instances,
	}, nil
}

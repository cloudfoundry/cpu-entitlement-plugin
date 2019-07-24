package metadata

import (
	"code.cloudfoundry.org/cli/plugin"
	models "code.cloudfoundry.org/cli/plugin/models"
)

type CFAppInfo struct {
	App      models.GetAppModel
	Username string
	Org      string
	Space    string
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

	return CFAppInfo{
		App:      app,
		Username: user,
		Org:      org.Name,
		Space:    space.Name,
	}, nil
}

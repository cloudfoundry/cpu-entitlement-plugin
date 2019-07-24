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

func GetCFAppInfo(cli plugin.CliConnection, appName string) (CFAppInfo, error) {
	app, err := cli.GetApp(appName)
	if err != nil {
		return CFAppInfo{}, err
	}

	user, err := cli.Username()
	if err != nil {
		return CFAppInfo{}, err
	}

	org, err := cli.GetCurrentOrg()
	if err != nil {
		return CFAppInfo{}, err
	}

	space, err := cli.GetCurrentSpace()
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

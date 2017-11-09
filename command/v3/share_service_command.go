package v3

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . ShareServiceActor

type ShareServiceActor interface {
	ShareServiceInstanceByOrganizationAndSpaceName(serviceInstanceName string, orgGUID string, spaceName string) (v3action.Warnings, error)
}

type ShareServiceCommand struct {
	RequiredArgs flag.ServiceInstance `positional-args:"yes"`
	//TODO flag.Space does not capture the command line value
	SpaceName       string      `short:"s" description:"Space to share the service instance into"`
	usage           interface{} `usage:"cf share-service SERVICE_INSTANCE -s OTHER_SPACE [-o OTHER_ORG]"`
	relatedCommands interface{} `related_commands:""`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       ShareServiceActor
}

func (cmd *ShareServiceCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		if v3Err, ok := err.(ccerror.V3UnexpectedResponseError); ok && v3Err.ResponseCode == http.StatusNotFound {
			return translatableerror.MinimumAPIVersionNotMetError{MinimumVersion: ccversion.MinVersionRunTaskV3}
		}
		return err
	}
	cmd.Actor = v3action.NewActor(client, config, nil, nil)

	return nil
}

func (cmd ShareServiceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, _ := cmd.Config.CurrentUser()
	// if err != nil {
	// 	return err
	// }

	cmd.UI.DisplayTextWithFlavor("Sharing service instance {{.ServiceInstanceName}} into org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"ServiceInstanceName": cmd.RequiredArgs.ServiceInstance,
		"OrgName":             cmd.Config.TargetedOrganization().Name,
		"SpaceName":           cmd.SpaceName,
		"Username":            user.Name,
	})

	fmt.Println(cmd.SpaceName)

	warnings, err := cmd.Actor.ShareServiceInstanceByOrganizationAndSpaceName(cmd.RequiredArgs.ServiceInstance, cmd.Config.TargetedOrganization().GUID, cmd.SpaceName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

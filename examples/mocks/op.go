package mocks

import (
	"context"

	"github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
)

func Deploy(client alien4cloud.Client) error {
	ctx := context.Background()
	appID, err := client.ApplicationService().CreateAppli(ctx, "myapp", "mytemplate")
	if err != nil {
		return err
	}

	envID, err := client.ApplicationService().GetEnvironmentIDbyName(ctx, appID, alien4cloud.DefaultEnvironmentName)
	if err != nil {
		return err
	}

	return client.DeploymentService().DeployApplication(ctx, appID, envID, "location")
}

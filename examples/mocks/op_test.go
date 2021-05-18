package mocks

import (
	"errors"
	"os"
	"testing"

	"github.com/alien4cloud/alien4cloud-go-client/v3/a4cmocks"
	"github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	"github.com/golang/mock/gomock"
)

func TestDeploy(t *testing.T) {

	tests := []struct {
		name          string
		mocksSetup    func(t *testing.T, mockCtrl *gomock.Controller) alien4cloud.Client
		wantErr       bool
		mocksFailDemo bool
	}{
		{"AllOK", func(t *testing.T, ctrl *gomock.Controller) alien4cloud.Client {
			appServiceMock := a4cmocks.NewMockApplicationService(ctrl)
			depServiceMock := a4cmocks.NewMockDeploymentService(ctrl)
			clientMock := a4cmocks.NewMockClient(ctrl)

			clientMock.EXPECT().ApplicationService().Return(appServiceMock).AnyTimes()
			clientMock.EXPECT().DeploymentService().Return(depServiceMock).AnyTimes()

			appServiceMock.EXPECT().CreateAppli(gomock.Any(), gomock.Any(), gomock.Any()).Return("appID", nil).Times(1)
			appServiceMock.EXPECT().GetEnvironmentIDbyName(gomock.Any(), gomock.Any(), gomock.Any()).Return("envID", nil).Times(1)

			depServiceMock.EXPECT().DeployApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

			return clientMock

		}, false, false},
		{"AppCreationFails", func(t *testing.T, ctrl *gomock.Controller) alien4cloud.Client {
			appServiceMock := a4cmocks.NewMockApplicationService(ctrl)
			depServiceMock := a4cmocks.NewMockDeploymentService(ctrl)
			clientMock := a4cmocks.NewMockClient(ctrl)

			clientMock.EXPECT().ApplicationService().Return(appServiceMock).AnyTimes()
			clientMock.EXPECT().DeploymentService().Return(depServiceMock).AnyTimes()

			appServiceMock.EXPECT().CreateAppli(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("this should fail")).Times(1)

			// Those should not be called
			appServiceMock.EXPECT().GetEnvironmentIDbyName(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			depServiceMock.EXPECT().DeployApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			return clientMock

		}, true, false},
		{"ExpectationsFails", func(t *testing.T, ctrl *gomock.Controller) alien4cloud.Client {
			appServiceMock := a4cmocks.NewMockApplicationService(ctrl)
			depServiceMock := a4cmocks.NewMockDeploymentService(ctrl)
			clientMock := a4cmocks.NewMockClient(ctrl)

			clientMock.EXPECT().ApplicationService().Return(appServiceMock).AnyTimes()
			clientMock.EXPECT().DeploymentService().Return(depServiceMock).AnyTimes()

			// Let say we expect createAppli to be called twice (which is a non-sense)
			appServiceMock.EXPECT().CreateAppli(gomock.Any(), gomock.Any(), gomock.Any()).Return("appID", nil).Times(2)

			// Those should not be called
			appServiceMock.EXPECT().GetEnvironmentIDbyName(gomock.Any(), gomock.Any(), gomock.Any()).Return("envID", nil).Times(1)
			depServiceMock.EXPECT().DeployApplication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

			return clientMock

		}, false, true},
	}
	_, allowMocksFail := os.LookupEnv("MOCKS_FAILS")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mocksFailDemo && !allowMocksFail {
				t.Skip("Skipping mock failure demo")
			}

			mockCtrl := gomock.NewController(t)

			client := tt.mocksSetup(t, mockCtrl)

			if err := Deploy(client); (err != nil) != tt.wantErr {
				t.Errorf("Deploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

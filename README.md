# alien4cloud-go-client

[![PkgGoDev](https://pkg.go.dev/badge/github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud)](https://pkg.go.dev/github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=alien4cloud_alien4cloud-go-client&metric=alert_status)](https://sonarcloud.io/dashboard?id=alien4cloud_alien4cloud-go-client) [![Go Report Card](https://goreportcard.com/badge/github.com/alien4cloud/alien4cloud-go-client)](https://goreportcard.com/report/github.com/alien4cloud/alien4cloud-go-client) [![license](https://img.shields.io/github/license/alien4cloud/alien4cloud-go-client.svg)](https://github.com/alien4cloud/alien4cloud-go-client/blob/master/LICENSE) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

Go client for [Alien4Cloud](https://github.com/alien4cloud/alien4cloud) REST API.

See examples describing how to:

* Catalog Management:
  * [upload a CSAR](examples/upload-csar/README.md)
  * [get available inputs of a topology template](examples/get-input-parameters/README.md)
* Application Management:
  * [create and deploy an application](examples/create-deploy-app/README.md)
  * [deploy an already created application](examples/deploy-app/README.md)
  * [get information on a given application using its ID](examples/get-application-by-id/README.md)
  * [set application inputs](examples/set-input-parameters/README.md)
  * [get inputs used for a given deployment](examples/get-deployment-input-parameters/README.md)
  * [run a workflow](examples/run-workflow/README.md)
  * [get status for a workflow](examples/get-workflow-status/README.md)
  * [undeploy and delete an application](examples/undeploy-delete-app/README.md)
* Users/Groups Management:
  * [create a user](examples/create-user/README.md)
  * [get user information](examples/get-user/README.md)
  * [search for users](examples/search-users/README.md)
* Operations for experts:
  * [call arbitrary API endpoint using raw requests](examples/raw-request/README.md)
* Testing
  * [use mocks to test your application](examples/mocks/README.md)

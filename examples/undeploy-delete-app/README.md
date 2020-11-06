# Undeploy an application and delete it

This example shows how the Alien4Cloud go client can be used to run workflows
on a deployed application.

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd examples/undeploy-delete-app
go build -o undeploy.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application to undeploy and detete
* the falg delete if you want the application to be delete after its undeployment

```bash
./undeploy.test -url https://1.2.3.4:8088 \
                -user myuser \
                -password mypasswd \
                -app myapp \
                -delete
```

This will undeploy the application, print undeployment logs,
and finally once the undeployment is done, delete the application.

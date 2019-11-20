# Create an application from a template and deploy it on a location

This example program shows how the Alien4Cloud go client can be used to run workflows
on a deployed application.

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Compile this example on your workstation:

```
cd examples/undeploy-delete-app
go build -o undeploy
```

Now, run this application providing in arguments:
* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application to undeploy and detete

```
./undeploy -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -app myapp
```

This will undeploy the application, print undeployment logs,
and finally once the undeployment is done, delete the application.

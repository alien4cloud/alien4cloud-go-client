# Create an application from a template and deploy it on a location

This example shows how the Alien4Cloud go client can be used to:

* create an application from a template in Alien4Cloud catalog
* optionally, deploy this application on a given location (if no location is specified, the first suited location is selected)
  * while the application is being deployed, display deployment logs
  * once done, display application components variables, if any

## Prerequisites

As a prerequisite before being able to use this example, you should add components
and an Application template to Alien4Cloud catalog.

For example, you could add a sample web application available in the [Ystia Forge](https://github.com/ystia/forge/blob/develop/org/ystia/README.rst)
as described in [Welcome sample section](https://github.com/ystia/forge/blob/develop/org/ystia/README.rst#welcome-sample).

This will add an application template **org.ystia.samples.topologies.welcome_basic** in the Alien4Cloud catalog.
See the [application description](https://github.com/ystia/forge/blob/develop/org/ystia/samples/topologies/welcome_basic/README.rst)
and corresponding [TOSCA file](https://github.com/ystia/forge/blob/develop/org/ystia/samples/topologies/welcome_basic/types.yml)
describing a web application deployed on a Compute Instance attached to a network,
and providing custom workflows **killWebServer**, **stopWebServer**, **startWebServer**.

## Running this example

Build this example:

```bash
cd examples/create-deploy-app
go build -o create.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of a user who has the **Application Manager** role
* the name of the application that will be create
* the application template in ALien4Cloud catalog that will be used
* specify you want to deploy the crrated applicatiob
* optionally, the name of the location where you want to deploy the application
  (by default, the first location suited for the deployment will be selected)

```bash
./create.test -url https://1.2.3.4:8088 \
              -user myuser \
              -password mypasswd \
              -app myapp \
              -template org.ystia.samples.topologies.welcome_basic \
              -deploy
```

This will create the application, deploy it on a location, print deployment logs,
and finally once the deployment is done, print output variables if any.
Here in the case of the Welcome sample, it will output the URL of the deployed Web application.

## What's next

To be able to run workflows on this deployed application, see [Run workflows](../run-workflow/README.md) example.

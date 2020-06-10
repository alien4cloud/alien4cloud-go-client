# Run a workflow

This example shows how the Alien4Cloud go client can be used to run a workflow on
a deployed application.

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd examples/run-workflow
go build -o run.test
```

Now, run this example providing in arguments:
* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the workflow to run.

For example, to re-use the Forge Web application sample deployed in [Create an deploy an application](../create-deploy-app/README.md) example,
which is providing workflows **killWebServer**, **stopWebServer**, **startWebServer**, this would give :

```bash
./run.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -app myapp \
           -workflow stopWebServer
```

You can also specify the option `-events` to see workflow events received during the workflow execution.
Else (default case), workflow logs will be printed.

## What's next

To undeploy and delete this application, see [Undeploy and Delete application](../undeploy-delete-app/README.md) example.

# Run a workflow

This example shows how the Alien4Cloud go client can be used to get the status of
a workflow execution.

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd examples/get-workflow-status
go build -o wfstatus.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the name of the workflow for which to get the status.

For example, to re-use the Forge Web application sample deployed in [Create an deploy an application](../create-deploy-app/README.md) example,
which is providing workflows **Install**, **killWebServer**, **stopWebServer**, **startWebServer**, this would give :

```bash
./wfstatus.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -app myapp \
           -workflow install
```

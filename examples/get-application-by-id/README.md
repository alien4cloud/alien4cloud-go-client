# Run a workflow

This example shows how the Alien4Cloud go client can be used to get an application from its ID.

## Prerequisites

An application has been created as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd get-application-by-id
go build -o getappbyid.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the ID of the application

For example, to re-use the Forge Web application sample deployed in [Create an deploy an application](../create-deploy-app/README.md) example,
which is providing workflows **killWebServer**, **stopWebServer**, **startWebServer**, this would give :

```bash
./getappbyid.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -id Myapp
```

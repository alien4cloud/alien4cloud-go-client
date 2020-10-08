# Get input properties values used for a given deployment

This example shows how the Alien4Cloud go client can be used to get the
values of input properties used for a given deployment

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd examples/get-deployment-input-parameters/
go build -o get-deployment-inputs.test
```

Now, to get a deployment input property values, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application

For example :

```bash
./get-deployment-inputs.test -url https://1.2.3.4:8088 \
                             -user myuser \
                             -password mypasswd \
                             -app MyApp
```

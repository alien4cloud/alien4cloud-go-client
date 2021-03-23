# Get an input paramter type

This example shows how the Alien4Cloud go client can be used to get details on an
input paramter type

## Prerequisites

An application has been created as described in [Create an deploy an application](../create-deploy-app/README.md) example, but not yet deployed.

## Running this example

Build this example:

```bash
cd examples/get-input-parameter-type/
go build -o getinput-type.test
```

Now, to set an application input property, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the input property name

For example :

```bash
./getinput-type.test -url https://1.2.3.4:8088 \
                -user myuser \
                -password mypasswd \
                -app MyApp \
                -property myprop \
```

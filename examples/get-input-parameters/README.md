# Get input properties

This example shows how the Alien4Cloud go client can be used to get the input
properties of a template and in which components these properties are used

## Prerequisites

An application template has been uplaoded in Alien4Cloud catalog.

## Running this example

Build this example:

```bash
cd examples/get-input-parameters/
go build -o getinputs.test
```

Now, to set an application input property, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the input property name
* the input property value.

For example :

```bash
./getinputs.test -url https://1.2.3.4:8088 \
                -user myuser \
                -password mypasswd \
                -template name:version
```

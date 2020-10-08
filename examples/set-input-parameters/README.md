# Set input properties and input artifacts

This example shows how the Alien4Cloud go client can be used to set an application
input property or input artifact.

## Prerequisites

An application has been deployed as described in [Create an deploy an application](../create-deploy-app/README.md) example.

## Running this example

Build this example:

```bash
cd examples/set-input-parameters/
go build -o setinput.test
```

Now, to set an application input property, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the input property name
* the input property value.

For example :

```bash
./setinput.test -url https://1.2.3.4:8088 \
                -user myuser \
                -password mypasswd \
                -app MyApp \
                -property myprop \
                -value myval
```

To set an application input artifact, run this example, providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* the name of the application
* the input artifact name
* the path to input artifact file.

For example :

```bash
./setinput.test -url https://1.2.3.4:8088 \
                -user myuser \
                -password mypasswd \
                -app MyApp \
                -artifact myartifact \
                -file /home/user/myfile.txt
```

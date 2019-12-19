# Run a workflow

This example shows how the Alien4Cloud go client can be used to upload a CSAR to Alien4Cloud catalog.

## Running this example

Build this example:

```bash
cd examples/upload-csar
go build -o run.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of the user who has deployed the application
* csar path

```bash
./run.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -csar /path/to/csar.zip
```

## What's next

To deploy an application, see [create and deploy an application](../create-deploy-app/README.md) example.

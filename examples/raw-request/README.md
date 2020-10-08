# Sending Raw Requests to Alien4cloud

This example shows how to use the Alien4Cloud go client to call arbitrary REST API endpoints.
This allows you to call endpoints not directly supported by this client.
Typically this allows to call API endpoints coming from plugins that could not be known in advance.

## Running this example

Build this example:

```bash
cd raw-request
go build -o raw.test
```

Now, run this example providing in arguments:

* the Alien4Cloud URL
* credentials of a user having the administrator

For example:

```bash
./raw.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd
```

This will return information about the current user's status and it's roles.

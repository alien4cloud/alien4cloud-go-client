# Gets a user

This example shows how the Alien4Cloud go client can be used to get a user parameters.

## Running this example

Build this example:

```bash
cd get-user
go build -o getuser.test
```

Now, run this example providing in arguments:
* the Alien4Cloud URL
* credentials of a user having the administrator
* user name for which to get parameters

For example:

```bash
./getuser.test -url https://1.2.3.4:8088 \
               -user adminuser \
               -password mypasswd \
               -username myuser
```

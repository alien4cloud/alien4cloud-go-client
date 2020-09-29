# Create a user

This example shows how the Alien4Cloud go client can be used to create a user.

## Running this example

Build this example:

```bash
cd create-user
go build -o create.test
```

Now, run this example providing in arguments:
* the Alien4Cloud URL
* credentials of a user having the administrator
* user properties

For example:

```bash
./create.test -url https://1.2.3.4:8088 \
              -user adminuser \
              -password mypasswd \
              -username mynewuser \
              -userpassword mynewuserpassword \
              -firstname John \
              -lastname Doe \
              -email "john.doe@acme.com" \
              -role APPLICATIONS_MANAGER -role ARCHITECT
```

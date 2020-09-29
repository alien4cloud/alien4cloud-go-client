# Search for users

This example shows how the Alien4Cloud go client can be used to search for users.

## Running this example

Build this example:

```bash
cd search-users
go build -o search.test
```

Now, run this example providing in arguments:
* the Alien4Cloud URL
* credentials of a user having the administrator
* search properties (optional):
  * from: start index of user to return
  * size: maximum number of users to return
  * query: string to search

For example:

```bash
./search.test -url https://1.2.3.4:8088 \
           -user myuser \
           -password mypasswd \
           -size 10
```
This will return the first 10 users, as well as the total number of users.
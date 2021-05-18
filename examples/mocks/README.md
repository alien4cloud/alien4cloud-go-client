# Mocking Alien4Cloud Client within your tests

This library provides out of the box mocks stubs that suitable to be used in go tests.

There is a lot of mocking frameworks in Go but we selected the [GoMock](https://github.com/golang/mock) framework for its nice and powerful expectations API.

Mocks generation is directly performed by this library and you do not need to take care of this to use mocks in your tests.

## Package description

This package contains a single function [`Deploy(alien4cloud.Client)`](op.go#L9). This is a dummy function that mimic an application
deployment. This is used to demonstrate how to use mocks during testing.
This function is tested by [`TestDeploy(*testing.T)`](op_test.go#L13) that illustrate how you could configure your mocks for testing an
application deployment.

This is not intended to be a tutorial on GoMock features. Please refer to its documentation for more details.

## Running this example

Run this example:

```bash
$ cd examples/mocks
$ go test -v .
=== RUN   TestDeploy
=== RUN   TestDeploy/AllOK
=== RUN   TestDeploy/AppCreationFails
=== RUN   TestDeploy/ExpectationsFails
    op_test.go:77: Skipping mock failure demo
--- PASS: TestDeploy (0.00s)
    --- PASS: TestDeploy/AllOK (0.00s)
    --- PASS: TestDeploy/AppCreationFails (0.00s)
    --- SKIP: TestDeploy/ExpectationsFails (0.00s)
PASS
ok      github.com/alien4cloud/alien4cloud-go-client/v3/examples/mocks  0.003s
```

As you can see there is a test that assert a normal behavior (`TestDeploy/AllOK`), a test that test an application
creation error path (`TestDeploy/AppCreationFails`) and a skipped test on expectations failures.

The skipped one will always fail as it demonstrate what happens when an expected call to a mock function is missing.
You can activate this test using an env variable.

```bash
$ MOCKS_FAILS=1 go test -v .
--- FAIL: TestDeploy (0.00s)
    --- FAIL: TestDeploy/ExpectationsFails (0.00s)
        controller.go:137: missing call(s) to *a4cmocks.MockApplicationService.CreateAppli(is anything, is anything, is anything) /home/a454241/workspaces/ystia/alien4cloud-go-client/examples/mocks/op_test.go:63
        controller.go:137: aborting test due to missing call(s)
FAIL
FAIL    github.com/alien4cloud/alien4cloud-go-client/v3/examples/mocks  0.003s
FAIL
```

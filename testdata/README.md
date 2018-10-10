# Testing

The `testdata` dir contains function that will create a *real* datastore
against which the datastore tests can run.

These are kept here because the `datastore` pkg is use to access databases
that already exist. The creation and destruction of databases is to set
 up a test environment and should not be part of the tests themselves.




## Refs

* https://medium.com/kongkow-it-medan/parallel-database-integration-test-on-go-application-8706b150ee2e
* https://testing.googleblog.com/2015/04/just-say-no-to-more-end-to-end-tests.html






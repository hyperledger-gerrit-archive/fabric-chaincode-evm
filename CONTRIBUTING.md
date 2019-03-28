# Contributions Welcome!

This repository is part of the Fabric project.
Please consult [Fabric's CONTRIBUTING documentation](http://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING.html) for information on how to contribute to this repository.

## Developer Advice

New code being added should be tested. All tests are automatically run after
changes are pushed to gerrit.

All tests must use [ginkgo](github.com/onsi/ginkgo/ginkgo), which will be built
and installed when invoking the makefile.

Ginkgo is not only a testing library, but also a convenient tool to use during
development.

As an example, an invocation of `ginkgo watch -skipPackages integration` can be
started on the repository as a whole. When done at the start of development,
this runs all the unit tests, and then re-runs them when any changes to the code
are detected.

With the addition of the `-notify` flag, testing can be left on in a terminal in
the background, and you can remain in the editor at all times, only checking on
the output when you see a failure.

The final command ends up looking like:

    ginkgo watch -notify -randomizeAllSpecs -requireSuite -race -cover -skipPackage integration -r`

The race detector should always be on during testing to ensure that races are
prevented. The race detector is enabled with the `-race` flag. Check out the
[Official Go Blog](https://blog.golang.org/race-detector) for more information
on the race detector.

Code coverage can be a useful tool during development. The `-cover` flag enables
measurement of code coverage. Check out the official go blog entry, [The Cover
Story](https://blog.golang.org/cover), for details on available measurement
types. Code coverage measurements can be used to verify that specific segments
of code are run during a test. This project (fabric-chaincode-evm) does not
target any specific code coverage percentage.

For example, if I wanted to measure the coverage of the `EstimateGas` function
provided by `ethservice`, I could target those tests by adding the focus
flag. Then I would only measure what code is run during those specific tests.

    $ ginkgo -cover -focus 'EstimateGas' ./fab3
    Running Suite: Fab3 Main Suite
    ==============================
    Random Seed: 1554233945
    Will run 1 of 62 specs

    SSSSSSSSSSSSSSSSSSSSSSSS•SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
    Ran 1 of 62 Specs in 0.001 seconds
    SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 61 Skipped
    PASS
    coverage: 1.3% of statements

    Ginkgo ran 1 suite in 6.386756543s
    Test Suite Passed

A development focused make target has been packaged as `dev-test` and can be
invoked with `make dev-test`.

Higher level integration tests should also be written after completing lower
level unit-testing, and are available to run with `make integration-test`.

<a rel="license" href="http://creativecommons.org/licenses/by/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>

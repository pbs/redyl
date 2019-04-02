# How to set up a development environment

* Clone this repository into a Go [workspace](https://golang.org/doc/code.html#Organization):

```
git clone git@github.com:pbs/redyl.git $GOPATH/src/github.com/pbs/redyl
```

* Using [`dep`](https://golang.github.io/dep/), ensure local dependencies are installed

```
cd $GOPATH/src/github.com/pbs/redyl
dep ensure
```

# How to run tests

* Run the test script:

```
./scripts/test.sh
```

# How to release

* Make sure the version number in [`internal/redyl/version/version.go`](/internal/redyl/version/version.go) matches the version you're releasing.
* Run the release script to cross-compile binaries for distribution:
```
./scripts/release.sh
```
* create a new release in github: https://github.com/pbs/redyl/releases/new
* add the binaries created locally in your `./bin` directory
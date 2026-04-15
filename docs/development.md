# How to set up a development environment

* Clone this repository:

```
git clone git@github.com:pbs/redyl.git
cd redyl
```

* Install dependencies:

```
go mod download
```

# How to build

```
./scripts/build.sh
```

The binary will be at `./bin/redyl`.

# How to run tests

```
./scripts/test.sh
```

# How to release

* Bump the version number in [`internal/redyl/version/version.go`](/internal/redyl/version/version.go).
* Commit and push the version bump.
* Run the release script to cross-compile binaries for distribution:
```
./scripts/release.sh
```
* Create a new release in GitHub: https://github.com/pbs/redyl/releases/new
* Tag the release matching the version (e.g., `v1.1.0`).
* Upload the binaries from your `./bin` directory.

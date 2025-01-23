# Contributing

## Building
To build the application, run the following command:
```bash
go build
```

This will create an executable file named `xmatters`.

## Installation
To install the application, run the following command:
```bash
go install
```

## Releasing
To release the application, run the following command (change the version number as needed):
```bash
VERSION=0.4.0
git tag -a v$VERSION -m "Release v$VERSION"
git push origin v$VERSION
goreleaser release --clean
```

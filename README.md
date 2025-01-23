# XMatters On Call Sync
A Go application that syncs on-call schedules from xMatters to a Google Calendar. It is designed to be run as a cron job.

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

# Add the following to your crontab
0 * * * * xmatters
```

## Releasing
To release the application, run the following command (change the version number as needed):
```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
goreleaser release --clean
```

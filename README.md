# XMatters On Call Sync
A Go application that syncs on-call schedules from xMatters to a Google Calendar. It is designed to be run as a cron job.

## Installation
See the `Install` section in the latest release

## Usage
First, message `@thomas.hart` on Slack your Google Calendar email address.

<sub>I cannot publish the Google Application, so I need to add you as a Test User before you can grant access.</sub>

Running the command, it will prompt you for your xMatters and Google Calendar credentials. It will then sync the on-call schedules from xMatters to the Google Calendar.
```shell
xmatters-sync
```

## Cron
To run the application as a cron job, add the following line to your crontab. This will run the application every 15 minutes.

To pick a different interval, see [crontab.guru](https://crontab.guru/).
```shell
*/15 * * * * xmatters-sync
```
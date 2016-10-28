![zenbot logo](/logo.png)  

[![Build Status](https://semaphoreci.com/api/v1/kamaln7/zenbot/branches/master/badge.svg)](https://semaphoreci.com/kamaln7/zenbot)

zenbot is a Slack bot that enforces users to have silent periods in which they are reprimanded for offending. Currently, the only available action is giving negative [karmabot](https://github.com/kamaln7/karmabot) karma points.

## Syntax

`./zen <duration e.g. 1h30m> [reason - optional]`

## Installation

### Build from Source

1. clone the repo:
    1. `git clone -b v1.0.0 https://github.com/kamaln7/zenbot.git`
2. run `go get` and then `go build` inside `cmd/zenbot`
    1. `cd zenbot`
    2. `go get`
    1. `cd cmd/zenbot`
    3. `go build`

### Download a Pre-built Release

1. head to [the repo's releases page](https://github.com/kamaln7/zenbot/releases) and download the appropriate latest release's binary for your system

## Usage

1. add a **Slack Bot** integration: `https://team.slack.com/apps/A0F7YS25R-bots` 
2. invite `zenbot` to any existing channels and all future channels (this is a limitation of Slack's bot API, unfortunately)
3. run `zenbot`. the following options are supported:

| option                  | required? | description                              | default        |
| ----------------------- | --------- | ---------------------------------------- | -------------- |
| `-token string`         | **yes**   | slack RTM token                          |                |
| `-debug=bool`           | no        | set debug mode                           | `false`        |

In addition, see the table below for the options related to the web UI.

**example:** `./karmabot -token xoxb-abcdefg`

I recommend passing zenbot's logs through [humanlog](https://github.com/aybabtme/humanlog). humanlog will format and color the JSON output as nice easy-to-read text.

## License

see [./LICENSE](/LICENSE)

![zenbot logo](/logo.png)  

[![Build Status](https://semaphoreci.com/api/v1/kamaln7/zenbot/branches/master/badge.svg)](https://semaphoreci.com/kamaln7/zenbot)

zenbot is a Slack bot that enforces "zen periods" for users, during which they are reprimanded for using Slack. Currently, the only available action is giving negative [karmabot](https://github.com/kamaln7/karmabot) karma points.

## Syntax

- add a zen: `./zen <duration e.g. 1h30m> [reason - optional]`
- cancel a zen: `./zen cancel [reason]`
- cancel all zens: `./zen cancel`

## Installation

### Build from Source

1. clone the repo:
    1. `git clone -b v1.1.1 https://github.com/kamaln7/zenbot.git`
2. run `go get` and then `go build` inside `cmd/zenbot`
    1. `cd zenbot`
    2. `go get`
    3. `cd cmd/zenbot`
    4. `go build`

### Download a Pre-built Release

1. head to [the repo's releases page](https://github.com/kamaln7/zenbot/releases) and download the appropriate latest release's binary for your system

## Usage

1. add a **Slack Bot** integration: `https://team.slack.com/apps/A0F7YS25R-bots`. an avatar is available [here](/avatar.png).
2. invite `zenbot` to any existing channels and all future channels (this is a limitation of Slack's bot API, unfortunately)
3. run `zenbot`. the following options are supported:

| option                   | required? | description                              | default |
| ------------------------ | --------- | ---------------------------------------- | ------- |
| `-token string`          | **yes**   | slack RTM token                          |         |
| `-debug=bool`            | no        | set debug mode                           | `false` |
| `-timeout string`        | no        | timeout between actions (karma downvotes) | `10s`   |
| `-whitelist.chan string` | no        | **may be specified multiple times** set a list of channel that zenbot may be used in. | `[]`    |

**example:** `./zenbot -token xoxb-abcdefg -whitelist.chan zen -whitelist.chan zenbot-test `

It is recommended to pass zenbot's logs through [humanlog](https://github.com/aybabtme/humanlog). humanlog will format and color the JSON output as nice easy-to-read text.

## Other

A systemd service unit file is available [here](/zenbot.service). Edit the service file and replace `xoxb-abcdefg` with your Slack RTM token. Install it as usual.

## License

see [./LICENSE](/LICENSE)

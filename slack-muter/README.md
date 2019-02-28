# Slack Muter

## Why?

In a remote team, the amount of emails and Slack messages can be overwhelming. The only solution is to limit your notifications
to only those you intend to respond to. When I'm on support, I intend to respond to all customer shared channel requests,
but when I'm not I only intend to respond to `#support-discussion` messages. 

But manually muting or unmuting all the shared channels is a pain. Hence this script. :)

## What? 

This script will take a slack API token and automatically mute or unmute your user account from all shared channels.

*Note*: This script requires python 3.6+

## Script Params

| Param               | Description                                                                                   | Required |
|---------------------|-----------------------------------------------------------------------------------------------|----------|
| `--slack-token`     | The Slack API token to use                                                                    | yes      |
| `-m` or `--mute`    | If present, all shared channels will be muted                                                 | xor      |
| `u` or `--unmute`   | If present, all shared channels will be unmuted                                               | xor      |
| `-d` or `--dry-run` | The presence of this flag will not actually mute the channels but will print the log messages | no       |


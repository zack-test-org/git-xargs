# Slack Muter

## Motivation

In a remote team, the amount of emails and Slack messages can be overwhelming. Arguably the best solution is to limit your notifications
to only those you intend to respond to. When I'm on support, I intend to respond to all customer shared channel requests,
but when I'm not I only intend to respond to `#support-discussion` messages. 

That means I want to mute all shared channels while on support, but unmute them when I'm not on support. Doing this manually
is a pain, so this script automates it.

## How It Works

Slack's API does not allow you to individually mute or unmute channels. Rather, you can only set a list of all channels
that are currently muted. As a result, naively muting only the shared channels would have the effect of _unmuting_ all 
your currently muted channels. For that reason, this script will check which channels you have currently muted and
be sure to set them again if necessary. 

## Requirements

*Note*: This script requires python 3.6+

## Script Params

| Param               | Description                                                                                   | Required |
|---------------------|-----------------------------------------------------------------------------------------------|----------|
| `--slack-token`     | The Slack API token to use                                                                    | yes      |
| `-l` or `--list`    | List all currently muted channels                                                 | xor      |
| `-m` or `--mute`    | Mute all shared channels                                                 | xor      |
| `u` or `--unmute`   | Unmute all shared channels                                               | xor      |

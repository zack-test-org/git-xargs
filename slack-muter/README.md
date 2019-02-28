# Slack Muter

## Motivation

In a remote team, the amount of emails and Slack messages can be overwhelming. Arguably the best solution is to limit 
the Slack channel you follow to only those you intend to respond to. When I'm on support, I intend to respond to all 
customer shared channel requests, but when I'm not, I only intend to respond to `#support-discussion` messages. 

That means I want to [mute](https://get.slack.help/hc/en-us/articles/204411433-Mute-a-channel) all shared channels while
not on support, but unmute them when I'm not on support. Doing this for our 60+ shared channels manually is painful to the
point of being prohibitive, so this script automates it.

## How It Works

Slack's API does not allow you to individually mute or unmute channels. Rather, you can only set a list of all channels
that are currently muted. As a result, naively muting only the shared channels would have the effect of _unmuting_ all 
your non-shared, previously muted channels. For that reason, this script will check which channels you have currently 
muted and be sure to set them again if necessary. 

## Usage

```
python slack_muter.py --slack-token XXX --list
python slack_muter.py --slack-token XXX --mute
python slack_muter.py --slack-token XXX --unmute
```

| Param               | Description                          | Required |
|---------------------|--------------------------------------|----------|
| `--slack-token`     | The Slack API token to use           | yes      |
| `-l` or `--list`    | List all currently muted channels    | xor      |
| `-m` or `--mute`    | Mute all shared channels             | xor      |
| `u` or `--unmute`   | Unmute all shared channels           | xor      |

## Requirements

*Note*: This script requires python 3.6+. The `python` binary on your local machine may be named `python3`.

## TODO

- Running `python slack_muter.py --slack-token XXX --mute` is effectively idempotent, but has a small bug where it keeps
  redundantly adding the same list of channels to mute. This means that running this command repeatedly adds an additional
  copy of each channel in Slack's settings. You can work around this by just running `python slack_muter.py --unmute`
  before running `--mute` a second time. 
# Slack Exporter

## Motivation

We have a [free Slack workspace for the Gruntwork community](https://gruntwork-community.slack.com), however all free 
Slack plans are limited to searching the previous 10,000 messages. I thought that only the most recent 10,000 messages
were exportable, in which case we would need a scheduled job to export messages to preserve history. But upon writing
this README, I realized I was wrong: Even free Slack lets you export the entire workspace history as JSON at any time.

Unfortunately, that makes this code interesting but useless today. It may be useful in the future if we decide to 
automatically publish Slack conversations in a searchable archive for users. 

## How It Works

The `slack_exporter.py` script queries the Slack APIs to get a list of users, channels, and for each channel, all messages.
It then outputs this information as a JSON string.

## Usage

```
export SLACK_TOKEN=xxx
python slack_exporter.py
```

| Param               | Description                          | Required |
|---------------------|--------------------------------------|----------|
| `--slack-token`     | The Slack API token to use, but you should use the env var `SLACK_TOKEN` instead | no      |

## Requirements

*Note*: This script requires python 3.6+. The `python` binary on your local machine may be named `python3`.
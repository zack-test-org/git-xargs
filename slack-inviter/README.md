# Slack Inviter

## Motivation

We have a [free Slack workspace for the Gruntwork community](https://gruntwork-community.slack.com) that can only be 
accessed by invitation. Customers are automatically invited when we run the [Add New User Zapier Zap](
https://zapier.com/app/editor/47068670/overview), however there may be many existing customers who are already registered
with Gruntwork and _only_ need to be added to Slack. In such a case, this script offers a way to do that via the CLI.  

## How It Works

The `slack_inviter.py` script calls the `users.admin.invite` method in the Slack API to invite a user, which happens to
be [undocumented](https://stackoverflow.com/a/36114710/2308858). It requires [getting a Legacy Slack Token](
https://api.slack.com/custom-integrations/legacy-tokens).

The script queries Slack for all public, non-archived channels and adds the user to those channels by default.

## Usage

```
export SLACK_TOKEN=xxx
python slack_inviter.py --email han.solo@acme.com 
```

| Param               | Description                          | Required |
|---------------------|--------------------------------------|----------|
| `--slack-token`     | The Slack API token to use, but you should use the env var `SLACK_TOKEN` instead | no      |
| `--email`           | The email of the user to be invited                                              | yes      |

## Requirements

*Note*: This script requires python 3.6+. The `python` binary on your local machine may be named `python3`.
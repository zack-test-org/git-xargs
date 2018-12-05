# Slack Auto Joiner

This script will take a slack API token and a list of user's display names and then automatically add those users
to all of Gruntwork's shared channels.

The main use case for this script is adding new Grunts to all of the shared channels that we have with our customers.
Once a Grunt is part of the team, we add all grunts to all shared channels upon creation so there's no real need to run
this script for users who are already part of the organization.

*Note*: This script requires python 3.6+

## Script Params

| Param               | Description                                                                                   | Required |
|---------------------|-----------------------------------------------------------------------------------------------|----------|
| `--slack-token`     | The Slack API token to use                                                                    | yes      |
| `u` or `--users`    | A space separated list of users display names                                                 | yes      |
| `-d` or `--dry-run` | The presence of this flag will not actually invite the users but will print the log messages. | no       |


#!/bin/bash

# The Github account where the discussions repository lives, e.g. gruntwork-io
export GITHUB_ACCOUNT=

# How much time in seconds between executing Github GraphQL queries for updates
export GITHUB_QUERY_SLEEP_INTERVAL=300

# The Github repository where the discussions reside
export GITHUB_REPO=

# The Github token used to make queries
export GITHUB_TOKEN=

# The bot collects members of this channel to see who is allowed to export conversations to Github
export SLACK_ADMIN_CHANNEL=

# When a question is answered, this emoji is used
export SLACK_ANSWERED_EMOJI=white_check_mark

# The slack domain is used to link to Slack archives, e.g. gruntwork-io.slack.com
export SLACK_DOMAIN=

# When the bot is sending the discussion to Github discussions, this emoji is used
export SLACK_LOOKING_EMOJI=eyes

# The bot looks for this emoji (activated by someone in the admin channel) to trigger an export to Github discussions
export SLACK_SEND_EMOJI=github

# The Slack token the bot uses
export SLACK_TOKEN=

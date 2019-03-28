import re
import requests
import logging
import argparse
import json
import os

logging.basicConfig()
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def get_all_users(slack_token):
    """
    Get a list of all Slack users based on https://api.slack.com/methods/users.list

    This function makes multiple calls to Slack API to account for paginated results.

    Args:
        slack_token: Authentication token bearing required scopes

    Returns:
        A list of user dictionaries
    """

    url = "https://slack.com/api/users.list"

    params = {
        'token': slack_token,
    }

    next_cursor = 'first-pass'
    users = []

    while next_cursor:
        response = requests.get(url=url, params=params).json()

        if not response['ok']:
            raise Exception(f'List of users could not be returned. Response from Slack: {response}')

        for user in response['members']:
            users.append(user)

        # For pagination to work properly, we must make the next request with the cursor returned from the previous
        # request, if applicable. See https://api.slack.com/methods/users.list
        next_cursor = ''

        response_metadata = response.get('response_metadata')

        if response_metadata:
            next_cursor = response_metadata.get('next_cursor')

            if next_cursor:
                params['next_cursor'] = next_cursor

    return users


def get_all_channels(slack_token):
    """
    Get a list of all channels in the Slack account based on https://api.slack.com/methods/conversations.list

    This function makes multiple calls to Slack API to account for paginated results.

    Args:
        slack_token: Authentication token bearing required scopes

    Returns:
        A list of channel dictionaries
    """

    url = "https://slack.com/api/conversations.list"

    first_pass = True
    next_cursor = ""

    channels = []

    while next_cursor != "" or first_pass:
        first_pass = False

        params = {
            'token': slack_token
        }

        if next_cursor != "":
            params.update({'cursor': next_cursor})

        response = requests.get(url=url, params=params).json()

        # Grab the next page token if present in response.
        response_metadata = response['response_metadata']
        next_cursor = response_metadata['next_cursor']

        for channel in response['channels']:
            channels.append(channel)

    return channels


def get_messages(slack_token, channel, count=1000, newest_ts='', oldest_ts=0):
    """
    Return a list of messages for the given Slack channel.

    Based on https://api.slack.com/methods/channels.history

    Args:
        slack_token: Authentication token bearing required scopes
        channel: Channel ID to fetch history for
        count: Number of messages to return, between 1 and 1000
        newest_ts: End of time range of messages to include in results
        oldest_ts: Start of time range of messages to include in results

    Returns:
        A portion of message events (https://api.slack.com/events/message) from the specified public channel.
    """

    url = "https://slack.com/api/channels.history"

    params = {
        'token': slack_token,
        'channel': channel,
        'count': count,
        'inclusive': True,
        'latest': newest_ts,
        'oldest': oldest_ts
    }

    has_more = True
    messages = []

    while has_more:
        response = requests.get(url=url, params=params).json()

        if not response['ok']:
            raise Exception(f'Could not fetch message history for channel. Response from Slack: {response}')

        for message in response['messages']:
            messages.append(message)

        # For pagination to work properly, we must request that the next page of messages start with newest timestamp we
        # received in the previous page of results, per https://api.slack.com/methods/channels.history.
        has_more = response['has_more']
        if has_more:
            newest_timestamp = messages[-1]['ts']
            params['latest'] = newest_timestamp

    return messages


def parse_args():
    """Parse command line args"""

    parser = argparse.ArgumentParser(
        description='This script fetches messages from all Slack channels and outputs them as a single JSON file.')

    parser.add_argument('--slack-token', help='The Slack API token to use, but please use the SLACK_TOKEN env var instead.')
    args = parser.parse_args()

    if not (args.slack_token or 'SLACK_TOKEN' in os.environ):
        parser.error('At least one of --slack-token or the env var SLACK_TOKEN must be set.')

    return args


def main():
    args = parse_args()

    if args.slack_token:
        slack_token = args.slack_token
    else:
        slack_token = os.environ['SLACK_TOKEN']

    output = {}

    print(f'Fetching all users...')
    users = get_all_users(slack_token)
    output['users'] = users

    print(f'Fetching all channels...')
    channels = get_all_channels(slack_token)
    output['channels'] = channels

    output_messages = []
    for channel in channels:
        print('Fetching messages for #{}...'.format(channel['name']))
        messages = get_messages(slack_token, channel['id'], 5)

        output_messages.append({
            'channel': channel['id'],
            'messages': messages
        })

    output['messages'] = output_messages

    output_json = json.dumps(output)

    print(output_json)


if __name__ == "__main__":
    main()

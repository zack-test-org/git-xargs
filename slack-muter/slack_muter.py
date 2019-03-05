import re
import requests
import logging
import argparse
import json

logging.basicConfig()
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def get_current_user_name(slack_token):
    """Get the current Slack user"""

    url = "https://slack.com/api/auth.test"

    params = {
        'token': slack_token,
    }

    response = requests.post(url=url, params=params).json()

    if response['ok']:
        return response['user']

    raise Exception(f'No user was found. Response from Slack: {response}')


def get_all_channels(slack_token):
    """Get a list of all channels in the Slack account

    This function makes multiple calls to Slack API to account for paginated results.
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


def get_shared_channels(slack_token):
    """Get a list of all shared channels in the Slack account

    We defined a shared channel as a Slack channel that is not archived and:
    - is shared with another Slack workspace, or
    - starts with a single "_" (e.g. "_banco-inter")
    """

    all_channels = get_all_channels(slack_token)
    shared_channels = []

    for channel in all_channels:
        is_shared = (bool(channel['is_shared'])
                     | bool(channel['is_ext_shared'])
                     | bool(re.match("^_[A-Za-z]+", channel['name']))) \
                     & (not bool(channel['is_archived']))

        if is_shared:
            shared_channels.append(channel)

    return shared_channels


def get_current_muted_channel_ids(slack_token):
    """Get a list of channel IDs that are currently muted"""

    url = "https://slack.com/api/users.prefs.set"

    params = {
        'token': slack_token,
    }

    response = requests.get(url=url, params=params).json()

    if not response['ok']:
        raise Exception(f'Could not get muted channels. Response from Slack: {response}')

    # We get a string of comma-separated channels back from Slack
    channels_str = response['prefs']['muted_channels']

    channels = channels_str.split(',')

    return channels


def get_list_as(property_name, list):
    """Given a list of dictionaries, return a list of just the given property name of each item in the list"""

    return [item[property_name] for item in list]


def mute_channels(slack_token, channel_ids):
    """Mute each channel in the list of channels

    Note that Slack does not allow muting an individual channel, only setting the list of all muted channels in the
    user's preferences.
    """

    url = "https://slack.com/api/users.prefs.set"

    channels_str = ','.join(channel_ids)

    params = {
        'token': slack_token,
        'prefs': json.dumps({'muted_channels': channels_str})
    }

    response = requests.post(url=url, params=params).json()

    if not response['ok']:
        raise Exception(f'Channel could not be muted. Response from Slack: {response}')


def get_channel_names(slack_token, channel_ids):
    """Given a list of channel IDs, return a list of channel names"""

    all_channels = get_all_channels(slack_token)

    channel_names = [
        channel['name']
        for channel in all_channels
        if channel['id'] in channel_ids
    ]

    channel_names.sort()

    return channel_names


def parse_args():
    """Parse command line args"""

    parser = argparse.ArgumentParser(
        description='This script lists all shared Slack channels and mutes / unmutes them.')

    parser.add_argument('--slack-token', required=True, help='The Slack API token to use')
    parser.add_argument('-l', '--list', action='store_true', help='List all currently muted channels')
    parser.add_argument('-m', '--mute', action='store_true', help='Mute all shared channels')
    parser.add_argument('-u', '--unmute', action='store_true', help='Unmute all shared channels')

    args = parser.parse_args()

    if not (args.mute or args.unmute or args.list):
        parser.error('Exactly one of --list, --mute, or --unmute must be supplied.')

    return args


def main():
    args = parse_args()
    slack_token = args.slack_token

    my_user_name = get_current_user_name(slack_token)
    print(f'Fetching channels for Slack user {my_user_name}...\n')

    my_muted_channel_ids = get_current_muted_channel_ids(slack_token)

    if args.list:
        channel_names = get_channel_names(slack_token, my_muted_channel_ids)

        print(f'You have {len(channel_names)} currently muted channels:')
        for name in channel_names:
            print(name)

    if args.mute:
        shared_channels = get_shared_channels(slack_token)
        shared_channel_ids = get_list_as('id', shared_channels)
        channel_ids_to_mute = shared_channel_ids + my_muted_channel_ids

        print(f'Found {len(my_muted_channel_ids)} currently muted channel(s)')
        print(f'Found {len(shared_channels)} shared channel(s)')
        print(f'Muting all {len(channel_ids_to_mute)} channel(s)...')

        mute_channels(slack_token, channel_ids_to_mute)

        print('Success!')

    if args.unmute:
        shared_channels = get_shared_channels(slack_token)
        shared_channel_ids = get_list_as('id', shared_channels)

        my_muted_nonshared_channel_ids = [id for id in my_muted_channel_ids if id not in shared_channel_ids]

        print(f'Found {len(my_muted_channel_ids)} currently muted channel(s)')
        print(f'Found {len(shared_channels)} shared channel(s)')
        print(f'Unmuting all shared channels so that only {len(my_muted_nonshared_channel_ids)} channel(s) will now be muted...')

        mute_channels(slack_token, my_muted_nonshared_channel_ids)

        print('Success!')



if __name__ == "__main__":
    main()

import requests
import logging
import argparse
import os

logging.basicConfig()
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def get_all_public_channels(slack_token):
    """
    Get a list of all public channels in the Slack account based on https://api.slack.com/methods/channels.list

    This function makes multiple calls to Slack API to account for paginated results.

    Args:
        slack_token: Authentication token bearing required scopes

    Returns:
        A list of channel dictionaries
    """

    url = "https://slack.com/api/channels.list"

    params = {
        'token': slack_token,
        'exclude_archived': True,
        'exclude_members': True
    }

    next_cursor = 'first-pass'
    channels = []

    while next_cursor:
        response = requests.get(url=url, params=params).json()

        if not response['ok']:
            raise Exception(f'List of public channels could not be returned. Response from Slack: {response}')

        for channel in response['channels']:
            if not channel['is_private']:
                channels.append(channel)

        # For pagination to work properly, we must make the next request with the cursor returned from the previous
        # request, if applicable.
        next_cursor = ''

        response_metadata = response.get('response_metadata')

        if response_metadata:
            next_cursor = response_metadata.get('next_cursor')

            if next_cursor:
                params['cursor'] = next_cursor

    return channels


def invite_user(slack_token, email, channel_ids):
    """
    Invite a user using the undocumented Slack API method per https://stackoverflow.com/a/36114710/2308858.

    Args:
        slack_token: Authentication token bearing required scopes
        email: The email of the user
        channel_ids: The list of channels IDs to which the user should be invited
    """

    url = "https://slack.com/api/users.admin.invite"
    channel_ids_str = ",".join(channel_ids)

    params = {
        'token': slack_token,
        'email': email,
        'channels': channel_ids_str,
    }

    response = requests.post(url=url, params=params).json()

    if not response['ok']:
        raise Exception(f'User {email} could not be invited to Slack. Response from Slack: {response}')


def get_as_list_of(property_name, list):
    """
    Given a list of dictionaries, return a list of just the given property name of each item in the list

    Example:
        Given a list [{ foo: bar }, { foo: baz }], get_as_list_of(foo) will return [bar, baz]
    """

    return [item[property_name] for item in list]


def parse_args():
    """Parse command line args"""

    parser = argparse.ArgumentParser(
        description='This script invites a user to the workspace and adds them to all public Slack channels.')

    parser.add_argument('-e', '--email', required=True, help='The email of the user you wish to add')
    args = parser.parse_args()

    if not ('SLACK_TOKEN' in os.environ):
        parser.error('The env var SLACK_TOKEN must be non-empty.')

    return args


def main():
    args = parse_args()
    slack_token = os.environ['SLACK_TOKEN']

    channels = get_all_public_channels(slack_token)

    print('Adding user to channels: ', end='')
    for channel in channels:
        print('#{} '.format(channel['name']), end='')
    print('')

    channel_ids = get_as_list_of('id', channels)
    invite_user(slack_token, args.email, channel_ids)

    print(f'The user has been invited!')


if __name__ == "__main__":
    main()

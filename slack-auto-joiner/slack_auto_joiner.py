# importing the requests library
import re
import requests
import logging
import argparse

logging.basicConfig()
logger = logging.getLogger()
logger.setLevel(logging.INFO)


def get_channels(slack_token):
    conversations_url = "https://slack.com/api/conversations.list"

    first_pass = True
    next_cursor = ""

    resultarray = []

    while next_cursor != "" or first_pass:
        first_pass = False

        params = {'token': slack_token}

        if next_cursor != "":
            params.update({'cursor': next_cursor})

        response = requests.get(url=conversations_url, params=params).json()

        # Grab the next page toke if present in response.
        response_metadata = response['response_metadata']
        next_cursor = response_metadata['next_cursor']

        count = 0
        for channel in response['channels']:
            is_shared = (bool(channel['is_shared'])
                         | bool(channel['is_ext_shared'])
                         | bool(re.match("^_[A-Za-z]+", channel['name']))) \
                        & (not bool(channel['is_archived']))

            if is_shared:
                logger.info(f"--->SHARED Channel Info: {channel['name']} - {channel['id']}")
                resultarray.append(channel)

            count += 1

    logger.info(f"There are: {len(resultarray)} shared channels")

    return resultarray


def invite_user_to_channel(slack_token, user, channel, dry_run):
    if is_user_in_channel(slack_token, user, channel):
        logger.info(f"User {user['profile']['display_name']} is already in channel {channel['name']}.")
    else:
        url = "https://slack.com/api/conversations.invite"
        user_id = user['id']
        display_name = user['profile']['display_name']
        data = {'token': slack_token,
                'users': user_id,
                'channel': channel['id']}

        if not dry_run:
            logger.info(f"Inviting {display_name} ({user_id}) to {channel['name']} with ID: {channel['id']}")
            r = requests.post(url=url, data=data)
            logger.info(r.text)
        else:
            logger.info(f"Would have invited: {display_name} ({user_id}) to {channel['name']} with ID: {channel['id']}")


def is_user_in_channel(slack_token, user, channel):
    url = "https://slack.com/api/conversations.members"

    params = {
        'token': slack_token,
        'channel': channel['id']
    }

    response = requests.get(url=url, params=params).json()

    if response['ok']:
        users_in_channel = response['members']
        return user['id'] in users_in_channel

    return False


def get_all_current_grunt_users(slack_token):
    url = "https://slack.com/api/users.list"
    # gruntwork_team_id="T0PJEPZ2L"

    params = {'token': slack_token}
    raw_response = requests.get(url=url, params=params)

    response = raw_response.json()
    all_users = []
    if response['ok']:
        for member in response['members']:
            if not bool(member['deleted']) and not bool(member['is_bot']) and not bool(member['is_restricted']):
                # logger.info(f"Member: {member['name']}")
                all_users.append(member)

    return all_users


"""Parse the arguments passed to this script
"""


def parse_args():
    parser = argparse.ArgumentParser(
        description='This script can add specified users to all of Gruntworks shared slack channels.')

    parser.add_argument('--slack-token', required=True, help='The Slack API token to use')
    parser.add_argument('-u', '--users', required=True, nargs='+', help='A space separated list of users display names')
    parser.add_argument('-d', '--dry-run', required=False, dest='dry_run', action='store_true', default=False,
                        help='The presence of this flag will not actually invite the users but will print the log '
                             'messages.')

    return parser.parse_args()


def main():
    args = parse_args()

    slack_token = args.slack_token
    all_grunts = get_all_current_grunt_users(slack_token)

    grunts_need_adding = args.users

    grunts_to_add = [g for g in all_grunts if grunt['profile']['display_name'] in grunts_need_adding]

    list(map(lambda grunt: logger.info(f"Going to invite {grunt['profile']['display_name']} (ID: {grunt['id']}) to all "
                                       f"shared channels."), grunts_to_add))

    all_shared_channels = get_channels(slack_token)

    for cur_channel in all_shared_channels:
        for grunt in grunts_to_add:
            invite_user_to_channel(slack_token, grunt, cur_channel, args.dry_run)


if __name__ == "__main__":
    main()

# Mute each channel in the list of channels
def mute_channels(slack_token, currently_muted_channels, new_channels):
    url = "https://slack.com/api/users.prefs.set"

    all_channels = new_channels + currently_muted_channels
    all_channels_str = ','.join(all_channels)

    params = {
        'token': slack_token,
        'prefs': json.dumps({'muted_channels': all_channels_str})
    }

    response = requests.post(url=url, params=params).json()

    if not response['ok']:
        raise Exception(f'Channel could not be muted. Response from Slack: {response}')
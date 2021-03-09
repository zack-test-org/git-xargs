# `add_remove_github_user_to_team.py` is a script that can be executed as a [Zapier code
# step](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to add or remove a user to or from a
# Gruntwork customer's GitHub team. To be able to version and code review this script, it lives in the
# [Gruntwork prototypes repo](https://github.com/gruntwork-io/prototypes). Every time you want to update the Zap,
# update the script in this repo first, submit a PR, and when approved, manually copy/paste the updated code into
# Zapier.

import requests
import logging
import os
import re


# Regex to match "next" URLs in the GitHub API Link header. Example: <https://api.github.com/foo/bar?page=2>; rel="next"
github_api_next_regex = re.compile('<(.+?)>; rel="next"')


def read_from_env(key, required=True):
    """
    Read the given key from the environment. This method first checks the input_data global, which is provided by the
    Zapier code step (https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps). If it's not in input_data,
    the method then looks for an environment variable. If that isn't set either, and required is set to True, this
    function raises an exception.
    :param key: The key to lookup
    :param required: If set to True and no value is found for the key, raise an Exception
    :return: The value for the key
    """
    value = read_input_data(key)
    if value:
        return value

    value = os.environ.get(key)
    if value:
        return value

    if required:
        raise Exception('Did not find value for key {} in either input_data or environment variables.'.format(key))
    else:
        return None


def read_input_data(key):
    """
    Reads the given key from the Zapier code step input_data. Note that input_data is magically added by the Zapier code
    step, so this is the only way I could find to look it up that works both when running locally (with no input_data)
    and in Zapier. https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps
    :param key: The key to read from input_data
    :return: The value of the key or None if it's not set
    """
    try:
        return input_data.get(key)
    except:
        return None


def find_github_team(name, github_creds, teams_api_url='', github_org='gruntwork-io'):
    """
    Find a GitHub team with the given name. Returns the team or None. Note that we have to use the teams list API
    (https://developer.github.com/v3/teams/#list-teams) to find the team because the API call to fetch a team by name
    (https://developer.github.com/v3/teams/#get-team-by-name) requires a "slug", which is not the original team name
    you used, but a version of that name with special characters modified in a variety of ways (e.g., whitespace
    replaced with dashes, dots replaced with dashes, ampersands and surrounding whitespace dropped, etc.) that don't
    seem to be published publicly.
    :param name: The name of the GitHub team
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param teams_api_url: The GitHub API URL to use. Used solely by this function to make recursive API calls
      paginating.
    :param github_org: The GitHub org in which to search.
    :return: The ID of the team or None
    """
    if teams_api_url == '':
        teams_api_url = 'https://api.github.com/orgs/' + github_org + '/teams?per_page=100'

    logging.info('Looking up team {} in GitHub API via URL {}...'.format(name, teams_api_url))
    response = requests.get(teams_api_url, auth=github_creds)
    if response.status_code != 200:
        raise Exception('Got response {} from GitHub when searching for teams at URL {}'.format(response.status_code, teams_api_url))

    teams = response.json()
    for team in teams:
        if team['name'] == name:
            logging.info('Found ID {} for GitHub team {}'.format(team['id'], name))
            return team['id']

    next_url = response.links.get('next')
    if next_url:
        return find_github_team(name, github_creds, next_url['url'])

    logging.info('No team with name {} found'.format(name))
    return None


def get_user_team_membership(github_id, team_id, github_creds):
    """
    Fetch information about the membership of the user with the given GitHub ID on the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#get-team-membership
    :param github_id: The user's GitHub ID
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The JSON response from the GitHub get-team-membership API.
    """
    logging.info('Looking up membership info of GitHub user {} on team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)
    response = requests.get(url, auth=github_creds)

    if response.status_code == 200:
        logging.info('User {} is a member of team {}'.format(github_id, team_id))
        return response.json()
    elif response.status_code == 404:
        logging.info('User {} is not a member of team {}'.format(github_id, team_id))
        return None
    else:
        raise Exception('Failed to look up membership for user {} in team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def add_user_to_team(github_id, team_id, github_creds):
    """
    Add the user with the given GitHub ID to the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#add-or-update-team-membership
    :param github_id: The user's GitHub ID
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The JSON response from the GitHub add-or-update-team-membership API
    """
    logging.info('Adding GitHub user {} to team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)

    payload = {
        'role': 'member'
    }

    response = requests.put(url, auth=github_creds, json=payload)

    if response.status_code == 200:
        logging.info('Successfully added user {} to team {}'.format(github_id, team_id))
        return response.json()
    else:
        raise Exception('Failed to add user {} to team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def remove_user_from_team(github_id, team_id, github_creds):
    """
    Remove the user with the given GitHub ID from the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#remove-team-membership
    :param github_id: The user's GitHub ID
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    logging.info('Deleting GitHub user {} from team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)
    response = requests.delete(url, auth=github_creds)

    if response.status_code == 204:
        logging.info('Successfully deleted user {} from team {}'.format(github_id, team_id))
        return {}
    elif response.status_code == 404:
        logging.info('Skipped deleting user {} not member of team {}'.format(github_id, team_id))
        return {}
    else:
        raise Exception('Failed to delete user {} from team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def remove_user_from_org(github_id, github_creds, github_org='gruntwork-io'):
    """
    Remove the user with the given GitHub ID from the Gruntwork org.
    https://docs.github.com/en/free-pro-team@latest/rest/reference/orgs#remove-an-organization-member
    :param github_id: The user's GitHub ID
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    logging.info('Deleting GitHub user {} from {} org'.format(github_id, github_org))

    url = 'https://api.github.com/orgs/{}/members/{}'.format(github_org, github_id)
    response = requests.delete(url, auth=github_creds)

    if response.status_code == 204:
        logging.info('Successfully deleted user {} from {} org'.format(github_id, github_org))
        return {}
    elif response.status_code == 404:
        logging.info('Skipped deleting user {} not member of {} org'.format(github_id, github_org))
        return {}
    else:
        raise Exception('Failed to delete user {} from {} org. Got response {} from GitHub with body: {}.'.format(github_id, github_org, response.status_code, response.json()))


def format_github_team_name(name):
    """
    Convert the given name to a GitHub team name for a customer. We do this by converting the name to a lower case,
    dash-separated string with a "client-" prefix. E.g., "Foo Bar.com" becomes "client-foo-bar-com".
    :param name: The name to dasherize
    :return: The GitHub-friendly team version of name.
    """
    return 'client-{}'.format(re.sub(r'[\s.]', '-', name).lower())


def do_add_user_to_team(github_id, company_name, github_creds, github_org='gruntwork-io'):
    """
    Add the user with the given GitHub ID to the GitHub team for the company with the given name.
    :param github_id: The user's GitHub ID
    :param company_name: The name of the company. Will be dasherized and looked up in the GitHub API.
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: If the user is already a member, an empty object. Otherwise, the return value of add_user_to_team.
    """
    team_name = format_github_team_name(company_name)
    team_id = find_github_team(team_name, github_creds, github_org=github_org)

    if not team_id:
        raise Exception('Did not find GitHub team called {}'.format(team_name))

    membership = get_user_team_membership(github_id, team_id, github_creds)
    if membership:
        logging.info('User {} is already a member of team {}. Will not add again.'.format(github_id, team_name))
        return {}

    return add_user_to_team(github_id, team_id, github_creds)


def do_remove_user_from_team(github_id, company_name, github_creds, github_org='gruntwork-io'):
    """
    Remove the user with the given GitHub ID from the GitHub team for the company with the given name.
    :param github_id: The user's GitHub ID
    :param company_name: The name of the company. Will be dasherized and looked up in the GitHub API.
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The return value of remove_user_from_team.
    """
    team_name = format_github_team_name(company_name)
    team_id = find_github_team(team_name, github_creds, github_org=github_org)

    if not team_id:
        raise Exception('Did not find GitHub team called {}'.format(team_name))

    membership = get_user_team_membership(github_id, team_id, github_creds)
    if not membership:
        logging.info('Skipped removing user {} because they are not a member of team {}'.format(github_id, team_id))
        return {}

    return remove_user_from_team(github_id, team_id, github_creds)


def is_affirmative_value(value):
    """
    Returns true if the given value represents an "affirmative" from the user (i.e., a "Yes")
    :param value:
    :return:
    """
    return value.lower() in ["yes", "true", "1"]


def is_negative_value(value):
    """
    Returns true if the given value represents an "negative" from the user (i.e., a "No")
    :param value:
    :return:
    """
    return value.lower() in ["no", "false", "0"]


def run():
    """
    Main entrypoint for the code. Reads data from the environment and adds or removes the specified user to/from the
    specified team. Returns the response body of the GitHub add user or delete user GitHub API call.
    """
    github_user = read_from_env('GITHUB_USER')
    github_pass = read_from_env('GITHUB_TOKEN')
    github_clients_user = read_from_env('GITHUB_CLIENTS_USER')
    github_clients_pass = read_from_env('GITHUB_CLIENTS_TOKEN')

    assert len(github_user) > 2, 'GitHub username does not seem to be valid (less than 3 characters long)'
    assert len(github_pass) > 2, 'GitHub password does not seem to be valid (less than 3 characters long)'
    assert len(github_clients_user) > 2, 'GitHub clients username does not seem to be valid (less than 3 characters long)'
    assert len(github_clients_pass) > 2, 'GitHub clients password does not seem to be valid (less than 3 characters long)'

    github_creds = (github_user, github_pass)
    github_clients_creds = (github_clients_user, github_clients_pass)

    github_id = read_from_env('github_id')
    user_active = read_from_env('user_active', required=False)

    assert len(github_id) > 1, 'GitHub ID does not seem to be valid (less than 2 characters long)'

    company_name = read_from_env('company_name')
    company_current_users = int(read_from_env('company_current_users'))
    company_max_users = int(read_from_env('company_max_users'))
    company_active = read_from_env('company_active')
    company_ref_arch = read_from_env('company_ref_arch')

    assert len(company_name) > 1, 'First name does not seem to be valid (less than 2 characters long)'
    assert len(user_active) > 1, 'User active field does not seem to be valid (less than 2 characters long)'
    assert len(company_active) > 1, 'Company active field does not seem to be valid (less than 2 characters long)'
    assert company_current_users >= 0, 'Company current users should not be negative'
    assert company_max_users >= 0, 'Company max users should not be negative'
    assert company_current_users <= company_max_users, 'Company current user must not exceed company max users'

    if not is_affirmative_value(company_active):
        raise Exception('Company {} is not active, cannot add or remove users!'.format(company_name))

    if is_affirmative_value(user_active):
        logging.info('The "active" input for the user is set to "Yes", so adding user to team.')
        if is_affirmative_value(company_ref_arch):
            do_add_user_to_team(github_id, company_name, github_clients_creds, github_org='gruntwork-clients')
        return do_add_user_to_team(github_id, company_name, github_creds)
    elif is_negative_value(user_active):
        logging.info('The "active" input for the user is set to "No", so removing user from the team and org.')
        if is_affirmative_value(company_ref_arch):
            do_remove_user_from_team(github_id, company_name, github_creds, github_org='gruntwork-clients')
            remove_user_from_org(github_id, github_creds, github_org='gruntwork-clients')
        remove_user_from_team_output = do_remove_user_from_team(github_id, company_name, github_creds)
        remove_user_from_org_output = remove_user_from_org(github_id, github_creds)
        # Merge the outputs from the two function calls in a python version compatible way
        remove_user_from_team_output.update(remove_user_from_org_output)
        return remove_user_from_team_output
    else:
        logging.info('The "active" input for the user is not set to "Yes" or "No", so assuming this entry is still a WIP and will not take any action.')
        return {}


# Zapier requires that you set a variable called output with your returned data
logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)
output = run()
logging.info(output)
